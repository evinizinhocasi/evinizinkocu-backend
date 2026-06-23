package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"evinizinkocu-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresExamRepository struct {
	db *pgxpool.Pool
}

func NewPostgresExamRepository(db *pgxpool.Pool) domain.TrialExamRepository {
	return &PostgresExamRepository{db: db}
}

func (r *PostgresExamRepository) CreateExam(ctx context.Context, e *domain.TrialExam) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO trial_exams (student_id, creator_id, exam_name, exam_date, exam_type_id, total_net, score, ranking, coach_comment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRow(ctx, query, e.StudentID, e.CreatorID, e.ExamName, e.ExamDate, e.ExamTypeID, e.TotalNet, e.Score, e.Ranking, e.CoachComment).
		Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return err
	}

	// Save subject results
	for _, res := range e.Results {
		res.TrialExamID = e.ID
		subQuery := `
			INSERT INTO trial_exam_subject_results (trial_exam_id, subject_id, correct, incorrect, blank, net)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`
		err = tx.QueryRow(ctx, subQuery, res.TrialExamID, res.SubjectID, res.Correct, res.Incorrect, res.Blank, res.Net).Scan(&res.ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresExamRepository) GetExamByID(ctx context.Context, id string) (*domain.TrialExam, error) {
	query := `
		SELECT id, student_id, creator_id, exam_name, exam_date, exam_type_id, total_net, score, ranking, coach_comment, created_at, updated_at
		FROM trial_exams
		WHERE id = $1
	`
	var e domain.TrialExam
	err := r.db.QueryRow(ctx, query, id).Scan(
		&e.ID, &e.StudentID, &e.CreatorID, &e.ExamName, &e.ExamDate, &e.ExamTypeID, &e.TotalNet, &e.Score, &e.Ranking, &e.CoachComment, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("trial exam not found")
		}
		return nil, err
	}

	// Fetch results
	resQuery := `
		SELECT r.id, r.trial_exam_id, r.subject_id, r.correct, r.incorrect, r.blank, r.net, s.name
		FROM trial_exam_subject_results r
		JOIN subjects s ON r.subject_id = s.id
		WHERE r.trial_exam_id = $1
	`
	rows, err := r.db.Query(ctx, resQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var res domain.TrialExamSubjectResult
		var subName string
		err := rows.Scan(&res.ID, &res.TrialExamID, &res.SubjectID, &res.Correct, &res.Incorrect, &res.Blank, &res.Net, &subName)
		if err != nil {
			return nil, err
		}
		res.Subject = &domain.Subject{ID: res.SubjectID, Name: subName}
		e.Results = append(e.Results, &res)
	}

	return &e, nil
}

func (r *PostgresExamRepository) ListExamsByStudent(ctx context.Context, studentID string, examTypeID string, startDate, endDate *time.Time) ([]*domain.TrialExam, error) {
	filters := "student_id = $1"
	args := []interface{}{studentID}
	placeholderIdx := 2

	if examTypeID != "" {
		filters += fmt.Sprintf(" AND exam_type_id = $%d", placeholderIdx)
		args = append(args, examTypeID)
		placeholderIdx++
	}
	if startDate != nil {
		filters += fmt.Sprintf(" AND exam_date >= $%d", placeholderIdx)
		args = append(args, *startDate)
		placeholderIdx++
	}
	if endDate != nil {
		filters += fmt.Sprintf(" AND exam_date <= $%d", placeholderIdx)
		args = append(args, *endDate)
		placeholderIdx++
	}

	query := fmt.Sprintf(`
		SELECT id, student_id, creator_id, exam_name, exam_date, exam_type_id, total_net, score, ranking, coach_comment, created_at, updated_at
		FROM trial_exams
		WHERE %s
		ORDER BY exam_date DESC, created_at DESC
	`, filters)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.TrialExam
	for rows.Next() {
		var e domain.TrialExam
		err := rows.Scan(
			&e.ID, &e.StudentID, &e.CreatorID, &e.ExamName, &e.ExamDate, &e.ExamTypeID, &e.TotalNet, &e.Score, &e.Ranking, &e.CoachComment, &e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, &e)
	}

	// Fetch results for all exams in batch or loop (since list size is reasonable)
	for _, e := range list {
		resQuery := `
			SELECT r.id, r.trial_exam_id, r.subject_id, r.correct, r.incorrect, r.blank, r.net, s.name
			FROM trial_exam_subject_results r
			JOIN subjects s ON r.subject_id = s.id
			WHERE r.trial_exam_id = $1
		`
		subRows, err := r.db.Query(ctx, resQuery, e.ID)
		if err != nil {
			return nil, err
		}
		for subRows.Next() {
			var res domain.TrialExamSubjectResult
			var subName string
			err := subRows.Scan(&res.ID, &res.TrialExamID, &res.SubjectID, &res.Correct, &res.Incorrect, &res.Blank, &res.Net, &subName)
			if err != nil {
				subRows.Close()
				return nil, err
			}
			res.Subject = &domain.Subject{ID: res.SubjectID, Name: subName}
			e.Results = append(e.Results, &res)
		}
		subRows.Close()
	}

	return list, nil
}

func (r *PostgresExamRepository) UpdateExam(ctx context.Context, e *domain.TrialExam) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE trial_exams
		SET exam_name = $1, exam_date = $2, total_net = $3, score = $4, ranking = $5, coach_comment = $6, updated_at = NOW()
		WHERE id = $7
	`
	_, err = tx.Exec(ctx, query, e.ExamName, e.ExamDate, e.TotalNet, e.Score, e.Ranking, e.CoachComment, e.ID)
	if err != nil {
		return err
	}

	// Clean existing results
	_, err = tx.Exec(ctx, "DELETE FROM trial_exam_subject_results WHERE trial_exam_id = $1", e.ID)
	if err != nil {
		return err
	}

	// Save new results
	for _, res := range e.Results {
		res.TrialExamID = e.ID
		subQuery := `
			INSERT INTO trial_exam_subject_results (trial_exam_id, subject_id, correct, incorrect, blank, net)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`
		err = tx.QueryRow(ctx, subQuery, res.TrialExamID, res.SubjectID, res.Correct, res.Incorrect, res.Blank, res.Net).Scan(&res.ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresExamRepository) DeleteExam(ctx context.Context, id string) error {
	query := `DELETE FROM trial_exams WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresExamRepository) SaveExamSubjectResults(ctx context.Context, trialExamID string, results []*domain.TrialExamSubjectResult) error {
	for _, res := range results {
		query := `
			INSERT INTO trial_exam_subject_results (trial_exam_id, subject_id, correct, incorrect, blank, net)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err := r.db.Exec(ctx, query, trialExamID, res.SubjectID, res.Correct, res.Incorrect, res.Blank, res.Net)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresExamRepository) DeleteExamSubjectResults(ctx context.Context, trialExamID string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM trial_exam_subject_results WHERE trial_exam_id = $1", trialExamID)
	return err
}
