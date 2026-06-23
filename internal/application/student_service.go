package application

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"evinizinkocu-backend/internal/domain"
	"evinizinkocu-backend/internal/infrastructure/mailer"
)

type StudentService struct {
	studentRepo domain.StudentRepository
	userRepo    domain.UserRepository
	coachRepo   domain.CoachRepository
	mailer      mailer.Mailer
}

func NewStudentService(
	studentRepo domain.StudentRepository,
	userRepo domain.UserRepository,
	coachRepo domain.CoachRepository,
	mailer mailer.Mailer,
) *StudentService {
	return &StudentService{
		studentRepo: studentRepo,
		userRepo:    userRepo,
		coachRepo:   coachRepo,
		mailer:      mailer,
	}
}

func (s *StudentService) CreateStudent(
	ctx context.Context,
	firstName, lastName, phone, email, classLevel, studyTrack, examTypeID, coachID string,
) (*domain.Student, error) {
	// 1. Verify Coach exists and capacity
	coach, err := s.coachRepo.GetCoachByID(ctx, coachID)
	if err != nil {
		return nil, domain.ErrCoachNotFound
	}

	currentCount, err := s.studentRepo.GetCoachStudentCount(ctx, coachID)
	if err != nil {
		return nil, err
	}

	if currentCount >= coach.StudentCapacity {
		return nil, domain.ErrCapacityExceeded
	}

	// 2. Generate unique username
	baseUsername := strings.ToLower(strings.ReplaceAll(firstName+lastName, " ", ""))
	baseUsername = s.turkishToEnglish(baseUsername)
	username := baseUsername
	suffix := 1

	for {
		_, err := s.userRepo.GetByUsername(ctx, username)
		if errors.Is(err, domain.ErrUserNotFound) {
			break
		}
		if err != nil {
			return nil, err
		}
		username = fmt.Sprintf("%s%d", baseUsername, suffix)
		suffix++
	}

	// 3. Generate temp password
	tempPassword, err := s.generateTempPassword()
	if err != nil {
		return nil, err
	}

	hashedPassword, err := domain.HashPassword(tempPassword)
	if err != nil {
		return nil, err
	}

	// 4. Create User Entity
	user := &domain.User{
		Email:              email,
		Username:           username,
		Phone:              phone,
		PasswordHash:       hashedPassword,
		FirstName:          firstName,
		LastName:           lastName,
		Role:               domain.RoleStudent,
		IsActive:           true,
		MustChangePassword: true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 5. Create Student Entity
	student := &domain.Student{
		ID:          user.ID,
		CoachID:     coachID,
		ClassLevel:  classLevel,
		StudyTrack:  studyTrack,
		ExamTypeID:  examTypeID,
		IsArchived:  false,
	}

	if err := s.studentRepo.CreateStudent(ctx, student); err != nil {
		return nil, err
	}

	student.User = user

	// 6. Send Welcome Email asynchronously
	go func() {
		_ = s.mailer.SendWelcomeCredentials(user.Email, user.FirstName+" "+user.LastName, "Öğrenci", user.Username, tempPassword)
	}()

	return student, nil
}

func (s *StudentService) GetStudentByID(ctx context.Context, id string) (*domain.Student, error) {
	return s.studentRepo.GetStudentByID(ctx, id)
}

func (s *StudentService) UpdateStudentProfile(
	ctx context.Context,
	studentID string,
	firstName, lastName, phone, classLevel, studyTrack, examTypeID string,
	isArchived bool,
) error {
	student, err := s.studentRepo.GetStudentByID(ctx, studentID)
	if err != nil {
		return err
	}

	u := student.User
	u.FirstName = firstName
	u.LastName = lastName
	u.Phone = phone

	if err := s.userRepo.Update(ctx, u); err != nil {
		return err
	}

	student.ClassLevel = classLevel
	student.StudyTrack = studyTrack
	student.ExamTypeID = examTypeID
	student.IsArchived = isArchived

	return s.studentRepo.UpdateStudent(ctx, student)
}

func (s *StudentService) ListStudentsByCoach(ctx context.Context, coachID string, includeArchived bool) ([]*domain.Student, error) {
	return s.studentRepo.ListStudentsByCoach(ctx, coachID, includeArchived)
}

func (s *StudentService) ListAllStudents(ctx context.Context) ([]*domain.Student, error) {
	return s.studentRepo.ListAllStudents(ctx)
}

func (s *StudentService) TransferStudent(ctx context.Context, studentID, newCoachID string) error {
	return s.studentRepo.TransferStudent(ctx, studentID, newCoachID)
}

// Helpers
func (s *StudentService) turkishToEnglish(str string) string {
	r := strings.NewReplacer(
		"ı", "i", "ğ", "g", "ü", "u", "ş", "s", "ö", "o", "ç", "c",
		"İ", "i", "Ğ", "g", "Ü", "u", "Ş", "s", "Ö", "o", "Ç", "c",
	)
	return r.Replace(str)
}

func (s *StudentService) ResendCredentials(ctx context.Context, studentID, requestingUserID, requestingRole string) error {
	student, err := s.studentRepo.GetStudentByID(ctx, studentID)
	if err != nil {
		return err
	}

	// Access control: coach can only resend to their own students
	if requestingRole == domain.RoleCoach && student.CoachID != requestingUserID {
		return errors.New("bu ogrenciye ait bilgileri sifirlama yetkiniz yok")
	}

	tempPassword, err := s.generateTempPassword()
	if err != nil {
		return err
	}

	hashedPassword, err := domain.HashPassword(tempPassword)
	if err != nil {
		return err
	}

	u := student.User
	u.PasswordHash = hashedPassword
	u.MustChangePassword = true

	if err := s.userRepo.Update(ctx, u); err != nil {
		return err
	}

	go func() {
		_ = s.mailer.SendWelcomeCredentials(u.Email, u.FirstName+" "+u.LastName, "Öğrenci", u.Username, tempPassword)
	}()

	return nil
}

func (s *StudentService) generateTempPassword() (string, error) {
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
