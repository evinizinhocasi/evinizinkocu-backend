package domain

import (
	"context"
)

type FAQ struct {
	ID           string `json:"id"`
	Question     string `json:"question"`
	Answer       string `json:"answer"`
	DisplayOrder int    `json:"display_order"`
}

type LegalDocument struct {
	Slug      string `json:"slug"` // 'privacy-policy', 'terms-of-use', 'kvkk'
	Title     string `json:"title"`
	Content   string `json:"content"`
	UpdatedAt string `json:"updated_at"`
}

type HeroSettings struct {
	HeroTitle    string `json:"hero_title"`
	HeroSubtitle string `json:"hero_subtitle"`
	ContactEmail string `json:"contact_email"`
	ContactPhone string `json:"contact_phone"`
	InstagramURL string `json:"instagram_url"`
	LinkedinURL  string `json:"linkedin_url"`
}

type CMSRepository interface {
	SaveSettings(ctx context.Context, key string, val interface{}) error
	GetSettings(ctx context.Context, key string, target interface{}) error

	// FAQs
	CreateFAQ(ctx context.Context, faq *FAQ) error
	GetFAQByID(ctx context.Context, id string) (*FAQ, error)
	UpdateFAQ(ctx context.Context, faq *FAQ) error
	DeleteFAQ(ctx context.Context, id string) error
	ListFAQs(ctx context.Context) ([]*FAQ, error)

	// Legal Documents
	SaveLegalDocument(ctx context.Context, doc *LegalDocument) error
	GetLegalDocument(ctx context.Context, slug string) (*LegalDocument, error)
}
