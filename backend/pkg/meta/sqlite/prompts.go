package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"ragpack/pkg/meta"
)

func (s *MetaStore) CreatePrompt(ctx context.Context, input meta.CreatePromptInput) (meta.Prompt, error) {
	now := time.Now().UTC()
	p := meta.Prompt{
		ID:        uuid.New().String(),
		Name:      input.Name,
		Slug:      slugify(input.Name),
		Content:   input.Content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO prompts (id, name, slug, content, created_at, updated_at)
		VALUES (:id, :name, :slug, :content, :created_at, :updated_at)
	`, p)
	if err != nil {
		return meta.Prompt{}, fmt.Errorf("sqlite: create prompt %q: %w", input.Name, err)
	}
	return p, nil
}

func (s *MetaStore) GetPromptBySlug(ctx context.Context, slug string) (meta.Prompt, error) {
	var p meta.Prompt
	err := s.db.GetContext(ctx, &p, `
		SELECT id, name, slug, content, created_at, updated_at
		FROM prompts WHERE slug = ?
	`, slug)
	if err != nil {
		return meta.Prompt{}, fmt.Errorf("sqlite: get prompt by slug %q: %w", slug, err)
	}
	return p, nil
}

func (s *MetaStore) ListPrompts(ctx context.Context, limit, offset int) ([]meta.Prompt, error) {
	var prompts []meta.Prompt
	err := s.db.SelectContext(ctx, &prompts, `
		SELECT id, name, slug, content, created_at, updated_at
		FROM prompts ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list prompts: %w", err)
	}
	return prompts, nil
}

func (s *MetaStore) CountPrompts(ctx context.Context) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM prompts`)
	if err != nil {
		return 0, fmt.Errorf("sqlite: count prompts: %w", err)
	}
	return count, nil
}

func (s *MetaStore) UpdatePrompt(ctx context.Context, slug string, input meta.UpdatePromptInput) (meta.Prompt, error) {
	p, err := s.GetPromptBySlug(ctx, slug)
	if err != nil {
		return meta.Prompt{}, err
	}

	if input.Name != nil {
		p.Name = *input.Name
		p.Slug = slugify(*input.Name)
	}
	if input.Content != nil {
		p.Content = *input.Content
	}
	p.UpdatedAt = time.Now().UTC()

	_, err = s.db.ExecContext(ctx, `
		UPDATE prompts SET name = ?, slug = ?, content = ?, updated_at = ? WHERE id = ?
	`, p.Name, p.Slug, p.Content, p.UpdatedAt, p.ID)
	if err != nil {
		return meta.Prompt{}, fmt.Errorf("sqlite: update prompt %q: %w", slug, err)
	}
	return p, nil
}

func (s *MetaStore) DeletePrompt(ctx context.Context, slug string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM prompts WHERE slug = ?`, slug)
	if err != nil {
		return fmt.Errorf("sqlite: delete prompt %q: %w", slug, err)
	}
	return nil
}
