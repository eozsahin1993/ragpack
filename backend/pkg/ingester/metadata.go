package ingester

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/araddon/dateparse"

	"ragpack/pkg/db"
	"ragpack/pkg/meta"
)

// metadataSlots holds one document's typed metadata-field values, resolved
// once and reused as-is on every chunk record for that document.
type metadataSlots struct {
	str     [20]*string
	num     [10]*float64
	boolean [10]*bool
	date    [10]*int64
	arr     [10][]string
}

// resolveMetadataSlots looks up the collection's metadata field mapping once
// (avoids redundant SQLite reads per batch) and routes job.Metadata into it.
func (wp *WorkerPool) resolveMetadataSlots(ctx context.Context, job meta.Job, collection meta.Collection) (metadataSlots, error) {
	var slots metadataSlots
	if job.Metadata == nil {
		return slots, nil
	}
	metaFields, err := wp.metaStore.ListMetadataFields(ctx, collection.ID)
	if err != nil {
		return slots, fmt.Errorf("list metadata fields: %w", err)
	}
	if len(metaFields) == 0 {
		return slots, nil
	}
	fieldMap := make(map[string]meta.MetadataField, len(metaFields))
	for _, f := range metaFields {
		fieldMap[f.Name] = f
	}
	var rawMeta map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(*job.Metadata), &rawMeta); jsonErr == nil {
		slots.str, slots.num, slots.boolean, slots.date, slots.arr = routeMetadataSlots(rawMeta, fieldMap, job.ID)
	}
	return slots, nil
}

// MergeMetadataSlots populates the typed slot maps in patch from user-supplied metadata.
// Unregistered fields and type mismatches are silently skipped.
func MergeMetadataSlots(patch db.ChunkPatch, raw map[string]any, fields []meta.MetadataField) db.ChunkPatch {
	fieldMap := make(map[string]meta.MetadataField, len(fields))
	for _, f := range fields {
		fieldMap[f.Name] = f
	}

	patch.MetaStr = make(map[int]*string)
	patch.MetaNum = make(map[int]*float64)
	patch.MetaBool = make(map[int]*bool)
	patch.MetaDate = make(map[int]*int64)
	patch.MetaArr = make(map[int][]string)

	for key, val := range raw {
		field, ok := fieldMap[key]
		if !ok {
			continue
		}
		switch field.Type {
		case "str":
			if s, ok := val.(string); ok {
				v := s
				patch.MetaStr[field.Slot] = &v
			}
		case "num":
			if n, ok := val.(float64); ok {
				v := n
				patch.MetaNum[field.Slot] = &v
			}
		case "bool":
			if b, ok := val.(bool); ok {
				v := b
				patch.MetaBool[field.Slot] = &v
			}
		case "date":
			switch v := val.(type) {
			case float64:
				t := int64(v)
				patch.MetaDate[field.Slot] = &t
			case string:
				if t, err := dateparse.ParseAny(v); err == nil {
					u := t.UTC().Unix()
					patch.MetaDate[field.Slot] = &u
				}
			}
		case "arr":
			if arr, ok := val.([]any); ok {
				strs := make([]string, 0, len(arr))
				for _, item := range arr {
					if s, ok := item.(string); ok {
						strs = append(strs, s)
					}
				}
				patch.MetaArr[field.Slot] = strs
			}
		}
	}
	return patch
}

// routeMetadataSlots maps user-supplied metadata values to their pre-declared Arrow slot arrays.
// Undeclared keys are logged as warnings. Type coercion failures are also logged and skipped.
func routeMetadataSlots(raw map[string]interface{}, fieldMap map[string]meta.MetadataField, jobID string) (
	metaStr [20]*string, metaNum [10]*float64, metaBool [10]*bool, metaDate [10]*int64, metaArr [10][]string,
) {
	for key, val := range raw {
		mfield, ok := fieldMap[key]
		if !ok {
			log.Printf("ingester: job %s: metadata field %q is not registered on this collection — skipped", jobID, key)
			continue
		}
		idx := mfield.Slot - 1
		switch mfield.Type {
		case "str":
			str, ok := coerceToString(val)
			if !ok {
				log.Printf("ingester: job %s: metadata field %q: cannot coerce %T to string — skipped", jobID, key, val)
				continue
			}
			metaStr[idx] = &str
		case "num":
			num, ok := coerceToFloat64(val)
			if !ok {
				log.Printf("ingester: job %s: metadata field %q: cannot coerce %T to number — skipped", jobID, key, val)
				continue
			}
			metaNum[idx] = &num
		case "bool":
			b, ok := coerceToBool(val)
			if !ok {
				log.Printf("ingester: job %s: metadata field %q: cannot coerce %T to bool — skipped", jobID, key, val)
				continue
			}
			metaBool[idx] = &b
		case "date":
			ts, ok := coerceToUnixTimestamp(val)
			if !ok {
				log.Printf("ingester: job %s: metadata field %q: cannot parse %T as date — skipped", jobID, key, val)
				continue
			}
			metaDate[idx] = &ts
		case "arr":
			arr, ok := coerceToStringSlice(val)
			if !ok {
				log.Printf("ingester: job %s: metadata field %q: expected array of strings, got %T — skipped", jobID, key, val)
				continue
			}
			metaArr[idx] = arr
		}
	}
	return
}

func coerceToString(val interface{}) (string, bool) {
	switch v := val.(type) {
	case string:
		return v, true
	case bool:
		if v {
			return "true", true
		}
		return "false", true
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), true
	default:
		return "", false
	}
}

func coerceToFloat64(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case string:
		num, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, false
		}
		return num, true
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

func coerceToStringSlice(val interface{}) ([]string, bool) {
	arr, ok := val.([]interface{})
	if !ok {
		return nil, false
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		str, ok := item.(string)
		if !ok {
			return nil, false
		}
		out = append(out, str)
	}
	return out, true
}

func coerceToBool(val interface{}) (bool, bool) {
	switch v := val.(type) {
	case bool:
		return v, true
	case string:
		switch strings.ToLower(v) {
		case "true", "1", "yes":
			return true, true
		case "false", "0", "no":
			return false, true
		}
	case float64:
		return v != 0, true
	}
	return false, false
}

func coerceToUnixTimestamp(val interface{}) (int64, bool) {
	switch v := val.(type) {
	case float64:
		return int64(v), true
	case string:
		t, err := dateparse.ParseAny(v)
		if err != nil {
			return 0, false
		}
		return t.UTC().Unix(), true
	}
	return 0, false
}
