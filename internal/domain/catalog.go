package domain

import (
	"context"
	"errors"
)

var (
	ErrCatalogNotFound = errors.New("catalog item not found")
)

type ExamType struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`    // e.g. LGS, TYT, AYT, YDT
	Divisor  int        `json:"divisor"` // e.g. 3 or 4
	Subjects []*Subject `json:"subjects,omitempty"`
}

type Subject struct {
	ID            string   `json:"id"`
	ExamTypeID    string   `json:"exam_type_id"`
	Name          string   `json:"name"`
	QuestionCount *int     `json:"question_count,omitempty"`
	Topics        []*Topic `json:"topics,omitempty"`
}

type Topic struct {
	ID        string `json:"id"`
	SubjectID string `json:"subject_id"`
	Name      string `json:"name"`
}

type CatalogRepository interface {
	CreateExamType(ctx context.Context, et *ExamType) error
	GetExamTypeByID(ctx context.Context, id string) (*ExamType, error)
	GetExamTypeByName(ctx context.Context, name string) (*ExamType, error)
	ListExamTypes(ctx context.Context) ([]*ExamType, error)
	UpdateExamType(ctx context.Context, et *ExamType) error
	DeleteExamType(ctx context.Context, id string) error

	CreateSubject(ctx context.Context, s *Subject) error
	GetSubjectByID(ctx context.Context, id string) (*Subject, error)
	ListSubjectsByExamType(ctx context.Context, examTypeID string) ([]*Subject, error)
	UpdateSubject(ctx context.Context, s *Subject) error
	DeleteSubject(ctx context.Context, id string) error

	CreateTopic(ctx context.Context, t *Topic) error
	GetTopicByID(ctx context.Context, id string) (*Topic, error)
	ListTopicsBySubject(ctx context.Context, subjectID string) ([]*Topic, error)
	UpdateTopic(ctx context.Context, t *Topic) error
	DeleteTopic(ctx context.Context, id string) error

	GetFullCatalog(ctx context.Context) ([]*ExamType, error)
}
