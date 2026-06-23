package application

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"evinizinkocu-backend/internal/domain"
	"evinizinkocu-backend/internal/infrastructure/jwt"
	"evinizinkocu-backend/internal/infrastructure/mailer"

	"golang.org/x/crypto/bcrypt"
)

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         *domain.User `json:"user"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthService struct {
	repo             domain.UserRepository
	mailer           mailer.Mailer
	jwtSecret        string
	refreshSecret    string
	accessDuration   time.Duration
	refreshDuration  time.Duration
}

func NewAuthService(
	repo domain.UserRepository,
	m mailer.Mailer,
	jwtSecret, refreshSecret string,
) *AuthService {
	return &AuthService{
		repo:            repo,
		mailer:          m,
		jwtSecret:       jwtSecret,
		refreshSecret:   refreshSecret,
		accessDuration:  15 * time.Minute,
		refreshDuration: 7 * 24 * time.Hour,
	}
}

func (s *AuthService) Login(ctx context.Context, identifier, password string) (*LoginResponse, error) {
	u, err := s.repo.GetByIdentifier(ctx, identifier)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if !u.IsActive {
		return nil, domain.ErrPassiveAccount
	}

	if !domain.CheckPasswordHash(password, u.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	// For coach check expiry
	if u.Role == domain.RoleCoach {
		// We will implement coach check later when writing CoachService or within the DB layer.
		// For now we check if they are disabled. If a coach is passive, they cannot log in.
	}

	// Generate Access Token
	accessToken, err := jwt.GenerateAccessToken(u.ID, u.Role, u.Email, s.jwtSecret, s.accessDuration)
	if err != nil {
		return nil, fmt.Errorf("failed generating access token: %w", err)
	}

	// Generate Refresh Token
	rawRefreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(rawRefreshTokenBytes); err != nil {
		return nil, err
	}
	rawRefreshToken := fmt.Sprintf("%x", rawRefreshTokenBytes)

	rt := &domain.RefreshToken{
		UserID:    u.ID,
		Token:     rawRefreshToken,
		ExpiresAt: time.Now().Add(s.refreshDuration),
	}

	if err := s.repo.SaveRefreshToken(ctx, rt); err != nil {
		return nil, fmt.Errorf("failed saving refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		User:         u,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, tokenStr string) (*TokenResponse, error) {
	rt, err := s.repo.GetRefreshToken(ctx, tokenStr)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	if rt.IsRevoked || rt.ExpiresAt.Before(time.Now()) {
		// Enforce rotating refresh token rule: revoke all user refresh tokens on theft detection
		_ = s.repo.RevokeAllUserRefreshTokens(ctx, rt.UserID)
		return nil, domain.ErrUnauthorized
	}

	// Revoke current token
	if err := s.repo.RevokeRefreshToken(ctx, tokenStr); err != nil {
		return nil, err
	}

	// Get User
	u, err := s.repo.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	if !u.IsActive {
		return nil, domain.ErrPassiveAccount
	}

	// Generate Access Token
	newAccessToken, err := jwt.GenerateAccessToken(u.ID, u.Role, u.Email, s.jwtSecret, s.accessDuration)
	if err != nil {
		return nil, err
	}

	// Generate new Refresh Token
	rawRefreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(rawRefreshTokenBytes); err != nil {
		return nil, err
	}
	newRawRefreshToken := fmt.Sprintf("%x", rawRefreshTokenBytes)

	newRt := &domain.RefreshToken{
		UserID:    u.ID,
		Token:     newRawRefreshToken,
		ExpiresAt: time.Now().Add(s.refreshDuration),
	}

	if err := s.repo.SaveRefreshToken(ctx, newRt); err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRawRefreshToken,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken, deviceToken string, userID string) error {
	if refreshToken != "" {
		_ = s.repo.RevokeRefreshToken(ctx, refreshToken)
	}
	if deviceToken != "" && userID != "" {
		_ = s.repo.RemoveDeviceToken(ctx, userID, deviceToken)
	}
	return nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if !domain.CheckPasswordHash(currentPassword, u.PasswordHash) {
		return errors.New("mevcut şifre hatalı")
	}

	hashed, err := domain.HashPassword(newPassword)
	if err != nil {
		return err
	}

	u.PasswordHash = hashed
	u.MustChangePassword = false
	return s.repo.Update(ctx, u)
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			// Do not leak email existence, return nil
			return nil
		}
		return err
	}

	// Generate 6 digit numeric code
	codeVal, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return err
	}
	code := fmt.Sprintf("%06d", codeVal.Int64()+100000)

	// Hash code before saving
	codeHashBytes, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = s.repo.SaveResetCode(ctx, u.ID, string(codeHashBytes), 10*time.Minute)
	if err != nil {
		return err
	}

	// Send SMTP
	go func() {
		_ = s.mailer.SendResetPasswordCode(u.Email, code)
	}()

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return domain.ErrUserNotFound
	}

	prc, err := s.repo.GetLatestActiveResetCode(ctx, u.ID)
	if err != nil {
		return errors.New("geçersiz veya süresi dolmuş kod")
	}

	// Validate code hash
	err = bcrypt.CompareHashAndPassword([]byte(prc.CodeHash), []byte(code))
	if err != nil {
		return errors.New("geçersiz kod")
	}

	// Update password
	hashed, err := domain.HashPassword(newPassword)
	if err != nil {
		return err
	}

	u.PasswordHash = hashed
	u.MustChangePassword = false
	if err := s.repo.Update(ctx, u); err != nil {
		return err
	}

	// Mark used
	return s.repo.MarkResetCodeUsed(ctx, prc.ID)
}

func (s *AuthService) RegisterDeviceToken(ctx context.Context, userID, token, platform string) error {
	dt := &domain.DeviceToken{
		UserID:   userID,
		Token:    token,
		Platform: platform,
	}
	return s.repo.SaveDeviceToken(ctx, dt)
}
