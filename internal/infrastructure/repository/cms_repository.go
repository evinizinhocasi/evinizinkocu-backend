package repository

import (
	"context"
	"encoding/json"
	"errors"

	"evinizinkocu-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresCMSRepository struct {
	db *pgxpool.Pool
}

func NewPostgresCMSRepository(db *pgxpool.Pool) domain.CMSRepository {
	return &PostgresCMSRepository{db: db}
}

func (r *PostgresCMSRepository) SaveSettings(ctx context.Context, key string, val interface{}) error {
	valBytes, err := json.Marshal(val)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO cms_settings (key, value, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE
		SET value = EXCLUDED.value, updated_at = NOW()
	`
	_, err = r.db.Exec(ctx, query, key, valBytes)
	return err
}

func (r *PostgresCMSRepository) GetSettings(ctx context.Context, key string, target interface{}) error {
	query := `SELECT value FROM cms_settings WHERE key = $1`
	var valBytes []byte
	err := r.db.QueryRow(ctx, query, key).Scan(&valBytes)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("settings key not found")
		}
		return err
	}

	return json.Unmarshal(valBytes, target)
}

// FAQs

func (r *PostgresCMSRepository) CreateFAQ(ctx context.Context, faq *domain.FAQ) error {
	query := `
		INSERT INTO faqs (question, answer, display_order, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id
	`
	return r.db.QueryRow(ctx, query, faq.Question, faq.Answer, faq.DisplayOrder).Scan(&faq.ID)
}

func (r *PostgresCMSRepository) GetFAQByID(ctx context.Context, id string) (*domain.FAQ, error) {
	query := `SELECT id, question, answer, display_order FROM faqs WHERE id = $1`
	var faq domain.FAQ
	err := r.db.QueryRow(ctx, query, id).Scan(&faq.ID, &faq.Question, &faq.Answer, &faq.DisplayOrder)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("FAQ not found")
		}
		return nil, err
	}
	return &faq, nil
}

func (r *PostgresCMSRepository) UpdateFAQ(ctx context.Context, faq *domain.FAQ) error {
	query := `
		UPDATE faqs
		SET question = $1, answer = $2, display_order = $3
		WHERE id = $4
	`
	_, err := r.db.Exec(ctx, query, faq.Question, faq.Answer, faq.DisplayOrder, faq.ID)
	return err
}

func (r *PostgresCMSRepository) DeleteFAQ(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM faqs WHERE id = $1", id)
	return err
}

func (r *PostgresCMSRepository) ListFAQs(ctx context.Context) ([]*domain.FAQ, error) {
	query := `SELECT id, question, answer, display_order FROM faqs ORDER BY display_order ASC, question ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.FAQ
	for rows.Next() {
		var faq domain.FAQ
		err := rows.Scan(&faq.ID, &faq.Question, &faq.Answer, &faq.DisplayOrder)
		if err != nil {
			return nil, err
		}
		list = append(list, &faq)
	}
	return list, nil
}

// Legal Documents

func (r *PostgresCMSRepository) SaveLegalDocument(ctx context.Context, doc *domain.LegalDocument) error {
	query := `
		INSERT INTO legal_documents (slug, title, content, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (slug) DO UPDATE
		SET title = EXCLUDED.title, content = EXCLUDED.content, updated_at = NOW()
	`
	_, err := r.db.Exec(ctx, query, doc.Slug, doc.Title, doc.Content)
	return err
}

func (r *PostgresCMSRepository) GetLegalDocument(ctx context.Context, slug string) (*domain.LegalDocument, error) {
	query := `SELECT slug, title, content, updated_at::text FROM legal_documents WHERE slug = $1`
	var doc domain.LegalDocument
	err := r.db.QueryRow(ctx, query, slug).Scan(&doc.Slug, &doc.Title, &doc.Content, &doc.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("document not found")
		}
		return nil, err
	}
	return &doc, nil
}
