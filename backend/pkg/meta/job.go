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
	ID           string     `db:"id"            json:"id"`
	CollectionID string     `db:"collection_id" json:"collection_id"`
	FileUri      string     `db:"file_uri"      json:"file_uri"`
	MimeType     string     `db:"mime_type"     json:"mime_type"`
	Status       JobStatus  `db:"status"        json:"status"`
	Error        *string    `db:"error"         json:"error,omitempty"`
	ExecutedAt   *time.Time `db:"executed_at"   json:"executed_at,omitempty"`
	CreatedAt    time.Time  `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"    json:"updated_at"`
}

type JobReader interface {
	GetJob(ctx context.Context, id string) (Job, error)
	ListAllJobs(ctx context.Context, limit, offset int) ([]Job, error)
	ListJobsByCollection(ctx context.Context, collectionID string, limit, offset int) ([]Job, error)
	ListJobsByCollectionAndStatus(ctx context.Context, collectionID string, status JobStatus, limit, offset int) ([]Job, error)
	ListJobsByStatus(ctx context.Context, status JobStatus, limit, offset int) ([]Job, error)
	CountAllJobs(ctx context.Context) (int, error)
	CountJobsByCollection(ctx context.Context, collectionID string) (int, error)
	CountJobsByCollectionAndStatus(ctx context.Context, collectionID string, status JobStatus) (int, error)
	CountJobsByStatus(ctx context.Context, status JobStatus) (int, error)
}

type JobWriter interface {
	CreateJob(ctx context.Context, collectionID, fileUri, mimeType string) (Job, error)
	UpdateJobStatus(ctx context.Context, id string, status JobStatus, jobError *string) error
}

type JobStore interface {
	JobReader
	JobWriter
}
