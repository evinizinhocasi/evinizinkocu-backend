package application

import (
	"context"
	"errors"
	"time"

	"evinizinkocu-backend/internal/domain"
)

type PaymentService struct {
	repo        domain.PaymentRepository
	coachRepo   domain.CoachRepository
	studentRepo domain.StudentRepository
}

func NewPaymentService(
	repo domain.PaymentRepository,
	coachRepo domain.CoachRepository,
	studentRepo domain.StudentRepository,
) *PaymentService {
	return &PaymentService{
		repo:        repo,
		coachRepo:   coachRepo,
		studentRepo: studentRepo,
	}
}

// Coach payments (Superadmin managed)

func (s *PaymentService) CreateCoachPayment(
	ctx context.Context,
	coachID string,
	amount float64,
	payDate time.Time,
	desc, status string,
) (*domain.CoachPayment, error) {
	if amount < 0 {
		return nil, errors.New("ödeme tutarı negatif olamaz")
	}

	_, err := s.coachRepo.GetCoachByID(ctx, coachID)
	if err != nil {
		return nil, domain.ErrCoachNotFound
	}

	p := &domain.CoachPayment{
		CoachID:     coachID,
		Amount:      amount,
		PaymentDate: payDate,
		Description: desc,
		Status:      status,
	}

	err = s.repo.CreateCoachPayment(ctx, p)
	return p, err
}

func (s *PaymentService) UpdateCoachPayment(
	ctx context.Context,
	id string,
	amount float64,
	payDate time.Time,
	desc, status string,
) error {
	p, err := s.repo.GetCoachPaymentByID(ctx, id)
	if err != nil {
		return err
	}

	if amount < 0 {
		return errors.New("ödeme tutarı negatif olamaz")
	}

	p.Amount = amount
	p.PaymentDate = payDate
	p.Description = desc
	p.Status = status

	return s.repo.UpdateCoachPayment(ctx, p)
}

func (s *PaymentService) ListCoachPayments(ctx context.Context, coachID string) ([]*domain.CoachPayment, error) {
	return s.repo.ListCoachPayments(ctx, coachID)
}

func (s *PaymentService) DeleteCoachPayment(ctx context.Context, id string) error {
	return s.repo.DeleteCoachPayment(ctx, id)
}

// Student payments (Coach managed)

func (s *PaymentService) CreateStudentPayment(
	ctx context.Context,
	studentID, coachID string,
	amount float64,
	payDate time.Time,
	desc, status string,
) (*domain.StudentPayment, error) {
	if amount < 0 {
		return nil, errors.New("ödeme tutarı negatif olamaz")
	}

	student, err := s.studentRepo.GetStudentByID(ctx, studentID)
	if err != nil {
		return nil, domain.ErrStudentNotFound
	}

	p := &domain.StudentPayment{
		StudentID:   studentID,
		CoachID:     coachID,
		Amount:      amount,
		PaymentDate: payDate,
		Description: desc,
		Status:      status,
	}

	// Verify coach owns student
	if student.CoachID != coachID {
		return nil, errors.New("bu öğrenci size ait değil")
	}

	err = s.repo.CreateStudentPayment(ctx, p)
	return p, err
}

func (s *PaymentService) UpdateStudentPayment(
	ctx context.Context,
	id string,
	amount float64,
	payDate time.Time,
	desc, status string,
) error {
	p, err := s.repo.GetStudentPaymentByID(ctx, id)
	if err != nil {
		return err
	}

	if amount < 0 {
		return errors.New("ödeme tutarı negatif olamaz")
	}

	p.Amount = amount
	p.PaymentDate = payDate
	p.Description = desc
	p.Status = status

	return s.repo.UpdateStudentPayment(ctx, p)
}

func (s *PaymentService) ListStudentPayments(ctx context.Context, studentID string) ([]*domain.StudentPayment, error) {
	return s.repo.ListStudentPayments(ctx, studentID)
}

func (s *PaymentService) ListStudentPaymentsByCoach(ctx context.Context, coachID string) ([]*domain.StudentPayment, error) {
	return s.repo.ListStudentPaymentsByCoach(ctx, coachID)
}

func (s *PaymentService) DeleteStudentPayment(ctx context.Context, id string) error {
	return s.repo.DeleteStudentPayment(ctx, id)
}

func (s *PaymentService) GetStudentPaymentByID(ctx context.Context, id string) (*domain.StudentPayment, error) {
	return s.repo.GetStudentPaymentByID(ctx, id)
}
