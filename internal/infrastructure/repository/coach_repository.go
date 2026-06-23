package repository

import (
	"context"
	"errors"
	"strings"

	"evinizinkocu-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresCoachRepository struct {
	db *pgxpool.Pool
}

func NewPostgresCoachRepository(db *pgxpool.Pool) domain.CoachRepository {
	return &PostgresCoachRepository{db: db}
}

func (r *PostgresCoachRepository) CreateCoach(ctx context.Context, c *domain.Coach) error {
	query := `
		INSERT INTO coaches (id, city, biography, specialization, social_links, student_capacity, auth_start_date, auth_end_date, permission_immediate_push, permission_scheduled_push, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
	`
	// Handle default social_links as empty JSON if empty
	links := c.SocialLinks
	if links == "" {
		links = "{}"
	}

	_, err := r.db.Exec(ctx, query,
		c.ID, c.City, c.Biography, c.Specialization, links, c.StudentCapacity, c.AuthStartDate, c.AuthEndDate, c.PermissionImmediatePush, c.PermissionScheduledPush,
	)
	return err
}

func (r *PostgresCoachRepository) GetCoachByID(ctx context.Context, id string) (*domain.Coach, error) {
	query := `
		SELECT c.id, c.city, c.biography, c.specialization, c.social_links, c.student_capacity, c.auth_start_date, c.auth_end_date, c.permission_immediate_push, c.permission_scheduled_push, c.created_at, c.updated_at,
		       u.id, u.email, u.username, u.phone, u.first_name, u.last_name, u.role, u.is_active, u.must_change_password
		FROM coaches c
		JOIN users u ON c.id = u.id
		WHERE c.id = $1
	`
	var c domain.Coach
	var u domain.User
	var links []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.City, &c.Biography, &c.Specialization, &links, &c.StudentCapacity, &c.AuthStartDate, &c.AuthEndDate, &c.PermissionImmediatePush, &c.PermissionScheduledPush, &c.CreatedAt, &c.UpdatedAt,
		&u.ID, &u.Email, &u.Username, &u.Phone, &u.FirstName, &u.LastName, &u.Role, &u.IsActive, &u.MustChangePassword,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCoachNotFound
		}
		return nil, err
	}
	c.SocialLinks = string(links)
	c.User = &u
	return &c, nil
}

func (r *PostgresCoachRepository) UpdateCoach(ctx context.Context, c *domain.Coach) error {
	query := `
		UPDATE coaches
		SET city = $1, biography = $2, specialization = $3, social_links = $4, student_capacity = $5, auth_start_date = $6, auth_end_date = $7, permission_immediate_push = $8, permission_scheduled_push = $9, updated_at = NOW()
		WHERE id = $10
	`
	links := c.SocialLinks
	if links == "" {
		links = "{}"
	}

	_, err := r.db.Exec(ctx, query,
		c.City, c.Biography, c.Specialization, links, c.StudentCapacity, c.AuthStartDate, c.AuthEndDate, c.PermissionImmediatePush, c.PermissionScheduledPush, c.ID,
	)
	return err
}

func (r *PostgresCoachRepository) ListCoaches(ctx context.Context) ([]*domain.Coach, error) {
	query := `
		SELECT c.id, c.city, c.biography, c.specialization, c.social_links, c.student_capacity, c.auth_start_date, c.auth_end_date, c.permission_immediate_push, c.permission_scheduled_push, c.created_at, c.updated_at,
		       u.id, u.email, u.username, u.phone, u.first_name, u.last_name, u.role, u.is_active, u.must_change_password
		FROM coaches c
		JOIN users u ON c.id = u.id
		ORDER BY u.first_name, u.last_name
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Coach
	for rows.Next() {
		var c domain.Coach
		var u domain.User
		var links []byte
		err := rows.Scan(
			&c.ID, &c.City, &c.Biography, &c.Specialization, &links, &c.StudentCapacity, &c.AuthStartDate, &c.AuthEndDate, &c.PermissionImmediatePush, &c.PermissionScheduledPush, &c.CreatedAt, &c.UpdatedAt,
			&u.ID, &u.Email, &u.Username, &u.Phone, &u.FirstName, &u.LastName, &u.Role, &u.IsActive, &u.MustChangePassword,
		)
		if err != nil {
			return nil, err
		}
		c.SocialLinks = string(links)
		c.User = &u
		list = append(list, &c)
	}
	return list, nil
}

// Applications
func (r *PostgresCoachRepository) CreateApplication(ctx context.Context, app *domain.CoachApplication) error {
	query := `
		INSERT INTO coach_applications (first_name, last_name, phone, email, city, specialization, explanation, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		app.FirstName, app.LastName, strings.TrimSpace(app.Phone), strings.ToLower(strings.TrimSpace(app.Email)), app.City, app.Specialization, app.Explanation, app.Status,
	).Scan(&app.ID, &app.CreatedAt, &app.UpdatedAt)
}

func (r *PostgresCoachRepository) GetApplicationByID(ctx context.Context, id string) (*domain.CoachApplication, error) {
	query := `
		SELECT id, first_name, last_name, phone, email, city, specialization, explanation, status, created_at, updated_at
		FROM coach_applications
		WHERE id = $1
	`
	var app domain.CoachApplication
	err := r.db.QueryRow(ctx, query, id).Scan(
		&app.ID, &app.FirstName, &app.LastName, &app.Phone, &app.Email, &app.City, &app.Specialization, &app.Explanation, &app.Status, &app.CreatedAt, &app.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCoachApplicationNotFound
		}
		return nil, err
	}
	return &app, nil
}

func (r *PostgresCoachRepository) UpdateApplicationStatus(ctx context.Context, id string, status string) error {
	query := `
		UPDATE coach_applications
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}

func (r *PostgresCoachRepository) ListApplications(ctx context.Context) ([]*domain.CoachApplication, error) {
	query := `
		SELECT id, first_name, last_name, phone, email, city, specialization, explanation, status, created_at, updated_at
		FROM coach_applications
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.CoachApplication
	for rows.Next() {
		var app domain.CoachApplication
		err := rows.Scan(
			&app.ID, &app.FirstName, &app.LastName, &app.Phone, &app.Email, &app.City, &app.Specialization, &app.Explanation, &app.Status, &app.CreatedAt, &app.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, &app)
	}
	return list, nil
}

// History
func (r *PostgresCoachRepository) SaveAuthorizationPeriod(ctx context.Context, ap *domain.CoachAuthorizationPeriod) error {
	query := `
		INSERT INTO coach_authorization_periods (coach_id, start_date, end_date, student_capacity, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query, ap.CoachID, ap.StartDate, ap.EndDate, ap.StudentCapacity).Scan(&ap.ID, &ap.CreatedAt)
}

func (r *PostgresCoachRepository) GetAuthorizationPeriods(ctx context.Context, coachID string) ([]*domain.CoachAuthorizationPeriod, error) {
	query := `
		SELECT id, coach_id, start_date, end_date, student_capacity, created_at
		FROM coach_authorization_periods
		WHERE coach_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, coachID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.CoachAuthorizationPeriod
	for rows.Next() {
		var ap domain.CoachAuthorizationPeriod
		err := rows.Scan(&ap.ID, &ap.CoachID, &ap.StartDate, &ap.EndDate, &ap.StudentCapacity, &ap.CreatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, &ap)
	}
	return list, nil
}

// Expiry checks
func (r *PostgresCoachRepository) GetExpiredCoaches(ctx context.Context) ([]string, error) {
	query := `
		SELECT id FROM coaches
		WHERE auth_end_date < CURRENT_DATE
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *PostgresCoachRepository) DeactivateCoaches(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	query := `
		UPDATE users
		SET is_active = FALSE, updated_at = NOW()
		WHERE id = ANY($1) AND role = 'coach' AND is_active = TRUE
	`
	_, err := r.db.Exec(ctx, query, ids)
	return err
}
