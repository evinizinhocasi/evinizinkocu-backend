package application

import (
	"context"

	"evinizinkocu-backend/internal/domain"
)

type CMSService struct {
	repo domain.CMSRepository
}

func NewCMSService(repo domain.CMSRepository) *CMSService {
	return &CMSService{repo: repo}
}

// Hero Settings
func (s *CMSService) GetHeroSettings(ctx context.Context) (*domain.HeroSettings, error) {
	var settings domain.HeroSettings
	err := s.repo.GetSettings(ctx, "hero_settings", &settings)
	if err != nil {
		// Return defaults
		return &domain.HeroSettings{
			HeroTitle:    "Evinizin Koçu",
			HeroSubtitle: "Başarıya Giden Yolda Güvenilir Rehberiniz",
			ContactEmail: "destek@evinizinkocu.com",
			ContactPhone: "+90 500 000 00 00",
		}, nil
	}
	return &settings, nil
}

func (s *CMSService) SaveHeroSettings(ctx context.Context, settings *domain.HeroSettings) error {
	return s.repo.SaveSettings(ctx, "hero_settings", settings)
}

// FAQs
func (s *CMSService) CreateFAQ(ctx context.Context, question, answer string, order int) (*domain.FAQ, error) {
	faq := &domain.FAQ{
		Question:     question,
		Answer:       answer,
		DisplayOrder: order,
	}
	err := s.repo.CreateFAQ(ctx, faq)
	return faq, err
}

func (s *CMSService) UpdateFAQ(ctx context.Context, id, question, answer string, order int) error {
	faq, err := s.repo.GetFAQByID(ctx, id)
	if err != nil {
		return err
	}
	faq.Question = question
	faq.Answer = answer
	faq.DisplayOrder = order
	return s.repo.UpdateFAQ(ctx, faq)
}

func (s *CMSService) DeleteFAQ(ctx context.Context, id string) error {
	return s.repo.DeleteFAQ(ctx, id)
}

func (s *CMSService) ListFAQs(ctx context.Context) ([]*domain.FAQ, error) {
	return s.repo.ListFAQs(ctx)
}

// Legal Documents
func (s *CMSService) SaveLegalDocument(ctx context.Context, slug, title, content string) error {
	doc := &domain.LegalDocument{
		Slug:    slug,
		Title:   title,
		Content: content,
	}
	return s.repo.SaveLegalDocument(ctx, doc)
}

func (s *CMSService) GetLegalDocument(ctx context.Context, slug string) (*domain.LegalDocument, error) {
	return s.repo.GetLegalDocument(ctx, slug)
}
