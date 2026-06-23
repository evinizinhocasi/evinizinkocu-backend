package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"evinizinkocu-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) domain.UserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, u *domain.User) error {
	email := strings.ToLower(strings.TrimSpace(u.Email))
	username := strings.ToLower(strings.TrimSpace(u.Username))
	phone := strings.TrimSpace(u.Phone)

	query := `
		INSERT INTO users (email, username, phone, password_hash, first_name, last_name, role, is_active, must_change_password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query, email, username, phone, u.PasswordHash, u.FirstName, u.LastName, u.Role, u.IsActive, u.MustChangePassword).
		Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "users_email_key") {
			return domain.ErrEmailConflict
		}
		if strings.Contains(err.Error(), "users_username_key") {
			return domain.ErrUsernameConflict
		}
		if strings.Contains(err.Error(), "users_phone_key") {
			return domain.ErrPhoneConflict
		}
		return err
	}
	return nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, email, username, phone, password_hash, first_name, last_name, role, is_active, must_change_password, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var u domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.Username, &u.Phone, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Role, &u.IsActive, &u.MustChangePassword, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, username, phone, password_hash, first_name, last_name, role, is_active, must_change_password, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var u domain.User
	err := r.db.QueryRow(ctx, query, strings.ToLower(strings.TrimSpace(email))).Scan(
		&u.ID, &u.Email, &u.Username, &u.Phone, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Role, &u.IsActive, &u.MustChangePassword, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, email, username, phone, password_hash, first_name, last_name, role, is_active, must_change_password, created_at, updated_at
		FROM users
		WHERE username = $1
	`
	var u domain.User
	err := r.db.QueryRow(ctx, query, strings.ToLower(strings.TrimSpace(username))).Scan(
		&u.ID, &u.Email, &u.Username, &u.Phone, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Role, &u.IsActive, &u.MustChangePassword, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	query := `
		SELECT id, email, username, phone, password_hash, first_name, last_name, role, is_active, must_change_password, created_at, updated_at
		FROM users
		WHERE phone = $1
	`
	var u domain.User
	err := r.db.QueryRow(ctx, query, strings.TrimSpace(phone)).Scan(
		&u.ID, &u.Email, &u.Username, &u.Phone, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Role, &u.IsActive, &u.MustChangePassword, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepository) GetByIdentifier(ctx context.Context, identifier string) (*domain.User, error) {
	normalized := strings.ToLower(strings.TrimSpace(identifier))
	query := `
		SELECT id, email, username, phone, password_hash, first_name, last_name, role, is_active, must_change_password, created_at, updated_at
		FROM users
		WHERE email = $1 OR username = $2
	`
	var u domain.User
	err := r.db.QueryRow(ctx, query, normalized, normalized).Scan(
		&u.ID, &u.Email, &u.Username, &u.Phone, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Role, &u.IsActive, &u.MustChangePassword, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, u *domain.User) error {
	query := `
		UPDATE users
		SET email = $1, username = $2, phone = $3, password_hash = $4, first_name = $5, last_name = $6, is_active = $7, must_change_password = $8, updated_at = NOW()
		WHERE id = $9
	`
	email := strings.ToLower(strings.TrimSpace(u.Email))
	username := strings.ToLower(strings.TrimSpace(u.Username))
	phone := strings.TrimSpace(u.Phone)

	_, err := r.db.Exec(ctx, query, email, username, phone, u.PasswordHash, u.FirstName, u.LastName, u.IsActive, u.MustChangePassword, u.ID)
	if err != nil {
		if strings.Contains(err.Error(), "users_email_key") {
			return domain.ErrEmailConflict
		}
		if strings.Contains(err.Error(), "users_username_key") {
			return domain.ErrUsernameConflict
		}
		if strings.Contains(err.Error(), "users_phone_key") {
			return domain.ErrPhoneConflict
		}
		return err
	}
	return nil
}

// Token methods
func (r *PostgresUserRepository) SaveRefreshToken(ctx context.Context, rt *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at, created_at, is_revoked)
		VALUES ($1, $2, $3, NOW(), $4)
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query, rt.UserID, rt.Token, rt.ExpiresAt, rt.IsRevoked).Scan(&rt.ID, &rt.CreatedAt)
}

func (r *PostgresUserRepository) GetRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at, is_revoked
		FROM refresh_tokens
		WHERE token = $1
	`
	var rt domain.RefreshToken
	err := r.db.QueryRow(ctx, query, token).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.CreatedAt, &rt.IsRevoked)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("refresh token not found")
		}
		return nil, err
	}
	return &rt, nil
}

func (r *PostgresUserRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	query := `UPDATE refresh_tokens SET is_revoked = TRUE WHERE token = $1`
	_, err := r.db.Exec(ctx, query, token)
	return err
}

func (r *PostgresUserRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	query := `UPDATE refresh_tokens SET is_revoked = TRUE WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// Reset password methods
func (r *PostgresUserRepository) SaveResetCode(ctx context.Context, userID, codeHash string, duration time.Duration) error {
	expiresAt := time.Now().Add(duration)
	query := `
		INSERT INTO password_reset_codes (user_id, code_hash, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err := r.db.Exec(ctx, query, userID, codeHash, expiresAt)
	return err
}

func (r *PostgresUserRepository) GetLatestActiveResetCode(ctx context.Context, userID string) (*domain.PasswordResetCode, error) {
	query := `
		SELECT id, user_id, code_hash, expires_at, created_at, used_at
		FROM password_reset_codes
		WHERE user_id = $1 AND used_at IS NULL AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`
	var prc domain.PasswordResetCode
	err := r.db.QueryRow(ctx, query, userID).Scan(&prc.ID, &prc.UserID, &prc.CodeHash, &prc.ExpiresAt, &prc.CreatedAt, &prc.UsedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("active reset code not found")
		}
		return nil, err
	}
	return &prc, nil
}

func (r *PostgresUserRepository) MarkResetCodeUsed(ctx context.Context, codeID string) error {
	query := `UPDATE password_reset_codes SET used_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, codeID)
	return err
}

// Device tokens methods
func (r *PostgresUserRepository) SaveDeviceToken(ctx context.Context, dt *domain.DeviceToken) error {
	query := `
		INSERT INTO device_tokens (user_id, token, platform, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, token) DO NOTHING
		RETURNING id, created_at
	`
	var id string
	var createdAt time.Time
	err := r.db.QueryRow(ctx, query, dt.UserID, dt.Token, dt.Platform).Scan(&id, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Row existed, nothing inserted
			return nil
		}
		return err
	}
	dt.ID = id
	dt.CreatedAt = createdAt
	return nil
}

func (r *PostgresUserRepository) RemoveDeviceToken(ctx context.Context, userID, token string) error {
	query := `DELETE FROM device_tokens WHERE user_id = $1 AND token = $2`
	_, err := r.db.Exec(ctx, query, userID, token)
	return err
}

func (r *PostgresUserRepository) GetDeviceTokensByUser(ctx context.Context, userID string) ([]string, error) {
	query := `SELECT token FROM device_tokens WHERE user_id = $1`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}
