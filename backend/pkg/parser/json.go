package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"strconv"
	"strings"
)

// JSONParser streams one Unit per element for JSON arrays, or one Unit for a
// root object. Each value is flattened to "key: value" text with dot-notation
// for nested objects and index-notation for arrays.
type JSONParser struct{}

func (p *JSONParser) Parse(_ context.Context, r io.ReadCloser) iter.Seq2[Unit, error] {
	return func(yield func(Unit, error) bool) {
		defer r.Close()

		raw, err := io.ReadAll(r)
		if err != nil {
			yield(Unit{}, fmt.Errorf("json: read: %w", err))
			return
		}

		var root interface{}
		if err := json.Unmarshal(raw, &root); err != nil {
			yield(Unit{}, fmt.Errorf("json: parse: %w", err))
			return
		}

		switch v := root.(type) {
		case []interface{}:
			for _, element := range v {
				text := strings.TrimSpace(flattenJSON("", element))
				if text == "" {
					continue
				}
				if !yield(Unit{Kind: UnitKindRow, Text: text, Metadata: map[string]string{}}, nil) {
					return
				}
			}
		default:
			text := strings.TrimSpace(flattenJSON("", v))
			if text != "" {
				yield(Unit{Kind: UnitKindRow, Text: text, Metadata: map[string]string{}}, nil)
			}
		}
	}
}

// flattenJSON recursively converts a JSON value into "key: value\n" lines
// using dot-notation for nested objects and index-notation for arrays.
func flattenJSON(prefix string, v interface{}) string {
	var sb strings.Builder
	flattenJSONInto(&sb, prefix, v)
	return sb.String()
}

func flattenJSONInto(sb *strings.Builder, prefix string, v interface{}) {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, child := range val {
			key := k
			if prefix != "" {
				key = prefix + "." + k
			}
			flattenJSONInto(sb, key, child)
		}
	case []interface{}:
		for i, child := range val {
			key := strconv.Itoa(i)
			if prefix != "" {
				key = prefix + "." + key
			}
			flattenJSONInto(sb, key, child)
		}
	case string:
		if val != "" {
			sb.WriteString(prefix)
			sb.WriteString(": ")
			sb.WriteString(val)
			sb.WriteByte('\n')
		}
	case float64:
		sb.WriteString(prefix)
		sb.WriteString(": ")
		sb.WriteString(strconv.FormatFloat(val, 'f', -1, 64))
		sb.WriteByte('\n')
	case bool:
		sb.WriteString(prefix)
		sb.WriteString(": ")
		if val {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
		sb.WriteByte('\n')
	case nil:
		// skip null values
	}
}
