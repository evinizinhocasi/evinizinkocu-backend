package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                   string
	DBDSN                  string
	JWTSecret              string
	RefreshJWTSecret       string
	SMTPHost               string
	SMTPPort               int
	SMTPUser               string
	SMTPPass               string
	SMTPFrom               string
	MailTo                 string
	FirebaseCredentialJSON string
}

func LoadConfig() *Config {
	// Optional load of .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or read error, relying on system environment variables")
	}

	smtpPortStr := getEnv("SMTP_PORT", "587")
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		log.Printf("Invalid SMTP_PORT: %s. Using default 587\n", smtpPortStr)
		smtpPort = 587
	}

	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		dbHost := getEnv("DB_HOST", "127.0.0.1")
		dbPort := getEnv("DB_PORT", "5450")
		dbUser := getEnv("DB_USER", "evinizinhocasi")
		dbPass := getEnv("DB_PASSWORD", "!evinizinhocasi34")
		dbName := getEnv("DB_NAME", "evinizinhocasi")
		dbDSN = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)
	}

	smtpUser := getEnv("SMTP_USER", "")

	return &Config{
		Port:                   getEnv("PORT", "8080"),
		DBDSN:                  dbDSN,
		JWTSecret:              getEnv("JWT_SECRET", "local_jwt_secret_development_key"),
		RefreshJWTSecret:       getEnv("REFRESH_JWT_SECRET", "local_refresh_jwt_secret_development_key"),
		SMTPHost:               getEnv("SMTP_HOST", "localhost"),
		SMTPPort:               smtpPort,
		SMTPUser:               smtpUser,
		SMTPPass:               getEnv("SMTP_PASS", ""),
		SMTPFrom:               getEnv("SMTP_FROM", smtpUser),
		MailTo:                 getEnv("MAIL_TO", "info@evinizinhocasi.com"),
		FirebaseCredentialJSON: getEnv("FIREBASE_CREDENTIALS_JSON", ""),
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
