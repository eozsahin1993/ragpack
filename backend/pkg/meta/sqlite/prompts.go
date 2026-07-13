package sqlite

import (
	"context"
	"fmt"
	"strings"
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
		IsSystem:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO prompts (id, name, slug, content, is_system, created_at, updated_at)
		VALUES (:id, :name, :slug, :content, :is_system, :created_at, :updated_at)
	`, p)
	if err != nil {
		return meta.Prompt{}, fmt.Errorf("sqlite: create prompt %q: %w", input.Name, err)
	}
	return p, nil
}

func (s *MetaStore) GetPromptBySlug(ctx context.Context, slug string) (meta.Prompt, error) {
	var p meta.Prompt
	err := s.db.GetContext(ctx, &p, `
		SELECT id, name, slug, content, is_system, created_at, updated_at
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
		SELECT id, name, slug, content, is_system, created_at, updated_at
		FROM prompts WHERE is_system = 0 ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list prompts: %w", err)
	}
	return prompts, nil
}

func (s *MetaStore) ListSystemPrompts(ctx context.Context) ([]meta.Prompt, error) {
	var prompts []meta.Prompt
	err := s.db.SelectContext(ctx, &prompts, `
		SELECT id, name, slug, content, is_system, created_at, updated_at
		FROM prompts
		WHERE is_system = 1
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list system prompts: %w", err)
	}
	return prompts, nil
}

func (s *MetaStore) CountPrompts(ctx context.Context) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM prompts WHERE is_system = 0`)
	if err != nil {
		return 0, fmt.Errorf("sqlite: count prompts: %w", err)
	}
	return count, nil
}

// UpdatePrompt writes only the columns actually present in input — never a
// full-row rewrite of fields read moments earlier. With more than one SQLite
// connection now in flight (see sqlite.go), two concurrent partial updates
// to the same prompt (e.g. one changing Name, one changing Content) could
// otherwise race: both read the same starting row, and whichever writes
// second silently clobbers the other's change with its own stale copy of
// the field it never meant to touch.
func (s *MetaStore) UpdatePrompt(ctx context.Context, slug string, input meta.UpdatePromptInput) (meta.Prompt, error) {
	p, err := s.GetPromptBySlug(ctx, slug)
	if err != nil {
		return meta.Prompt{}, err
	}
	if p.IsSystem {
		return meta.Prompt{}, meta.ErrSystemReadOnly
	}

	var clauses []string
	var args []any
	if input.Name != nil {
		p.Name = *input.Name
		p.Slug = slugify(*input.Name)
		clauses = append(clauses, "name = ?", "slug = ?")
		args = append(args, p.Name, p.Slug)
	}
	if input.Content != nil {
		p.Content = *input.Content
		clauses = append(clauses, "content = ?")
		args = append(args, p.Content)
	}
	if len(clauses) == 0 {
		return p, nil
	}

	p.UpdatedAt = time.Now().UTC()
	clauses = append(clauses, "updated_at = ?")
	args = append(args, p.UpdatedAt, p.ID)

	_, err = s.db.ExecContext(ctx, "UPDATE prompts SET "+strings.Join(clauses, ", ")+" WHERE id = ?", args...)
	if err != nil {
		return meta.Prompt{}, fmt.Errorf("sqlite: update prompt %q: %w", slug, err)
	}
	return p, nil
}

func (s *MetaStore) upsertSystemPrompts(ctx context.Context) error {
	now := time.Now().UTC()
	for _, seed := range systemPrompts {
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO prompts (id, name, slug, content, is_system, created_at, updated_at)
			VALUES (?, ?, ?, ?, 1, ?, ?)
			ON CONFLICT(slug) DO UPDATE SET
				name       = excluded.name,
				content    = excluded.content,
				updated_at = excluded.updated_at
			WHERE is_system = 1
			  AND (prompts.name != excluded.name OR prompts.content != excluded.content)
		`, seed.id, seed.name, seed.slug, seed.content, now, now)
		if err != nil {
			return fmt.Errorf("upsert system prompt %q: %w", seed.slug, err)
		}
	}
	return nil
}

func (s *MetaStore) DeletePrompt(ctx context.Context, slug string) error {
	p, err := s.GetPromptBySlug(ctx, slug)
	if err != nil {
		return err
	}
	if p.IsSystem {
		return meta.ErrSystemReadOnly
	}

	_, err = s.db.ExecContext(ctx, `DELETE FROM prompts WHERE slug = ?`, slug)
	if err != nil {
		return fmt.Errorf("sqlite: delete prompt %q: %w", slug, err)
	}
	return nil
}
