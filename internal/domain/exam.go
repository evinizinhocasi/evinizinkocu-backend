package domain

import (
	"context"
	"time"
)

type TrialExam struct {
	ID           string                    `json:"id"`
	StudentID    string                    `json:"student_id"`
	CreatorID    string                    `json:"creator_id"`
	ExamName     string                    `json:"exam_name"`
	ExamDate     time.Time                 `json:"exam_date"`
	ExamTypeID   string                    `json:"exam_type_id"`
	ExamType     *ExamType                 `json:"exam_type,omitempty"`
	TotalNet     float64                   `json:"total_net"`
	Score        *float64                  `json:"score,omitempty"`
	Ranking      *int                      `json:"ranking,omitempty"`
	CoachComment string                    `json:"coach_comment,omitempty"`
	Results      []*TrialExamSubjectResult `json:"results,omitempty"`
	CreatedAt    time.Time                 `json:"created_at"`
	UpdatedAt    time.Time                 `json:"updated_at"`
}

type TrialExamSubjectResult struct {
	ID          string   `json:"id"`
	TrialExamID string   `json:"trial_exam_id"`
	SubjectID   string   `json:"subject_id"`
	Subject     *Subject `json:"subject,omitempty"`
	Correct     int      `json:"correct"`
	Incorrect   int      `json:"incorrect"`
	Blank       int      `json:"blank"`
	Net         float64  `json:"net"`
}

type TrialExamRepository interface {
	CreateExam(ctx context.Context, exam *TrialExam) error
	GetExamByID(ctx context.Context, id string) (*TrialExam, error)
	ListExamsByStudent(ctx context.Context, studentID string, examTypeID string, startDate, endDate *time.Time) ([]*TrialExam, error)
	UpdateExam(ctx context.Context, exam *TrialExam) error
	DeleteExam(ctx context.Context, id string) error

	// subject results
	SaveExamSubjectResults(ctx context.Context, trialExamID string, results []*TrialExamSubjectResult) error
	DeleteExamSubjectResults(ctx context.Context, trialExamID string) error
}
