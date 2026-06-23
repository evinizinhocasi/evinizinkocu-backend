package domain

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Roles
const (
	RoleSuperadmin = "superadmin"
	RoleCoach      = "coach"
	RoleStudent    = "student"
)

// Standard Domain Errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailConflict      = errors.New("email already in use")
	ErrUsernameConflict   = errors.New("username already in use")
	ErrPhoneConflict      = errors.New("phone number already in use")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrForbidden          = errors.New("forbidden resource")
	ErrPassiveAccount     = errors.New("account is passive")
	ErrMustChangePassword = errors.New("password change required")
)

type User struct {
	ID                 string    `json:"id"`
	Email              string    `json:"email"`
	Username           string    `json:"username"`
	Phone              string    `json:"phone"`
	PasswordHash       string    `json:"-"`
	FirstName          string    `json:"first_name"`
	LastName           string    `json:"last_name"`
	Role               string    `json:"role"`
	IsActive           bool      `json:"is_active"`
	MustChangePassword bool      `json:"must_change_password"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	IsRevoked bool      `json:"is_revoked"`
}

type PasswordResetCode struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	CodeHash  string    `json:"code_hash"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UsedAt    *time.Time `json:"used_at"`
}

type DeviceToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	Platform  string    `json:"platform"`
	CreatedAt time.Time `json:"created_at"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByPhone(ctx context.Context, phone string) (*User, error)
	GetByIdentifier(ctx context.Context, identifier string) (*User, error)
	Update(ctx context.Context, user *User) error

	// Token management
	SaveRefreshToken(ctx context.Context, rt *RefreshToken) error
	GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
	RevokeAllUserRefreshTokens(ctx context.Context, userID string) error

	// Password reset
	SaveResetCode(ctx context.Context, userID, codeHash string, duration time.Duration) error
	GetLatestActiveResetCode(ctx context.Context, userID string) (*PasswordResetCode, error)
	MarkResetCodeUsed(ctx context.Context, codeID string) error

	// Device tokens
	SaveDeviceToken(ctx context.Context, dt *DeviceToken) error
	RemoveDeviceToken(ctx context.Context, userID, token string) error
	GetDeviceTokensByUser(ctx context.Context, userID string) ([]string, error)
}

// Password utility helpers
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
