package fcm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"golang.org/x/oauth2/google"
)

type FCMService interface {
	SendPushNotification(ctx context.Context, token, title, body string) error
	SendPushToTokens(ctx context.Context, tokens []string, title, body string) error
}

type GoogleFCMService struct {
	credentialJSON string
	projectID      string
}

func NewGoogleFCMService(credentialJSON string) FCMService {
	// Parse project ID from credential JSON if available
	var projectID string
	if credentialJSON != "" {
		var cred struct {
			ProjectID string `json:"project_id"`
		}
		if err := json.Unmarshal([]byte(credentialJSON), &cred); err == nil {
			projectID = cred.ProjectID
		}
	}

	return &GoogleFCMService{
		credentialJSON: credentialJSON,
		projectID:      projectID,
	}
}

func (s *GoogleFCMService) SendPushNotification(ctx context.Context, token, title, body string) error {
	if s.credentialJSON == "" || s.projectID == "" {
		log.Printf("\n--- [CONSOLE PUSH NOTIFICATION] ---\nToken: %s\nTitle: %s\nBody: %s\n-----------------------------------\n", token, title, body)
		return nil
	}

	// Dynamic token acquisition for FCM HTTP v1 using oauth2
	credentials, err := google.CredentialsFromJSON(ctx, []byte(s.credentialJSON), "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return fmt.Errorf("failed parsing firebase credentials: %w", err)
	}

	tokenSource := credentials.TokenSource
	oauthToken, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("failed fetching oauth token: %w", err)
	}

	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", s.projectID)

	// Build HTTP v1 request payload
	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"token": token,
			"notification": map[string]string{
				"title": title,
				"body":  body,
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", oauthToken.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("FCM server returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	return nil
}

func (s *GoogleFCMService) SendPushToTokens(ctx context.Context, tokens []string, title, body string) error {
	for _, t := range tokens {
		err := s.SendPushNotification(ctx, t, title, body)
		if err != nil {
			log.Printf("FCM: failed sending to token %s: %v\n", t, err)
		}
	}
	return nil
}
