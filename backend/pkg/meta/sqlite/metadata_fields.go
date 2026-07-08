package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"ragpack/pkg/meta"
)

// maxSlotsPerType defines how many LanceDB slot columns exist per metadata field type.
var maxSlotsPerType = map[string]int{
	"str":  20,
	"num":  10,
	"bool": 10,
	"date": 10,
	"arr":  10,
}

func (s *MetaStore) RegisterMetadataFields(ctx context.Context, collectionID string, inputs []meta.MetadataFieldInput) ([]meta.MetadataField, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("sqlite: begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	result := make([]meta.MetadataField, 0, len(inputs))
	for _, input := range inputs {
		maxSlots, ok := maxSlotsPerType[input.Type]
		if !ok {
			return nil, fmt.Errorf("sqlite: unknown metadata field type %q", input.Type)
		}

		// Build CTE values for the slot range of this type.
		values := make([]string, maxSlots)
		for i := range maxSlots {
			values[i] = fmt.Sprintf("(%d)", i+1)
		}
		cte := fmt.Sprintf(`
			WITH slots(n) AS (VALUES %s)
			SELECT MIN(n) FROM slots
			WHERE n NOT IN (
				SELECT slot FROM collection_metadata_fields WHERE collection_id = ? AND type = ?
			)
		`, strings.Join(values, ","))

		var slot sql.NullInt64
		if err := tx.QueryRowContext(ctx, cte, collectionID, input.Type).Scan(&slot); err != nil || !slot.Valid {
			return nil, fmt.Errorf("sqlite: no available %s slots (max %d reached) for collection %s", input.Type, maxSlots, collectionID)
		}

		field := meta.MetadataField{
			ID:           uuid.New().String(),
			CollectionID: collectionID,
			Name:         input.Name,
			Type:         input.Type,
			Slot:         int(slot.Int64),
			CreatedAt:    time.Now().UTC(),
		}

		_, err = tx.NamedExecContext(ctx, `
			INSERT INTO collection_metadata_fields (id, collection_id, name, type, slot, created_at)
			VALUES (:id, :collection_id, :name, :type, :slot, :created_at)
		`, field)
		if err != nil {
			return nil, fmt.Errorf("sqlite: register metadata field %q: %w", input.Name, err)
		}

		result = append(result, field)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("sqlite: commit: %w", err)
	}
	return result, nil
}

func (s *MetaStore) ListMetadataFields(ctx context.Context, collectionID string) ([]meta.MetadataField, error) {
	var fields []meta.MetadataField
	err := s.db.SelectContext(ctx, &fields, `
		SELECT id, collection_id, name, type, slot, created_at
		FROM collection_metadata_fields
		WHERE collection_id = ?
		ORDER BY type, slot
	`, collectionID)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list metadata fields for collection %s: %w", collectionID, err)
	}
	return fields, nil
}

func (s *MetaStore) GetMetadataFieldByName(ctx context.Context, collectionID, name string) (meta.MetadataField, error) {
	var field meta.MetadataField
	err := s.db.GetContext(ctx, &field, `
		SELECT id, collection_id, name, type, slot, created_at
		FROM collection_metadata_fields
		WHERE collection_id = ? AND name = ?
	`, collectionID, name)
	if err != nil {
		return meta.MetadataField{}, fmt.Errorf("sqlite: get metadata field %q for collection %s: %w", name, collectionID, err)
	}
	return field, nil
}

func (s *MetaStore) DeleteMetadataField(ctx context.Context, collectionID, name string) (meta.MetadataField, error) {
	field, err := s.GetMetadataFieldByName(ctx, collectionID, name)
	if err != nil {
		return meta.MetadataField{}, err
	}

	_, err = s.db.ExecContext(ctx, `
		DELETE FROM collection_metadata_fields WHERE collection_id = ? AND name = ?
	`, collectionID, name)
	if err != nil {
		return meta.MetadataField{}, fmt.Errorf("sqlite: delete metadata field %q: %w", name, err)
	}
	return field, nil
}
