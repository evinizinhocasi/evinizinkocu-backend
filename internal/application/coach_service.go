package application

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"evinizinkocu-backend/internal/domain"
	"evinizinkocu-backend/internal/infrastructure/mailer"
)

type CoachService struct {
	coachRepo  domain.CoachRepository
	userRepo   domain.UserRepository
	mailer     mailer.Mailer
	adminEmail string
}

func NewCoachService(
	coachRepo domain.CoachRepository,
	userRepo domain.UserRepository,
	mailer mailer.Mailer,
	adminEmail string,
) *CoachService {
	return &CoachService{
		coachRepo:  coachRepo,
		userRepo:   userRepo,
		mailer:     mailer,
		adminEmail: adminEmail,
	}
}

func (s *CoachService) CreateApplication(ctx context.Context, first, last, phone, email, city, specialization, explanation string) error {
	app := &domain.CoachApplication{
		FirstName:      first,
		LastName:       last,
		Phone:          phone,
		Email:          email,
		City:           city,
		Specialization: specialization,
		Explanation:    explanation,
		Status:         "pending",
	}

	err := s.coachRepo.CreateApplication(ctx, app)
	if err != nil {
		return err
	}

	// Alert superadmin
	go func() {
		_ = s.mailer.SendCoachApplicationAlert(s.adminEmail, first+" "+last, email)
	}()

	return nil
}

func (s *CoachService) CreateCoach(
	ctx context.Context,
	firstName, lastName, phone, email, city, specialization string,
	capacity int,
	start, end time.Time,
	immediatePush, scheduledPush bool,
) error {
	// 1. Generate unique username
	baseUsername := strings.ToLower(strings.ReplaceAll(firstName+lastName, " ", ""))
	// Strip special Turkish characters
	baseUsername = s.turkishToEnglish(baseUsername)
	username := baseUsername
	suffix := 1

	for {
		_, err := s.userRepo.GetByUsername(ctx, username)
		if errors.Is(err, domain.ErrUserNotFound) {
			break
		}
		if err != nil {
			return err
		}
		username = fmt.Sprintf("%s%d", baseUsername, suffix)
		suffix++
	}

	// 2. Generate secure temp password
	tempPassword, err := s.generateTempPassword()
	if err != nil {
		return err
	}
	hashedPassword, err := domain.HashPassword(tempPassword)
	if err != nil {
		return err
	}

	// Create Coach User
	user := &domain.User{
		Email:              email,
		Username:           username,
		Phone:              phone,
		PasswordHash:       hashedPassword,
		FirstName:          firstName,
		LastName:           lastName,
		Role:               domain.RoleCoach,
		IsActive:           true,
		MustChangePassword: true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return err
	}

	// Create Coach Details
	coach := &domain.Coach{
		ID:                      user.ID,
		City:                    city,
		Specialization:          specialization,
		StudentCapacity:         capacity,
		AuthStartDate:           start,
		AuthEndDate:             end,
		PermissionImmediatePush: immediatePush,
		PermissionScheduledPush: scheduledPush,
	}

	if err := s.coachRepo.CreateCoach(ctx, coach); err != nil {
		return err
	}

	// Save authorization period history
	ap := &domain.CoachAuthorizationPeriod{
		CoachID:         user.ID,
		StartDate:       start,
		EndDate:         end,
		StudentCapacity: capacity,
	}
	_ = s.coachRepo.SaveAuthorizationPeriod(ctx, ap)

	// Send SMTP
	go func() {
		_ = s.mailer.SendWelcomeCredentials(user.Email, user.FirstName+" "+user.LastName, "Koç", user.Username, tempPassword)
	}()

	return nil
}

func (s *CoachService) ApproveApplication(
	ctx context.Context,
	appID string,
	capacity int,
	start, end time.Time,
	immediatePush, scheduledPush bool,
) error {
	app, err := s.coachRepo.GetApplicationByID(ctx, appID)
	if err != nil {
		return err
	}

	if app.Status != "pending" {
		return errors.New("bu başvuru zaten değerlendirilmiş")
	}

	err = s.CreateCoach(
		ctx,
		app.FirstName,
		app.LastName,
		app.Phone,
		app.Email,
		app.City,
		app.Specialization,
		capacity,
		start,
		end,
		immediatePush,
		scheduledPush,
	)
	if err != nil {
		return err
	}

	// Update status
	if err := s.coachRepo.UpdateApplicationStatus(ctx, appID, "approved"); err != nil {
		return err
	}

	return nil
}

func (s *CoachService) RejectApplication(ctx context.Context, appID string) error {
	app, err := s.coachRepo.GetApplicationByID(ctx, appID)
	if err != nil {
		return err
	}

	if app.Status != "pending" {
		return errors.New("bu başvuru zaten değerlendirilmiş")
	}

	return s.coachRepo.UpdateApplicationStatus(ctx, appID, "rejected")
}

func (s *CoachService) GetCoachProfile(ctx context.Context, id string) (*domain.Coach, error) {
	return s.coachRepo.GetCoachByID(ctx, id)
}

func (s *CoachService) UpdateCoachProfile(
	ctx context.Context,
	id string,
	first, last, phone, city, biography, specialization, socialLinks string,
) error {
	c, err := s.coachRepo.GetCoachByID(ctx, id)
	if err != nil {
		return err
	}

	u := c.User
	u.FirstName = first
	u.LastName = last
	u.Phone = phone

	if err := s.userRepo.Update(ctx, u); err != nil {
		return err
	}

	c.City = city
	c.Biography = biography
	c.Specialization = specialization
	c.SocialLinks = socialLinks

	return s.coachRepo.UpdateCoach(ctx, c)
}

func (s *CoachService) UpdateCoachSettings(
	ctx context.Context,
	coachID string,
	capacity int,
	start, end time.Time,
	immediatePush, scheduledPush bool,
) error {
	c, err := s.coachRepo.GetCoachByID(ctx, coachID)
	if err != nil {
		return err
	}

	c.StudentCapacity = capacity
	c.AuthStartDate = start
	c.AuthEndDate = end
	c.PermissionImmediatePush = immediatePush
	c.PermissionScheduledPush = scheduledPush

	if err := s.coachRepo.UpdateCoach(ctx, c); err != nil {
		return err
	}

	// Log change
	ap := &domain.CoachAuthorizationPeriod{
		CoachID:         coachID,
		StartDate:       start,
		EndDate:         end,
		StudentCapacity: capacity,
	}
	_ = s.coachRepo.SaveAuthorizationPeriod(ctx, ap)

	return nil
}

func (s *CoachService) ListCoaches(ctx context.Context) ([]*domain.Coach, error) {
	return s.coachRepo.ListCoaches(ctx)
}

func (s *CoachService) ListApplications(ctx context.Context) ([]*domain.CoachApplication, error) {
	return s.coachRepo.ListApplications(ctx)
}

func (s *CoachService) CheckAndProcessExpiredCoaches(ctx context.Context) error {
	expiredIds, err := s.coachRepo.GetExpiredCoaches(ctx)
	if err != nil {
		return err
	}

	if len(expiredIds) == 0 {
		return nil
	}

	log.Printf("Deactivating %d expired coaches: %v\n", len(expiredIds), expiredIds)
	err = s.coachRepo.DeactivateCoaches(ctx, expiredIds)
	if err != nil {
		return err
	}

	// Send deactivation notification emails
	for _, id := range expiredIds {
		c, err := s.coachRepo.GetCoachByID(ctx, id)
		if err == nil {
			go func(email, name string) {
				_ = s.mailer.SendDeactivationNotice(email, name)
			}(c.User.Email, c.User.FirstName+" "+c.User.LastName)
		}
	}

	return nil
}

// Helpers
func (s *CoachService) turkishToEnglish(str string) string {
	r := strings.NewReplacer(
		"ı", "i", "ğ", "g", "ü", "u", "ş", "s", "ö", "o", "ç", "c",
		"İ", "i", "Ğ", "g", "Ü", "u", "Ş", "s", "Ö", "o", "Ç", "c",
	)
	return r.Replace(str)
}

func (s *CoachService) ResendCredentials(ctx context.Context, coachID string) error {
	c, err := s.coachRepo.GetCoachByID(ctx, coachID)
	if err != nil {
		return err
	}

	tempPassword, err := s.generateTempPassword()
	if err != nil {
		return err
	}

	hashedPassword, err := domain.HashPassword(tempPassword)
	if err != nil {
		return err
	}

	u := c.User
	u.PasswordHash = hashedPassword
	u.MustChangePassword = true

	if err := s.userRepo.Update(ctx, u); err != nil {
		return err
	}

	go func() {
		_ = s.mailer.SendWelcomeCredentials(u.Email, u.FirstName+" "+u.LastName, "Koç", u.Username, tempPassword)
	}()

	return nil
}

func (s *CoachService) generateTempPassword() (string, error) {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, 12)
	for i := range password {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		password[i] = chars[num.Int64()]
	}
	return string(password), nil
}
