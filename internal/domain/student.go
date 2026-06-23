package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrStudentNotFound = errors.New("student not found")
)

type Student struct {
	ID          string    `json:"id"`
	User        *User     `json:"user,omitempty"`
	CoachID     string    `json:"coach_id"`
	CoachName   string    `json:"coach_name"`
	ClassLevel  string    `json:"class_level"` // e.g. 12, mezun
	StudyTrack  string    `json:"study_track"`  // 'Sayisal', 'Sozel', 'Esit Agirlik', 'Dil'
	ExamTypeID   string    `json:"exam_type_id"`
	ExamTypeName string    `json:"exam_type_name"`
	ExamType     *ExamType `json:"exam_type,omitempty"`
	IsArchived   bool      `json:"is_archived"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type StudentRepository interface {
	CreateStudent(ctx context.Context, student *Student) error
	GetStudentByID(ctx context.Context, id string) (*Student, error)
	UpdateStudent(ctx context.Context, student *Student) error
	ListStudentsByCoach(ctx context.Context, coachID string, includeArchived bool) ([]*Student, error)
	ListAllStudents(ctx context.Context) ([]*Student, error)

	// Transfer student
	TransferStudent(ctx context.Context, studentID, newCoachID string) error

	// Capacity checks
	GetCoachStudentCount(ctx context.Context, coachID string) (int, error)
}
