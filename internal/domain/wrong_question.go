package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type WrongQuestion struct {
	ID            uuid.UUID  `json:"id"`
	StudentID     uuid.UUID  `json:"studentId"`
	CreatorUserID uuid.UUID  `json:"creatorUserId"`
	CreatorName   string     `json:"creatorName"`
	SubjectID     *uuid.UUID `json:"subjectId,omitempty"`
	SubjectName   string     `json:"subjectName"`
	ImageURL      string     `json:"imageUrl"`
	Title         string     `json:"title"`
	Note          string     `json:"note"`
	IsResolved    bool       `json:"isResolved"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type CreateWrongQuestionInput struct {
	StudentID     uuid.UUID
	CreatorUserID uuid.UUID
	SubjectID     *uuid.UUID
	ImageURL      string
	Title         string
	Note          string
}

type UpdateWrongQuestionInput struct {
	ID        uuid.UUID  `json:"id"`
	SubjectID *uuid.UUID `json:"subjectId,omitempty"`
	Title     string     `json:"title"`
	Note      string     `json:"note"`
}

type ToggleResolvedInput struct {
	ID         uuid.UUID `json:"id"`
	IsResolved bool      `json:"isResolved"`
}

type WrongQuestionRepository interface {
	Create(ctx context.Context, input CreateWrongQuestionInput) (*WrongQuestion, error)
	ListByStudent(ctx context.Context, studentID uuid.UUID, subjectID *uuid.UUID) ([]WrongQuestion, error)
	GetByID(ctx context.Context, id uuid.UUID) (*WrongQuestion, error)
	Update(ctx context.Context, input UpdateWrongQuestionInput) error
	ToggleResolved(ctx context.Context, id uuid.UUID, isResolved bool) error
	Delete(ctx context.Context, id uuid.UUID) error
}
