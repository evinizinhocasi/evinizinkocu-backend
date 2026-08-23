package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TargetSchoolType string

const (
	TargetSchoolTypeHighSchool TargetSchoolType = "high_school"
	TargetSchoolTypeUniversity TargetSchoolType = "university"
)

type TargetSchool struct {
	ID         uuid.UUID        `json:"id"`
	Name       string           `json:"name"`
	Type       TargetSchoolType `json:"type"` // "high_school" or "university"
	City       string           `json:"city"`
	PhotoURL   string           `json:"photoUrl"`
	MinScore   float64          `json:"minScore"`
	Percentile float64          `json:"percentile"`
	Ranking    int              `json:"ranking"`
	Department string           `json:"department"`
	IsActive   bool             `json:"isActive"`
	CreatedAt  time.Time        `json:"createdAt"`
	UpdatedAt  time.Time        `json:"updatedAt"`
}

type CreateTargetSchoolInput struct {
	Name       string
	Type       TargetSchoolType
	City       string
	PhotoURL   string
	MinScore   float64
	Percentile float64
	Ranking    int
	Department string
}

type UpdateTargetSchoolInput struct {
	ID         uuid.UUID
	Name       string
	Type       TargetSchoolType
	City       string
	PhotoURL   string
	MinScore   float64
	Percentile float64
	Ranking    int
	Department string
}

type UpdateStudentTargetSchoolInput struct {
	StudentID         uuid.UUID  `json:"studentId"`
	TargetSchoolID    *uuid.UUID `json:"targetSchoolId"`
	TargetSchoolName  string     `json:"targetSchoolName"`
	TargetSchoolPhoto string     `json:"targetSchoolPhoto"`
	TargetSchoolType  string     `json:"targetSchoolType"`
}

type TargetSchoolRepository interface {
	List(ctx context.Context, schoolType string, search string) ([]TargetSchool, error)
	GetByID(ctx context.Context, id uuid.UUID) (*TargetSchool, error)
	Create(ctx context.Context, input CreateTargetSchoolInput) (*TargetSchool, error)
	Update(ctx context.Context, input UpdateTargetSchoolInput) (*TargetSchool, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
