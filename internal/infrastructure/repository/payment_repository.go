package repository

import (
	"context"
	"errors"

	"evinizinkocu-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresPaymentRepository struct {
	db *pgxpool.Pool
}

func NewPostgresPaymentRepository(db *pgxpool.Pool) domain.PaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

// Coach payments (Superadmin managed)

func (r *PostgresPaymentRepository) CreateCoachPayment(ctx context.Context, p *domain.CoachPayment) error {
	query := `
		INSERT INTO coach_payments (coach_id, amount, payment_date, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, p.CoachID, p.Amount, p.PaymentDate, p.Description, p.Status).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *PostgresPaymentRepository) GetCoachPaymentByID(ctx context.Context, id string) (*domain.CoachPayment, error) {
	query := `
		SELECT id, coach_id, amount, payment_date, description, status, created_at, updated_at
		FROM coach_payments WHERE id = $1
	`
	var p domain.CoachPayment
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.CoachID, &p.Amount, &p.PaymentDate, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("payment record not found")
		}
		return nil, err
	}
	return &p, nil
}

func (r *PostgresPaymentRepository) UpdateCoachPayment(ctx context.Context, p *domain.CoachPayment) error {
	query := `
		UPDATE coach_payments
		SET amount = $1, payment_date = $2, description = $3, status = $4, updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query, p.Amount, p.PaymentDate, p.Description, p.Status, p.ID)
	return err
}

func (r *PostgresPaymentRepository) ListCoachPayments(ctx context.Context, coachID string) ([]*domain.CoachPayment, error) {
	var rows pgx.Rows
	var err error

	if coachID != "" {
		query := `
			SELECT id, coach_id, amount, payment_date, description, status, created_at, updated_at
			FROM coach_payments
			WHERE coach_id = $1
			ORDER BY payment_date DESC, created_at DESC
		`
		rows, err = r.db.Query(ctx, query, coachID)
	} else {
		query := `
			SELECT id, coach_id, amount, payment_date, description, status, created_at, updated_at
			FROM coach_payments
			ORDER BY payment_date DESC, created_at DESC
		`
		rows, err = r.db.Query(ctx, query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.CoachPayment
	for rows.Next() {
		var p domain.CoachPayment
		err := rows.Scan(&p.ID, &p.CoachID, &p.Amount, &p.PaymentDate, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, nil
}

func (r *PostgresPaymentRepository) DeleteCoachPayment(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM coach_payments WHERE id = $1", id)
	return err
}

// Student payments (Coach managed)

func (r *PostgresPaymentRepository) CreateStudentPayment(ctx context.Context, p *domain.StudentPayment) error {
	query := `
		INSERT INTO student_payments (student_id, coach_id, amount, payment_date, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, p.StudentID, p.CoachID, p.Amount, p.PaymentDate, p.Description, p.Status).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *PostgresPaymentRepository) GetStudentPaymentByID(ctx context.Context, id string) (*domain.StudentPayment, error) {
	query := `
		SELECT id, student_id, coach_id, amount, payment_date, description, status, created_at, updated_at
		FROM student_payments WHERE id = $1
	`
	var p domain.StudentPayment
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.StudentID, &p.CoachID, &p.Amount, &p.PaymentDate, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("payment record not found")
		}
		return nil, err
	}
	return &p, nil
}

func (r *PostgresPaymentRepository) UpdateStudentPayment(ctx context.Context, p *domain.StudentPayment) error {
	query := `
		UPDATE student_payments
		SET amount = $1, payment_date = $2, description = $3, status = $4, updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query, p.Amount, p.PaymentDate, p.Description, p.Status, p.ID)
	return err
}

func (r *PostgresPaymentRepository) ListStudentPayments(ctx context.Context, studentID string) ([]*domain.StudentPayment, error) {
	query := `
		SELECT id, student_id, coach_id, amount, payment_date, description, status, created_at, updated_at
		FROM student_payments
		WHERE student_id = $1
		ORDER BY payment_date DESC, created_at DESC
	`
	rows, err := r.db.Query(ctx, query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.StudentPayment
	for rows.Next() {
		var p domain.StudentPayment
		err := rows.Scan(&p.ID, &p.StudentID, &p.CoachID, &p.Amount, &p.PaymentDate, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, nil
}

func (r *PostgresPaymentRepository) ListStudentPaymentsByCoach(ctx context.Context, coachID string) ([]*domain.StudentPayment, error) {
	query := `
		SELECT id, student_id, coach_id, amount, payment_date, description, status, created_at, updated_at
		FROM student_payments
		WHERE coach_id = $1
		ORDER BY payment_date DESC, created_at DESC
	`
	rows, err := r.db.Query(ctx, query, coachID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.StudentPayment
	for rows.Next() {
		var p domain.StudentPayment
		err := rows.Scan(&p.ID, &p.StudentID, &p.CoachID, &p.Amount, &p.PaymentDate, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, nil
}

func (r *PostgresPaymentRepository) DeleteStudentPayment(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM student_payments WHERE id = $1", id)
	return err
}
