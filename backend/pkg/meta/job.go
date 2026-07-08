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

type JobIntent string

const (
	JobIntentIngest  JobIntent = "ingest"
	JobIntentRefresh JobIntent = "refresh"
)

type Job struct {
	ID           string     `db:"id"            json:"id"`
	CollectionID string     `db:"collection_id" json:"collection_id"`
	FileUri      string     `db:"file_uri"      json:"file_uri"`
	MimeType     string     `db:"mime_type"     json:"mime_type"`
	DisplayName  *string    `db:"display_name"  json:"display_name,omitempty"`
	ExtraJSON    *string    `db:"extra_json"    json:"extra_json,omitempty"`
	Metadata     *string    `db:"metadata"      json:"metadata,omitempty"` // JSON blob: user-supplied filterable metadata
	Intent       JobIntent  `db:"intent"        json:"intent"`
	Force        bool       `db:"force"         json:"force"`
	Status       JobStatus  `db:"status"        json:"status"`
	Error        *string    `db:"error"         json:"error,omitempty"`
	ExecutedAt   *time.Time `db:"executed_at"   json:"executed_at,omitempty"`
	CreatedAt    time.Time  `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"    json:"updated_at"`
}

// JobFilter holds optional predicates for listing/counting jobs.
// Nil fields are ignored (no filtering on that column).
type JobFilter struct {
	CollectionID *string
	Status       *JobStatus
}

type JobReader interface {
	GetJob(ctx context.Context, id string) (Job, error)
	ListJobs(ctx context.Context, filter JobFilter, limit, offset int) ([]Job, error)
	CountJobs(ctx context.Context, filter JobFilter) (int, error)
}

type JobWriter interface {
	CreateJob(ctx context.Context, collectionID, fileUri, mimeType string, intent JobIntent, force bool, extraJSON *string, metadata *string) (Job, error)
	UpdateJobStatus(ctx context.Context, id string, status JobStatus, jobError *string) error
	DeleteJob(ctx context.Context, id string) error
}

type JobStore interface {
	JobReader
	JobWriter
}
