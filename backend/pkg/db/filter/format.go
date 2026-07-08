package filter

import (
	"fmt"
	"strings"
	"time"

	"github.com/araddon/dateparse"
)

// formatValue converts a Go value to a DataFusion SQL literal for the given field type.
func formatValue(val interface{}, fieldType string) (string, error) {
	if val == nil {
		return "NULL", nil
	}
	switch v := val.(type) {
	case bool:
		if fieldType == "bool" {
			if v {
				return "true", nil
			}
			return "false", nil
		}
		if v {
			return "'true'", nil
		}
		return "'false'", nil
	case float64:
		switch fieldType {
		case "str":
			return quoteSQLString(fmt.Sprintf("%g", v)), nil
		case "date":
			return fmt.Sprintf("%d", int64(v)), nil
		default:
			return fmt.Sprintf("%g", v), nil
		}
	case string:
		switch fieldType {
		case "timestamp", "date":
			return formatTimestamp(v)
		case "bool":
			switch strings.ToLower(v) {
			case "true", "1", "yes":
				return "true", nil
			case "false", "0", "no":
				return "false", nil
			default:
				return "", fmt.Errorf("filter: cannot parse %q as bool", v)
			}
		default:
			return quoteSQLString(v), nil
		}
	default:
		return "", fmt.Errorf("filter: unsupported value type %T", val)
	}
}

// formatTimestamp converts an ISO date string or relative expression to a unix int64 SQL literal.
func formatTimestamp(s string) (string, error) {
	parsed, err := dateparse.ParseAny(s)
	if err != nil {
		parsed, err = parseRelativeDate(s)
		if err != nil {
			return "", fmt.Errorf("filter: cannot parse date/time %q: %w", s, err)
		}
	}
	return fmt.Sprintf("%d", parsed.UTC().Unix()), nil
}

// parseRelativeDate handles common relative date strings like "last week", "7 days ago".
func parseRelativeDate(s string) (time.Time, error) {
	now := time.Now().UTC()
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "today":
		return truncateDay(now), nil
	case "yesterday":
		return truncateDay(now.AddDate(0, 0, -1)), nil
	case "last week":
		return truncateDay(now.AddDate(0, 0, -7)), nil
	case "last month":
		return truncateDay(now.AddDate(0, -1, 0)), nil
	case "last year":
		return truncateDay(now.AddDate(-1, 0, 0)), nil
	}
	var count int
	var unit string
	if _, err := fmt.Sscanf(s, "%d %s ago", &count, &unit); err == nil {
		unit = strings.TrimSuffix(unit, "s")
		switch unit {
		case "day":
			return truncateDay(now.AddDate(0, 0, -count)), nil
		case "week":
			return truncateDay(now.AddDate(0, 0, -count*7)), nil
		case "month":
			return truncateDay(now.AddDate(0, -count, 0)), nil
		case "year":
			return truncateDay(now.AddDate(-count, 0, 0)), nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized relative date: %q", s)
}

func truncateDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func requireString(val interface{}, op string) (string, error) {
	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("filter: %s requires a string value, got %T", op, val)
	}
	return str, nil
}

// quoteSQLString escapes a string for safe embedding in a DataFusion SQL predicate.
// Single quotes are doubled per SQL standard.
func quoteSQLString(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
