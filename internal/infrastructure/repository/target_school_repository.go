package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"evinizinkocu-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresTargetSchoolRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresTargetSchoolRepository(pool *pgxpool.Pool) domain.TargetSchoolRepository {
	return &PostgresTargetSchoolRepository{pool: pool}
}

func (r *PostgresTargetSchoolRepository) List(ctx context.Context, schoolType string, search string) ([]domain.TargetSchool, error) {
	query := `
		SELECT
			id,
			name,
			type,
			city,
			photo_url,
			min_score,
			percentile,
			ranking,
			department,
			is_active,
			created_at,
			updated_at
		FROM target_schools
		WHERE is_active = true
	`
	var args []any
	argIdx := 1

	if schoolType != "" && schoolType != "all" {
		query += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, schoolType)
		argIdx++
	}

	search = strings.TrimSpace(search)
	if search != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR city ILIKE $%d OR department ILIKE $%d)", argIdx, argIdx, argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}

	query += " ORDER BY ranking ASC, name ASC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query target schools: %w", err)
	}
	defer rows.Close()

	var list []domain.TargetSchool
	for rows.Next() {
		var item domain.TargetSchool
		var typeStr string

		err := rows.Scan(
			&item.ID,
			&item.Name,
			&typeStr,
			&item.City,
			&item.PhotoURL,
			&item.MinScore,
			&item.Percentile,
			&item.Ranking,
			&item.Department,
			&item.IsActive,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan target school: %w", err)
		}

		item.Type = domain.TargetSchoolType(typeStr)
		list = append(list, item)
	}

	if list == nil {
		list = []domain.TargetSchool{}
	}

	return list, nil
}

func (r *PostgresTargetSchoolRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.TargetSchool, error) {
	query := `
		SELECT
			id,
			name,
			type,
			city,
			photo_url,
			min_score,
			percentile,
			ranking,
			department,
			is_active,
			created_at,
			updated_at
		FROM target_schools
		WHERE id = $1
	`
	var item domain.TargetSchool
	var typeStr string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.Name,
		&typeStr,
		&item.City,
		&item.PhotoURL,
		&item.MinScore,
		&item.Percentile,
		&item.Ranking,
		&item.Department,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get target school: %w", err)
	}

	item.Type = domain.TargetSchoolType(typeStr)
	return &item, nil
}

func (r *PostgresTargetSchoolRepository) Create(ctx context.Context, input domain.CreateTargetSchoolInput) (*domain.TargetSchool, error) {
	query := `
		INSERT INTO target_schools (
			name,
			type,
			city,
			photo_url,
			min_score,
			percentile,
			ranking,
			department
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, is_active, created_at, updated_at
	`
	var id uuid.UUID
	var isActive bool
	var createdAt, updatedAt time.Time

	err := r.pool.QueryRow(ctx, query,
		strings.TrimSpace(input.Name),
		string(input.Type),
		strings.TrimSpace(input.City),
		strings.TrimSpace(input.PhotoURL),
		input.MinScore,
		input.Percentile,
		input.Ranking,
		strings.TrimSpace(input.Department),
	).Scan(&id, &isActive, &createdAt, &updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert target school: %w", err)
	}

	return &domain.TargetSchool{
		ID:         id,
		Name:       input.Name,
		Type:       input.Type,
		City:       input.City,
		PhotoURL:   input.PhotoURL,
		MinScore:   input.MinScore,
		Percentile: input.Percentile,
		Ranking:    input.Ranking,
		Department: input.Department,
		IsActive:   isActive,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

func (r *PostgresTargetSchoolRepository) Update(ctx context.Context, input domain.UpdateTargetSchoolInput) (*domain.TargetSchool, error) {
	query := `
		UPDATE target_schools
		SET
			name = $1,
			type = $2,
			city = $3,
			photo_url = $4,
			min_score = $5,
			percentile = $6,
			ranking = $7,
			department = $8,
			updated_at = NOW()
		WHERE id = $9
		RETURNING is_active, created_at, updated_at
	`
	var isActive bool
	var createdAt, updatedAt time.Time

	err := r.pool.QueryRow(ctx, query,
		strings.TrimSpace(input.Name),
		string(input.Type),
		strings.TrimSpace(input.City),
		strings.TrimSpace(input.PhotoURL),
		input.MinScore,
		input.Percentile,
		input.Ranking,
		strings.TrimSpace(input.Department),
		input.ID,
	).Scan(&isActive, &createdAt, &updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update target school: %w", err)
	}

	return &domain.TargetSchool{
		ID:         input.ID,
		Name:       input.Name,
		Type:       input.Type,
		City:       input.City,
		PhotoURL:   input.PhotoURL,
		MinScore:   input.MinScore,
		Percentile: input.Percentile,
		Ranking:    input.Ranking,
		Department: input.Department,
		IsActive:   isActive,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

func (r *PostgresTargetSchoolRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE target_schools SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to deactivate target school: %w", err)
	}
	return nil
}
