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

type PostgresWrongQuestionRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresWrongQuestionRepository(pool *pgxpool.Pool) domain.WrongQuestionRepository {
	return &PostgresWrongQuestionRepository{pool: pool}
}

func (r *PostgresWrongQuestionRepository) Create(ctx context.Context, input domain.CreateWrongQuestionInput) (*domain.WrongQuestion, error) {
	query := `
		INSERT INTO wrong_questions (
			student_id,
			creator_user_id,
			subject_id,
			image_url,
			title,
			note
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, is_resolved, created_at, updated_at
	`

	var id uuid.UUID
	var isResolved bool
	var createdAt, updatedAt time.Time

	err := r.pool.QueryRow(ctx, query,
		input.StudentID,
		input.CreatorUserID,
		input.SubjectID,
		input.ImageURL,
		strings.TrimSpace(input.Title),
		strings.TrimSpace(input.Note),
	).Scan(&id, &isResolved, &createdAt, &updatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert wrong question: %w", err)
	}

	// Fetch creator and subject names
	var creatorFirstName, creatorLastName string
	_ = r.pool.QueryRow(ctx, "SELECT first_name, last_name FROM users WHERE id = $1", input.CreatorUserID).Scan(&creatorFirstName, &creatorLastName)
	creatorName := strings.TrimSpace(creatorFirstName + " " + creatorLastName)

	subjectName := "Genel"
	if input.SubjectID != nil {
		_ = r.pool.QueryRow(ctx, "SELECT name FROM subjects WHERE id = $1", *input.SubjectID).Scan(&subjectName)
	}

	return &domain.WrongQuestion{
		ID:            id,
		StudentID:     input.StudentID,
		CreatorUserID: input.CreatorUserID,
		CreatorName:   creatorName,
		SubjectID:     input.SubjectID,
		SubjectName:   subjectName,
		ImageURL:      input.ImageURL,
		Title:         input.Title,
		Note:          input.Note,
		IsResolved:    isResolved,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}, nil
}

func (r *PostgresWrongQuestionRepository) ListByStudent(ctx context.Context, studentID uuid.UUID, subjectID *uuid.UUID) ([]domain.WrongQuestion, error) {
	query := `
		SELECT
			wq.id,
			wq.student_id,
			wq.creator_user_id,
			CONCAT(cu.first_name, ' ', cu.last_name) AS creator_name,
			wq.subject_id,
			COALESCE(sub.name, 'Genel') AS subject_name,
			wq.image_url,
			wq.title,
			wq.note,
			wq.is_resolved,
			wq.created_at,
			wq.updated_at
		FROM wrong_questions wq
		JOIN users cu ON cu.id = wq.creator_user_id
		LEFT JOIN subjects sub ON sub.id = wq.subject_id
		WHERE wq.student_id = $1
	`
	args := []any{studentID}

	if subjectID != nil {
		query += " AND wq.subject_id = $2"
		args = append(args, *subjectID)
	}

	query += " ORDER BY wq.created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list wrong questions: %w", err)
	}
	defer rows.Close()

	var list []domain.WrongQuestion
	for rows.Next() {
		var item domain.WrongQuestion
		var subjectIDNull *uuid.UUID

		err := rows.Scan(
			&item.ID,
			&item.StudentID,
			&item.CreatorUserID,
			&item.CreatorName,
			&subjectIDNull,
			&item.SubjectName,
			&item.ImageURL,
			&item.Title,
			&item.Note,
			&item.IsResolved,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan wrong question: %w", err)
		}

		item.SubjectID = subjectIDNull
		list = append(list, item)
	}

	return list, nil
}

func (r *PostgresWrongQuestionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.WrongQuestion, error) {
	query := `
		SELECT
			wq.id,
			wq.student_id,
			wq.creator_user_id,
			CONCAT(cu.first_name, ' ', cu.last_name) AS creator_name,
			wq.subject_id,
			COALESCE(sub.name, 'Genel') AS subject_name,
			wq.image_url,
			wq.title,
			wq.note,
			wq.is_resolved,
			wq.created_at,
			wq.updated_at
		FROM wrong_questions wq
		JOIN users cu ON cu.id = wq.creator_user_id
		LEFT JOIN subjects sub ON sub.id = wq.subject_id
		WHERE wq.id = $1
	`

	var item domain.WrongQuestion
	var subjectIDNull *uuid.UUID

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.StudentID,
		&item.CreatorUserID,
		&item.CreatorName,
		&subjectIDNull,
		&item.SubjectName,
		&item.ImageURL,
		&item.Title,
		&item.Note,
		&item.IsResolved,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get wrong question by id: %w", err)
	}

	item.SubjectID = subjectIDNull
	return &item, nil
}

func (r *PostgresWrongQuestionRepository) Update(ctx context.Context, input domain.UpdateWrongQuestionInput) error {
	query := `
		UPDATE wrong_questions
		SET title = $1, note = $2, subject_id = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err := r.pool.Exec(ctx, query, strings.TrimSpace(input.Title), strings.TrimSpace(input.Note), input.SubjectID, input.ID)
	if err != nil {
		return fmt.Errorf("failed to update wrong question: %w", err)
	}
	return nil
}

func (r *PostgresWrongQuestionRepository) ToggleResolved(ctx context.Context, id uuid.UUID, isResolved bool) error {
	query := `
		UPDATE wrong_questions
		SET is_resolved = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.pool.Exec(ctx, query, isResolved, id)
	if err != nil {
		return fmt.Errorf("failed to toggle wrong question status: %w", err)
	}
	return nil
}

func (r *PostgresWrongQuestionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM wrong_questions WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete wrong question: %w", err)
	}
	return nil
}
