package application

import (
	"context"

	"evinizinkocu-backend/internal/domain"
)

type CatalogService struct {
	repo domain.CatalogRepository
}

func NewCatalogService(repo domain.CatalogRepository) *CatalogService {
	return &CatalogService{repo: repo}
}

// GetFullCatalog fetches all nested catalog details
func (s *CatalogService) GetFullCatalog(ctx context.Context) ([]*domain.ExamType, error) {
	return s.repo.GetFullCatalog(ctx)
}

// ExamType Management
func (s *CatalogService) CreateExamType(ctx context.Context, name string, divisor int) (*domain.ExamType, error) {
	et := &domain.ExamType{
		Name:    name,
		Divisor: divisor,
	}
	err := s.repo.CreateExamType(ctx, et)
	return et, err
}

func (s *CatalogService) UpdateExamType(ctx context.Context, id, name string, divisor int) error {
	et, err := s.repo.GetExamTypeByID(ctx, id)
	if err != nil {
		return err
	}
	et.Name = name
	et.Divisor = divisor
	return s.repo.UpdateExamType(ctx, et)
}

func (s *CatalogService) DeleteExamType(ctx context.Context, id string) error {
	return s.repo.DeleteExamType(ctx, id)
}

// Subject Management
func (s *CatalogService) CreateSubject(ctx context.Context, examTypeID, name string, questionCount *int) (*domain.Subject, error) {
	sub := &domain.Subject{
		ExamTypeID:    examTypeID,
		Name:          name,
		QuestionCount: questionCount,
	}
	err := s.repo.CreateSubject(ctx, sub)
	return sub, err
}

func (s *CatalogService) UpdateSubject(ctx context.Context, id, name string, questionCount *int) error {
	sub, err := s.repo.GetSubjectByID(ctx, id)
	if err != nil {
		return err
	}
	sub.Name = name
	sub.QuestionCount = questionCount
	return s.repo.UpdateSubject(ctx, sub)
}

func (s *CatalogService) DeleteSubject(ctx context.Context, id string) error {
	return s.repo.DeleteSubject(ctx, id)
}

// Topic Management
func (s *CatalogService) CreateTopic(ctx context.Context, subjectID, name string) (*domain.Topic, error) {
	t := &domain.Topic{
		SubjectID: subjectID,
		Name:      name,
	}
	err := s.repo.CreateTopic(ctx, t)
	return t, err
}

func (s *CatalogService) UpdateTopic(ctx context.Context, id, name string) error {
	t, err := s.repo.GetTopicByID(ctx, id)
	if err != nil {
		return err
	}
	t.Name = name
	return s.repo.UpdateTopic(ctx, t)
}

func (s *CatalogService) DeleteTopic(ctx context.Context, id string) error {
	return s.repo.DeleteTopic(ctx, id)
}
