package meta

import (
	"context"
	"errors"
	"time"
)

var ErrSystemReadOnly = errors.New("system prompts cannot be modified or deleted")

type Prompt struct {
	ID        string    `db:"id"         json:"id"`
	Name      string    `db:"name"       json:"name"`
	Slug      string    `db:"slug"       json:"slug"`
	Content   string    `db:"content"    json:"content"`
	IsSystem  bool      `db:"is_system"  json:"is_system"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
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
	ListSystemPrompts(ctx context.Context) ([]Prompt, error)
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
