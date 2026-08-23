package repository

import (
	"context"
	"errors"
	"fmt"

	"evinizinkocu-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStudentRepository struct {
	db *pgxpool.Pool
}

func NewPostgresStudentRepository(db *pgxpool.Pool) domain.StudentRepository {
	return &PostgresStudentRepository{db: db}
}

func (r *PostgresStudentRepository) CreateStudent(ctx context.Context, s *domain.Student) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO students (
			id,
			coach_id,
			class_level,
			study_track,
			exam_type_id,
			target_school_id,
			target_school_name,
			target_school_photo,
			target_school_type,
			is_archived,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
	`
	_, err = tx.Exec(ctx, query,
		s.ID,
		s.CoachID,
		s.ClassLevel,
		s.StudyTrack,
		s.ExamTypeID,
		s.TargetSchoolID,
		s.TargetSchoolName,
		s.TargetSchoolPhoto,
		s.TargetSchoolType,
		s.IsArchived,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PostgresStudentRepository) GetStudentByID(ctx context.Context, id string) (*domain.Student, error) {
	query := `
		SELECT s.id, s.coach_id, s.class_level, s.study_track, s.exam_type_id,
		       s.target_school_id, COALESCE(s.target_school_name, ''), COALESCE(s.target_school_photo, ''), COALESCE(s.target_school_type, ''),
		       s.is_archived, s.created_at, s.updated_at,
		       u.id, u.email, u.username, u.phone, u.first_name, u.last_name, u.role, u.is_active, u.must_change_password,
		       COALESCE(et.name, ''),
		       COALESCE(cu.first_name || ' ' || cu.last_name, 'Atanmamış')
		FROM students s
		JOIN users u ON s.id = u.id
		LEFT JOIN exam_types et ON s.exam_type_id = et.id
		LEFT JOIN users cu ON s.coach_id = cu.id
		WHERE s.id = $1
	`
	var s domain.Student
	var u domain.User
	var targetSchoolID *string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.CoachID, &s.ClassLevel, &s.StudyTrack, &s.ExamTypeID,
		&targetSchoolID, &s.TargetSchoolName, &s.TargetSchoolPhoto, &s.TargetSchoolType,
		&s.IsArchived, &s.CreatedAt, &s.UpdatedAt,
		&u.ID, &u.Email, &u.Username, &u.Phone, &u.FirstName, &u.LastName, &u.Role, &u.IsActive, &u.MustChangePassword,
		&s.ExamTypeName, &s.CoachName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrStudentNotFound
		}
		return nil, err
	}
	s.TargetSchoolID = targetSchoolID
	s.User = &u
	return &s, nil
}

func (r *PostgresStudentRepository) UpdateStudent(ctx context.Context, s *domain.Student) error {
	query := `
		UPDATE students
		SET coach_id = $1, class_level = $2, study_track = $3, exam_type_id = $4, is_archived = $5, updated_at = NOW()
		WHERE id = $6
	`
	_, err := r.db.Exec(ctx, query, s.CoachID, s.ClassLevel, s.StudyTrack, s.ExamTypeID, s.IsArchived, s.ID)
	return err
}

func (r *PostgresStudentRepository) UpdateTargetSchool(
	ctx context.Context,
	studentID string,
	targetSchoolID *string,
	targetSchoolName string,
	targetSchoolPhoto string,
	targetSchoolType string,
) error {
	query := `
		UPDATE students
		SET
			target_school_id = $1,
			target_school_name = $2,
			target_school_photo = $3,
			target_school_type = $4,
			updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query, targetSchoolID, targetSchoolName, targetSchoolPhoto, targetSchoolType, studentID)
	return err
}

func (r *PostgresStudentRepository) ListStudentsByCoach(ctx context.Context, coachID string, includeArchived bool) ([]*domain.Student, error) {
	archiveFilter := "AND s.is_archived = FALSE"
	if includeArchived {
		archiveFilter = ""
	}

	query := fmt.Sprintf(`
		SELECT s.id, s.coach_id, s.class_level, s.study_track, s.exam_type_id,
		       s.target_school_id, COALESCE(s.target_school_name, ''), COALESCE(s.target_school_photo, ''), COALESCE(s.target_school_type, ''),
		       s.is_archived, s.created_at, s.updated_at,
		       u.id, u.email, u.username, u.phone, u.first_name, u.last_name, u.role, u.is_active, u.must_change_password,
		       COALESCE(et.name, ''),
		       COALESCE(cu.first_name || ' ' || cu.last_name, 'Atanmamış')
		FROM students s
		JOIN users u ON s.id = u.id
		LEFT JOIN exam_types et ON s.exam_type_id = et.id
		LEFT JOIN users cu ON s.coach_id = cu.id
		WHERE s.coach_id = $1 %s
		ORDER BY u.first_name, u.last_name
	`, archiveFilter)

	rows, err := r.db.Query(ctx, query, coachID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Student
	for rows.Next() {
		var s domain.Student
		var u domain.User
		var targetSchoolID *string
		err := rows.Scan(
			&s.ID, &s.CoachID, &s.ClassLevel, &s.StudyTrack, &s.ExamTypeID,
			&targetSchoolID, &s.TargetSchoolName, &s.TargetSchoolPhoto, &s.TargetSchoolType,
			&s.IsArchived, &s.CreatedAt, &s.UpdatedAt,
			&u.ID, &u.Email, &u.Username, &u.Phone, &u.FirstName, &u.LastName, &u.Role, &u.IsActive, &u.MustChangePassword,
			&s.ExamTypeName, &s.CoachName,
		)
		if err != nil {
			return nil, err
		}
		s.TargetSchoolID = targetSchoolID
		s.User = &u
		list = append(list, &s)
	}
	return list, nil
}

func (r *PostgresStudentRepository) ListAllStudents(ctx context.Context) ([]*domain.Student, error) {
	query := `
		SELECT s.id, s.coach_id, s.class_level, s.study_track, s.exam_type_id,
		       s.target_school_id, COALESCE(s.target_school_name, ''), COALESCE(s.target_school_photo, ''), COALESCE(s.target_school_type, ''),
		       s.is_archived, s.created_at, s.updated_at,
		       u.id, u.email, u.username, u.phone, u.first_name, u.last_name, u.role, u.is_active, u.must_change_password,
		       COALESCE(et.name, ''),
		       COALESCE(cu.first_name || ' ' || cu.last_name, 'Atanmamış')
		FROM students s
		JOIN users u ON s.id = u.id
		LEFT JOIN exam_types et ON s.exam_type_id = et.id
		LEFT JOIN users cu ON s.coach_id = cu.id
		ORDER BY u.first_name, u.last_name
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Student
	for rows.Next() {
		var s domain.Student
		var u domain.User
		var targetSchoolID *string
		err := rows.Scan(
			&s.ID, &s.CoachID, &s.ClassLevel, &s.StudyTrack, &s.ExamTypeID,
			&targetSchoolID, &s.TargetSchoolName, &s.TargetSchoolPhoto, &s.TargetSchoolType,
			&s.IsArchived, &s.CreatedAt, &s.UpdatedAt,
			&u.ID, &u.Email, &u.Username, &u.Phone, &u.FirstName, &u.LastName, &u.Role, &u.IsActive, &u.MustChangePassword,
			&s.ExamTypeName, &s.CoachName,
		)
		if err != nil {
			return nil, err
		}
		s.TargetSchoolID = targetSchoolID
		s.User = &u
		list = append(list, &s)
	}
	return list, nil
}

func (r *PostgresStudentRepository) TransferStudent(ctx context.Context, studentID, newCoachID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Verify coach exists and has capacity
	var capacity int
	err = tx.QueryRow(ctx, "SELECT student_capacity FROM coaches WHERE id = $1", newCoachID).Scan(&capacity)
	if err != nil {
		return fmt.Errorf("new coach not found: %w", err)
	}

	var currentCount int
	err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM students WHERE coach_id = $1", newCoachID).Scan(&currentCount)
	if err != nil {
		return err
	}

	if currentCount >= capacity {
		return domain.ErrCapacityExceeded
	}

	// Update student coach_id
	query := `UPDATE students SET coach_id = $1, updated_at = NOW() WHERE id = $2`
	_, err = tx.Exec(ctx, query, newCoachID, studentID)
	if err != nil {
		return err
	}

	// Transfer is transactional, commit now
	return tx.Commit(ctx)
}

func (r *PostgresStudentRepository) GetCoachStudentCount(ctx context.Context, coachID string) (int, error) {
	// Active/passive/archived non-deleted all count towards capacity
	query := `SELECT COUNT(*) FROM students WHERE coach_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, coachID).Scan(&count)
	return count, err
}
