package meta

import (
	"context"
	"time"
)

type Prompt struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	Slug      string    `db:"slug"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type CreatePromptInput struct {
	Name    string
	Content string
}

type UpdatePromptInput struct {
	Name    *string
	Content *string
}

type PromptReader interface {
	GetPromptBySlug(ctx context.Context, slug string) (Prompt, error)
	ListPrompts(ctx context.Context, limit, offset int) ([]Prompt, error)
	CountPrompts(ctx context.Context) (int, error)
}

type PromptWriter interface {
	CreatePrompt(ctx context.Context, input CreatePromptInput) (Prompt, error)
	UpdatePrompt(ctx context.Context, slug string, input UpdatePromptInput) (Prompt, error)
	DeletePrompt(ctx context.Context, slug string) error
}

type PromptStore interface {
	PromptReader
	PromptWriter
}
