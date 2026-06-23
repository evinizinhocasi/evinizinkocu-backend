package repository

import (
	"context"
	"errors"

	"evinizinkocu-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresCatalogRepository struct {
	db *pgxpool.Pool
}

func NewPostgresCatalogRepository(db *pgxpool.Pool) domain.CatalogRepository {
	return &PostgresCatalogRepository{db: db}
}

// ExamType CRUD
func (r *PostgresCatalogRepository) CreateExamType(ctx context.Context, et *domain.ExamType) error {
	query := `
		INSERT INTO exam_types (name, divisor, created_at)
		VALUES ($1, $2, NOW())
		RETURNING id
	`
	return r.db.QueryRow(ctx, query, et.Name, et.Divisor).Scan(&et.ID)
}

func (r *PostgresCatalogRepository) GetExamTypeByID(ctx context.Context, id string) (*domain.ExamType, error) {
	query := `SELECT id, name, divisor FROM exam_types WHERE id = $1`
	var et domain.ExamType
	err := r.db.QueryRow(ctx, query, id).Scan(&et.ID, &et.Name, &et.Divisor)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCatalogNotFound
		}
		return nil, err
	}
	return &et, nil
}

func (r *PostgresCatalogRepository) GetExamTypeByName(ctx context.Context, name string) (*domain.ExamType, error) {
	query := `SELECT id, name, divisor FROM exam_types WHERE name = $1`
	var et domain.ExamType
	err := r.db.QueryRow(ctx, query, name).Scan(&et.ID, &et.Name, &et.Divisor)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCatalogNotFound
		}
		return nil, err
	}
	return &et, nil
}

func (r *PostgresCatalogRepository) ListExamTypes(ctx context.Context) ([]*domain.ExamType, error) {
	query := `SELECT id, name, divisor FROM exam_types ORDER BY name`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.ExamType
	for rows.Next() {
		var et domain.ExamType
		if err := rows.Scan(&et.ID, &et.Name, &et.Divisor); err != nil {
			return nil, err
		}
		list = append(list, &et)
	}
	return list, nil
}

func (r *PostgresCatalogRepository) UpdateExamType(ctx context.Context, et *domain.ExamType) error {
	query := `UPDATE exam_types SET name = $1, divisor = $2 WHERE id = $3`
	_, err := r.db.Exec(ctx, query, et.Name, et.Divisor, et.ID)
	return err
}

func (r *PostgresCatalogRepository) DeleteExamType(ctx context.Context, id string) error {
	query := `DELETE FROM exam_types WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Subject CRUD
func (r *PostgresCatalogRepository) CreateSubject(ctx context.Context, s *domain.Subject) error {
	query := `
		INSERT INTO subjects (exam_type_id, name, question_count)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	return r.db.QueryRow(ctx, query, s.ExamTypeID, s.Name, s.QuestionCount).Scan(&s.ID)
}

func (r *PostgresCatalogRepository) GetSubjectByID(ctx context.Context, id string) (*domain.Subject, error) {
	query := `SELECT id, exam_type_id, name, question_count FROM subjects WHERE id = $1`
	var s domain.Subject
	err := r.db.QueryRow(ctx, query, id).Scan(&s.ID, &s.ExamTypeID, &s.Name, &s.QuestionCount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCatalogNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *PostgresCatalogRepository) ListSubjectsByExamType(ctx context.Context, examTypeID string) ([]*domain.Subject, error) {
	query := `SELECT id, exam_type_id, name, question_count FROM subjects WHERE exam_type_id = $1 ORDER BY name`
	rows, err := r.db.Query(ctx, query, examTypeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Subject
	for rows.Next() {
		var s domain.Subject
		if err := rows.Scan(&s.ID, &s.ExamTypeID, &s.Name, &s.QuestionCount); err != nil {
			return nil, err
		}
		list = append(list, &s)
	}
	return list, nil
}

func (r *PostgresCatalogRepository) UpdateSubject(ctx context.Context, s *domain.Subject) error {
	query := `UPDATE subjects SET name = $1, question_count = $2 WHERE id = $3`
	_, err := r.db.Exec(ctx, query, s.Name, s.QuestionCount, s.ID)
	return err
}

func (r *PostgresCatalogRepository) DeleteSubject(ctx context.Context, id string) error {
	query := `DELETE FROM subjects WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Topic CRUD
func (r *PostgresCatalogRepository) CreateTopic(ctx context.Context, t *domain.Topic) error {
	query := `
		INSERT INTO topics (subject_id, name)
		VALUES ($1, $2)
		RETURNING id
	`
	return r.db.QueryRow(ctx, query, t.SubjectID, t.Name).Scan(&t.ID)
}

func (r *PostgresCatalogRepository) GetTopicByID(ctx context.Context, id string) (*domain.Topic, error) {
	query := `SELECT id, subject_id, name FROM topics WHERE id = $1`
	var t domain.Topic
	err := r.db.QueryRow(ctx, query, id).Scan(&t.ID, &t.SubjectID, &t.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCatalogNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *PostgresCatalogRepository) ListTopicsBySubject(ctx context.Context, subjectID string) ([]*domain.Topic, error) {
	query := `SELECT id, subject_id, name FROM topics WHERE subject_id = $1 ORDER BY name`
	rows, err := r.db.Query(ctx, query, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Topic
	for rows.Next() {
		var t domain.Topic
		if err := rows.Scan(&t.ID, &t.SubjectID, &t.Name); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}
	return list, nil
}

func (r *PostgresCatalogRepository) UpdateTopic(ctx context.Context, t *domain.Topic) error {
	query := `UPDATE topics SET name = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, t.Name, t.ID)
	return err
}

func (r *PostgresCatalogRepository) DeleteTopic(ctx context.Context, id string) error {
	query := `DELETE FROM topics WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Optimized loader
func (r *PostgresCatalogRepository) GetFullCatalog(ctx context.Context) ([]*domain.ExamType, error) {
	// 1. Fetch all ExamTypes
	exams, err := r.ListExamTypes(ctx)
	if err != nil {
		return nil, err
	}

	examMap := make(map[string]*domain.ExamType)
	for _, e := range exams {
		e.Subjects = []*domain.Subject{}
		examMap[e.ID] = e
	}

	// 2. Fetch all Subjects
	subjectQuery := `SELECT id, exam_type_id, name, question_count FROM subjects ORDER BY name`
	subRows, err := r.db.Query(ctx, subjectQuery)
	if err != nil {
		return nil, err
	}
	defer subRows.Close()

	subjectMap := make(map[string]*domain.Subject)
	var subjects []*domain.Subject
	for subRows.Next() {
		var s domain.Subject
		if err := subRows.Scan(&s.ID, &s.ExamTypeID, &s.Name, &s.QuestionCount); err != nil {
			return nil, err
		}
		s.Topics = []*domain.Topic{}
		subjects = append(subjects, &s)
		subjectMap[s.ID] = &s
	}

	// 3. Fetch all Topics
	topicQuery := `SELECT id, subject_id, name FROM topics ORDER BY name`
	topicRows, err := r.db.Query(ctx, topicQuery)
	if err != nil {
		return nil, err
	}
	defer topicRows.Close()

	for topicRows.Next() {
		var t domain.Topic
		if err := topicRows.Scan(&t.ID, &t.SubjectID, &t.Name); err != nil {
			return nil, err
		}
		if parentSub, exists := subjectMap[t.SubjectID]; exists {
			parentSub.Topics = append(parentSub.Topics, &t)
		}
	}

	// Connect subjects to exams
	for _, s := range subjects {
		if parentExam, exists := examMap[s.ExamTypeID]; exists {
			parentExam.Subjects = append(parentExam.Subjects, s)
		}
	}

	return exams, nil
}
