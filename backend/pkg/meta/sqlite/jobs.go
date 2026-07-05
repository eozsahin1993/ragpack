package sqlite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"ragpack/pkg/meta"
	"ragpack/pkg/util"
)

func (s *MetaStore) CreateJob(ctx context.Context, collectionID, fileUri, mimeType string, intent meta.JobIntent, force bool) (meta.Job, error) {
	now := time.Now().UTC()
	displayName := util.NameFromURI(fileUri)
	j := meta.Job{
		ID:           uuid.NewString(),
		CollectionID: collectionID,
		FileUri:      fileUri,
		MimeType:     mimeType,
		DisplayName:  util.NonEmptyStr(displayName),
		Intent:       intent,
		Force:        force,
		Status:       meta.JobStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO jobs (id, collection_id, file_uri, mime_type, display_name, intent, force, status, executed_at, created_at, updated_at)
		VALUES (:id, :collection_id, :file_uri, :mime_type, :display_name, :intent, :force, :status, :executed_at, :created_at, :updated_at)
	`, j)
	if err != nil {
		return meta.Job{}, fmt.Errorf("sqlite: create job: %w", err)
	}
	return j, nil
}

func (s *MetaStore) GetJob(ctx context.Context, id string) (meta.Job, error) {
	var j meta.Job
	err := s.db.GetContext(ctx, &j, `SELECT * FROM jobs WHERE id = ?`, id)
	if err != nil {
		return meta.Job{}, fmt.Errorf("sqlite: get job %q: %w", id, err)
	}
	return j, nil
}

func (s *MetaStore) ListJobs(ctx context.Context, filter meta.JobFilter, limit, offset int) ([]meta.Job, error) {
	where, args := jobWhere(filter)
	q := "SELECT * FROM jobs" + where + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	var jobs []meta.Job
	if err := s.db.SelectContext(ctx, &jobs, q, args...); err != nil {
		return nil, fmt.Errorf("sqlite: list jobs: %w", err)
	}
	return jobs, nil
}

func (s *MetaStore) CountJobs(ctx context.Context, filter meta.JobFilter) (int, error) {
	where, args := jobWhere(filter)
	var count int
	if err := s.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM jobs"+where, args...); err != nil {
		return 0, fmt.Errorf("sqlite: count jobs: %w", err)
	}
	return count, nil
}

func jobWhere(f meta.JobFilter) (string, []any) {
	var clauses []string
	var args []any
	if f.CollectionID != nil {
		clauses = append(clauses, "collection_id = ?")
		args = append(args, *f.CollectionID)
	}
	if f.Status != nil {
		clauses = append(clauses, "status = ?")
		args = append(args, *f.Status)
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (s *MetaStore) UpdateJobStatus(ctx context.Context, id string, status meta.JobStatus, jobError *string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE jobs SET status = ?, error = ?, updated_at = ? WHERE id = ?`, status, jobError, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("sqlite: update job %q status: %w", id, err)
	}
	return nil
}

func (s *MetaStore) DeleteJob(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM jobs WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("sqlite: delete job %q: %w", id, err)
	}
	return nil
}
