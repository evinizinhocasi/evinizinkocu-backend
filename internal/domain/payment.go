package domain

import (
	"context"
	"time"
)

type CoachPayment struct {
	ID          string    `json:"id"`
	CoachID     string    `json:"coach_id"`
	Coach       *Coach    `json:"coach,omitempty"`
	Amount      float64   `json:"amount"`
	PaymentDate time.Time `json:"payment_date"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"` // 'paid', 'pending', 'overdue', 'cancelled'
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type StudentPayment struct {
	ID          string    `json:"id"`
	StudentID   string    `json:"student_id"`
	Student     *Student  `json:"student,omitempty"`
	CoachID     string    `json:"coach_id"`
	Amount      float64   `json:"amount"`
	PaymentDate time.Time `json:"payment_date"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"` // 'paid', 'pending', 'overdue', 'cancelled'
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PaymentRepository interface {
	// Coach payments (Superadmin managed)
	CreateCoachPayment(ctx context.Context, p *CoachPayment) error
	GetCoachPaymentByID(ctx context.Context, id string) (*CoachPayment, error)
	UpdateCoachPayment(ctx context.Context, p *CoachPayment) error
	ListCoachPayments(ctx context.Context, coachID string) ([]*CoachPayment, error)
	DeleteCoachPayment(ctx context.Context, id string) error

	// Student payments (Coach managed)
	CreateStudentPayment(ctx context.Context, p *StudentPayment) error
	GetStudentPaymentByID(ctx context.Context, id string) (*StudentPayment, error)
	UpdateStudentPayment(ctx context.Context, p *StudentPayment) error
	ListStudentPayments(ctx context.Context, studentID string) ([]*StudentPayment, error)
	ListStudentPaymentsByCoach(ctx context.Context, coachID string) ([]*StudentPayment, error)
	DeleteStudentPayment(ctx context.Context, id string) error
}
