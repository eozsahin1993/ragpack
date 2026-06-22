package meta

import (
	"context"
	"time"
)

type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusComplete   JobStatus = "complete"
	JobStatusFailed     JobStatus = "failed"
)

type Job struct {
	ID           string     `db:"id"`
	CollectionID string     `db:"collection_id"`
	FileUri      string     `db:"file_uri"`
	MimeType     string     `db:"mime_type"`
	Status       JobStatus  `db:"status"`
	Error        *string    `db:"error"`
	ExecutedAt   *time.Time `db:"executed_at"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

type JobReader interface {
	GetJob(ctx context.Context, id string) (Job, error)
	ListJobsByCollection(ctx context.Context, collectionID string) ([]Job, error)
	ListJobsByStatus(ctx context.Context, status JobStatus) ([]Job, error)
}

type JobWriter interface {
	CreateJob(ctx context.Context, collectionID, fileUri, mimeType string) (Job, error)
	UpdateJobStatus(ctx context.Context, id string, status JobStatus, jobError *string) error
}

type JobStore interface {
	JobReader
	JobWriter
}
