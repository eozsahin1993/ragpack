package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"ragpack/pkg/meta"
)

func (s *MetaStore) CreateJob(ctx context.Context, collectionID, fileUri, mimeType string) (meta.Job, error) {
	now := time.Now().UTC()
	j := meta.Job{
		ID:           uuid.NewString(),
		CollectionID: collectionID,
		FileUri:      fileUri,
		MimeType:     mimeType,
		Status:       meta.JobStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO jobs (id, collection_id, file_uri, mime_type, status, executed_at, created_at, updated_at)
		VALUES (:id, :collection_id, :file_uri, :mime_type, :status, :executed_at, :created_at, :updated_at)
	`, j)
	if err != nil {
		return meta.Job{}, fmt.Errorf("sqlite: create job: %w", err)
	}
	return j, nil
}

func (s *MetaStore) GetJob(ctx context.Context, id string) (meta.Job, error) {
	var j meta.Job
	err := s.db.GetContext(ctx, &j, `
		SELECT id, collection_id, file_uri, mime_type, status, error, executed_at, created_at, updated_at
		FROM jobs
		WHERE id = ?
	`, id)
	if err != nil {
		return meta.Job{}, fmt.Errorf("sqlite: get job %q: %w", id, err)
	}
	return j, nil
}

func (s *MetaStore) ListAllJobs(ctx context.Context, limit, offset int) ([]meta.Job, error) {
	var jobs []meta.Job
	err := s.db.SelectContext(ctx, &jobs, `
		SELECT id, collection_id, file_uri, mime_type, status, error, executed_at, created_at, updated_at
		FROM jobs
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list all jobs: %w", err)
	}
	return jobs, nil
}

func (s *MetaStore) ListJobsByCollection(ctx context.Context, collectionID string, limit, offset int) ([]meta.Job, error) {
	var jobs []meta.Job
	err := s.db.SelectContext(ctx, &jobs, `
		SELECT id, collection_id, file_uri, mime_type, status, error, executed_at, created_at, updated_at
		FROM jobs
		WHERE collection_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, collectionID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list jobs for collection %q: %w", collectionID, err)
	}
	return jobs, nil
}

func (s *MetaStore) ListJobsByCollectionAndStatus(ctx context.Context, collectionID string, status meta.JobStatus, limit, offset int) ([]meta.Job, error) {
	var jobs []meta.Job
	err := s.db.SelectContext(ctx, &jobs, `
		SELECT id, collection_id, file_uri, mime_type, status, error, executed_at, created_at, updated_at
		FROM jobs
		WHERE collection_id = ? AND status = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, collectionID, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list jobs for collection %q with status %q: %w", collectionID, status, err)
	}
	return jobs, nil
}

func (s *MetaStore) ListJobsByStatus(ctx context.Context, status meta.JobStatus, limit, offset int) ([]meta.Job, error) {
	var jobs []meta.Job
	err := s.db.SelectContext(ctx, &jobs, `
		SELECT id, collection_id, file_uri, mime_type, status, error, executed_at, created_at, updated_at
		FROM jobs
		WHERE status = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list jobs by status %q: %w", status, err)
	}
	return jobs, nil
}

func (s *MetaStore) CountAllJobs(ctx context.Context) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM jobs`)
	if err != nil {
		return 0, fmt.Errorf("sqlite: count all jobs: %w", err)
	}
	return count, nil
}

func (s *MetaStore) CountJobsByCollection(ctx context.Context, collectionID string) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM jobs WHERE collection_id = ?`, collectionID)
	if err != nil {
		return 0, fmt.Errorf("sqlite: count jobs for collection %q: %w", collectionID, err)
	}
	return count, nil
}

func (s *MetaStore) CountJobsByCollectionAndStatus(ctx context.Context, collectionID string, status meta.JobStatus) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM jobs WHERE collection_id = ? AND status = ?`, collectionID, status)
	if err != nil {
		return 0, fmt.Errorf("sqlite: count jobs for collection %q with status %q: %w", collectionID, status, err)
	}
	return count, nil
}

func (s *MetaStore) CountJobsByStatus(ctx context.Context, status meta.JobStatus) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM jobs WHERE status = ?`, status)
	if err != nil {
		return 0, fmt.Errorf("sqlite: count jobs by status %q: %w", status, err)
	}
	return count, nil
}

func (s *MetaStore) UpdateJobStatus(ctx context.Context, id string, status meta.JobStatus, jobError *string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE jobs
		SET status = ?, error = ?, updated_at = ?
		WHERE id = ?
	`, status, jobError, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("sqlite: update job %q status: %w", id, err)
	}
	return nil
}
