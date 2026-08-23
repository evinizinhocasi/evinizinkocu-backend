package storage

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"evinizinkocu-backend/internal/config"
)

type R2Storage struct {
	accountID string
	accessKey string
	secretKey string
	bucket    string
	publicURL string
}

func NewR2Storage(cfg *config.Config) *R2Storage {
	return &R2Storage{
		accountID: strings.TrimSpace(cfg.R2AccountID),
		accessKey: strings.TrimSpace(cfg.R2AccessKeyID),
		secretKey: strings.TrimSpace(cfg.R2SecretAccessKey),
		bucket:    strings.TrimSpace(cfg.R2BucketName),
		publicURL: strings.TrimRight(strings.TrimSpace(cfg.R2PublicURL), "/"),
	}
}

// UploadFile uploads a file to Cloudflare R2 using standard S3 SigV4.
// If R2 credentials are missing or incomplete, it falls back to saving locally in ./uploads/wrong_questions.
func (s *R2Storage) UploadFile(fileName string, content []byte, contentType string) (string, error) {
	if s.accountID == "" || s.accessKey == "" || s.secretKey == "" || s.bucket == "" {
		uploadDir := filepath.Join(".", "uploads", "wrong_questions")
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			return "", fmt.Errorf("local dir creation error: %v", err)
		}
		localPath := filepath.Join(uploadDir, fileName)
		if err := os.WriteFile(localPath, content, 0644); err != nil {
			return "", fmt.Errorf("local file save error: %v", err)
		}
		return fmt.Sprintf("/uploads/wrong_questions/%s", fileName), nil
	}

	region := "auto"
	service := "s3"
	host := fmt.Sprintf("%s.r2.cloudflarestorage.com", s.accountID)
	endpoint := fmt.Sprintf("https://%s/%s/%s", host, s.bucket, fileName)

	t := time.Now().UTC()
	amzDate := t.Format("20060102T150405Z")
	dateStamp := t.Format("20060102")

	// Payload hash
	payloadHash := sha256Hash(content)

	// Canonical Request
	canonicalURI := fmt.Sprintf("/%s/%s", s.bucket, fileName)
	canonicalQueryString := ""
	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\nx-amz-content-sha256:%s\nx-amz-date:%s\n",
		contentType, host, payloadHash, amzDate)
	signedHeaders := "content-type;host;x-amz-content-sha256;x-amz-date"

	canonicalRequest := fmt.Sprintf("PUT\n%s\n%s\n%s\n%s\n%s",
		canonicalURI, canonicalQueryString, canonicalHeaders, signedHeaders, payloadHash)

	// String to Sign
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, region, service)
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s",
		amzDate, credentialScope, sha256Hash([]byte(canonicalRequest)))

	// Calculate Signature
	signingKey := getSignatureKey(s.secretKey, dateStamp, region, service)
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	// Authorization Header
	authorizationHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.accessKey, credentialScope, signedHeaders, signature)

	// Make HTTP PUT Request
	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewReader(content))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Host", host)
	req.Header.Set("x-amz-content-sha256", payloadHash)
	req.Header.Set("x-amz-date", amzDate)
	req.Header.Set("Authorization", authorizationHeader)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("R2 upload request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("R2 upload failed [%d]: %s", resp.StatusCode, string(bodyBytes))
	}

	// Format public URL
	if s.publicURL != "" && !strings.Contains(s.publicURL, ".r2.dev") {
		return fmt.Sprintf("%s/%s", s.publicURL, fileName), nil
	}

	if s.publicURL != "" {
		return fmt.Sprintf("%s/%s", s.publicURL, fileName), nil
	}

	return fmt.Sprintf("/media/%s", fileName), nil
}

// DownloadFile downloads a file from Cloudflare R2 using standard S3 SigV4.
func (s *R2Storage) DownloadFile(fileName string) ([]byte, string, error) {
	if s.accountID == "" || s.accessKey == "" || s.secretKey == "" || s.bucket == "" {
		return nil, "", fmt.Errorf("R2 credentials not configured")
	}

	region := "auto"
	service := "s3"
	host := fmt.Sprintf("%s.r2.cloudflarestorage.com", s.accountID)
	endpoint := fmt.Sprintf("https://%s/%s/%s", host, s.bucket, fileName)

	t := time.Now().UTC()
	amzDate := t.Format("20060102T150405Z")
	dateStamp := t.Format("20060102")

	payloadHash := sha256Hash([]byte(""))

	canonicalURI := fmt.Sprintf("/%s/%s", s.bucket, fileName)
	canonicalQueryString := ""
	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-content-sha256:%s\nx-amz-date:%s\n",
		host, payloadHash, amzDate)
	signedHeaders := "host;x-amz-content-sha256;x-amz-date"

	canonicalRequest := fmt.Sprintf("GET\n%s\n%s\n%s\n%s\n%s",
		canonicalURI, canonicalQueryString, canonicalHeaders, signedHeaders, payloadHash)

	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, region, service)
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s",
		amzDate, credentialScope, sha256Hash([]byte(canonicalRequest)))

	signingKey := getSignatureKey(s.secretKey, dateStamp, region, service)
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	authorizationHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.accessKey, credentialScope, signedHeaders, signature)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("Host", host)
	req.Header.Set("x-amz-content-sha256", payloadHash)
	req.Header.Set("x-amz-date", amzDate)
	req.Header.Set("Authorization", authorizationHeader)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("R2 fetch request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("R2 fetch status: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return bodyBytes, contentType, nil
}

func sha256Hash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func hmacSHA256(key []byte, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func getSignatureKey(key string, dateStamp string, regionName string, serviceName string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+key), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(regionName))
	kService := hmacSHA256(kRegion, []byte(serviceName))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))
	return kSigning
}
