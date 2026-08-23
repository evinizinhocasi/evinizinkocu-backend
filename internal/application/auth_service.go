package application

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"evinizinkocu-backend/internal/domain"
	"evinizinkocu-backend/internal/infrastructure/jwt"
	"evinizinkocu-backend/internal/infrastructure/mailer"

	"github.com/google/uuid"
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
	coachRepo        domain.CoachRepository
	mailer           mailer.Mailer
	jwtSecret        string
	refreshSecret    string
	accessDuration   time.Duration
	refreshDuration  time.Duration
}

func NewAuthService(
	repo domain.UserRepository,
	coachRepo domain.CoachRepository,
	m mailer.Mailer,
	jwtSecret, refreshSecret string,
) *AuthService {
	return &AuthService{
		repo:            repo,
		coachRepo:       coachRepo,
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

func (s *AuthService) RegisterCoach(
	ctx context.Context,
	firstName, lastName, phone, email, password, city, specialization string,
) (*LoginResponse, error) {
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	phone = strings.TrimSpace(phone)
	email = strings.ToLower(strings.TrimSpace(email))
	city = strings.TrimSpace(city)
	specialization = strings.TrimSpace(specialization)

	if firstName == "" || lastName == "" || email == "" || password == "" || phone == "" {
		return nil, errors.New("tüm zorunlu alanları doldurunuz")
	}

	// 1. Check uniqueness
	if _, err := s.repo.GetByEmail(ctx, email); err == nil {
		return nil, domain.ErrEmailConflict
	}
	if _, err := s.repo.GetByPhone(ctx, phone); err == nil {
		return nil, domain.ErrPhoneConflict
	}

	// 2. Generate unique username
	baseUsername := strings.ToLower(strings.ReplaceAll(firstName+lastName, " ", ""))
	baseUsername = s.turkishToEnglish(baseUsername)
	if baseUsername == "" {
		baseUsername = "koc"
	}
	username := baseUsername
	suffix := 1

	for {
		_, err := s.repo.GetByUsername(ctx, username)
		if errors.Is(err, domain.ErrUserNotFound) {
			break
		}
		if err != nil {
			return nil, err
		}
		username = fmt.Sprintf("%s%d", baseUsername, suffix)
		suffix++
	}

	// 3. Hash password
	hashedPassword, err := domain.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// 4. Create User
	userID := uuid.New().String()
	user := &domain.User{
		ID:                 userID,
		Email:              email,
		Username:           username,
		Phone:              phone,
		PasswordHash:       hashedPassword,
		FirstName:          firstName,
		LastName:           lastName,
		Role:               domain.RoleCoach,
		IsActive:           true,
		MustChangePassword: false,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 5. Create Coach with StudentCapacity = 1 (FREE 1-Student Capacity)
	coach := &domain.Coach{
		ID:                      userID,
		City:                    city,
		Biography:               "",
		Specialization:          specialization,
		SocialLinks:             "{}",
		StudentCapacity:         1, // Free Tier: 1 Student
		AuthStartDate:           time.Now(),
		AuthEndDate:             time.Now().AddDate(50, 0, 0),
		PermissionImmediatePush: true,
		PermissionScheduledPush: true,
	}

	if err := s.coachRepo.CreateCoach(ctx, coach); err != nil {
		return nil, err
	}

	// 6. Generate Tokens for immediate login
	accessToken, err := jwt.GenerateAccessToken(user.ID, user.Role, user.Email, s.jwtSecret, s.accessDuration)
	if err != nil {
		return nil, fmt.Errorf("failed generating access token: %w", err)
	}

	rawRefreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(rawRefreshTokenBytes); err != nil {
		return nil, err
	}
	rawRefreshToken := fmt.Sprintf("%x", rawRefreshTokenBytes)

	rt := &domain.RefreshToken{
		UserID:    user.ID,
		Token:     rawRefreshToken,
		ExpiresAt: time.Now().Add(s.refreshDuration),
	}

	if err := s.repo.SaveRefreshToken(ctx, rt); err != nil {
		return nil, fmt.Errorf("failed saving refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		User:         user,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, token string) (*TokenResponse, error) {
	rt, err := s.repo.GetRefreshToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if rt.IsRevoked || time.Now().After(rt.ExpiresAt) {
		return nil, errors.New("refresh token expired or revoked")
	}

	u, err := s.repo.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}

	if !u.IsActive {
		return nil, domain.ErrPassiveAccount
	}

	accessToken, err := jwt.GenerateAccessToken(u.ID, u.Role, u.Email, s.jwtSecret, s.accessDuration)
	if err != nil {
		return nil, err
	}

	rawRefreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(rawRefreshTokenBytes); err != nil {
		return nil, err
	}
	newRefreshToken := fmt.Sprintf("%x", rawRefreshTokenBytes)

	_ = s.repo.RevokeRefreshToken(ctx, token)

	newRt := &domain.RefreshToken{
		UserID:    u.ID,
		Token:     newRefreshToken,
		ExpiresAt: time.Now().Add(s.refreshDuration),
	}

	if err := s.repo.SaveRefreshToken(ctx, newRt); err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken, deviceToken, userID string) error {
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
			return nil
		}
		return err
	}

	codeVal, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return err
	}
	code := fmt.Sprintf("%06d", codeVal.Int64()+100000)

	codeHashBytes, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = s.repo.SaveResetCode(ctx, u.ID, string(codeHashBytes), 10*time.Minute)
	if err != nil {
		return err
	}

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

	err = bcrypt.CompareHashAndPassword([]byte(prc.CodeHash), []byte(code))
	if err != nil {
		return errors.New("geçersiz kod")
	}

	hashed, err := domain.HashPassword(newPassword)
	if err != nil {
		return err
	}

	u.PasswordHash = hashed
	u.MustChangePassword = false
	if err := s.repo.Update(ctx, u); err != nil {
		return err
	}

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

func (s *AuthService) turkishToEnglish(str string) string {
	r := strings.NewReplacer(
		"ı", "i", "ğ", "g", "ü", "u", "ş", "s", "ö", "o", "ç", "c",
		"İ", "i", "Ğ", "g", "Ü", "u", "Ş", "s", "Ö", "o", "Ç", "c",
	)
	return r.Replace(str)
}
