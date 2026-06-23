package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"ragpack/pkg/meta"
)

func (s *MetaStore) CreateDocument(ctx context.Context, collectionID, jobID, fileUri, mimeType string) (meta.Document, error) {
	now := time.Now().UTC()
	d := meta.Document{
		ID:           uuid.NewString(),
		CollectionID: collectionID,
		JobID:        jobID,
		FileUri:      fileUri,
		MimeType:     mimeType,
		ChunkCount:   0,
		Status:       meta.DocumentStatusIngesting,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// ON CONFLICT handles the requeue case: a job that was processing when the
	// server crashed is re-run from scratch, so we reset its document back to ingesting.
	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO documents (id, collection_id, job_id, file_uri, mime_type, external_id, extra_json, chunk_count, status, error, created_at, updated_at)
		VALUES (:id, :collection_id, :job_id, :file_uri, :mime_type, :external_id, :extra_json, :chunk_count, :status, :error, :created_at, :updated_at)
		ON CONFLICT(job_id) DO UPDATE SET
			status      = 'ingesting',
			chunk_count = 0,
			error       = NULL,
			updated_at  = excluded.updated_at
	`, d)
	if err != nil {
		return meta.Document{}, fmt.Errorf("sqlite: create document: %w", err)
	}

	// Re-fetch to get the canonical row (may have been the existing record).
	var existing meta.Document
	if err := s.db.GetContext(ctx, &existing, `
		SELECT id, collection_id, job_id, file_uri, mime_type, external_id, extra_json, chunk_count, status, error, created_at, updated_at
		FROM documents WHERE job_id = ?
	`, jobID); err != nil {
		return meta.Document{}, fmt.Errorf("sqlite: fetch document after upsert: %w", err)
	}
	return existing, nil
}

func (s *MetaStore) GetDocument(ctx context.Context, id string) (meta.Document, error) {
	var d meta.Document
	err := s.db.GetContext(ctx, &d, `
		SELECT id, collection_id, job_id, file_uri, mime_type, external_id, extra_json, chunk_count, status, error, created_at, updated_at
		FROM documents
		WHERE id = ?
	`, id)
	if err != nil {
		return meta.Document{}, fmt.Errorf("sqlite: get document %q: %w", id, err)
	}
	return d, nil
}

func (s *MetaStore) ListDocumentsByCollection(ctx context.Context, collectionID string, limit, offset int) ([]meta.Document, error) {
	var docs []meta.Document
	err := s.db.SelectContext(ctx, &docs, `
		SELECT id, collection_id, job_id, file_uri, mime_type, external_id, extra_json, chunk_count, status, error, created_at, updated_at
		FROM documents
		WHERE collection_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, collectionID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list documents for collection %q: %w", collectionID, err)
	}
	return docs, nil
}

func (s *MetaStore) CountDocumentsByCollection(ctx context.Context, collectionID string) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM documents WHERE collection_id = ?
	`, collectionID)
	if err != nil {
		return 0, fmt.Errorf("sqlite: count documents for collection %q: %w", collectionID, err)
	}
	return count, nil
}

func (s *MetaStore) UpdateDocumentStatus(ctx context.Context, id string, status meta.DocumentStatus, chunkCount int, docError *string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE documents
		SET status = ?, chunk_count = ?, error = ?, updated_at = ?
		WHERE id = ?
	`, status, chunkCount, docError, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("sqlite: update document %q status: %w", id, err)
	}
	return nil
}

func (s *MetaStore) DeleteDocumentsByCollection(ctx context.Context, collectionID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM documents WHERE collection_id = ?`, collectionID)
	if err != nil {
		return fmt.Errorf("sqlite: delete documents for collection %q: %w", collectionID, err)
	}
	return nil
}
