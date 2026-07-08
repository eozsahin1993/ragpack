package filter

import (
	"fmt"
	"strings"
)

func compileNode(node interface{}, fieldMap FieldMap) (string, error) {
	obj, ok := node.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("filter: expected object, got %T", node)
	}

	parts := make([]string, 0, len(obj))
	for key, val := range obj {
		switch key {
		case "$and":
			clause, err := compileLogical("AND", val, fieldMap)
			if err != nil {
				return "", err
			}
			parts = append(parts, clause)
		case "$or":
			clause, err := compileLogical("OR", val, fieldMap)
			if err != nil {
				return "", err
			}
			parts = append(parts, clause)
		default:
			clause, err := compileFieldExpr(key, val, fieldMap)
			if err != nil {
				return "", err
			}
			parts = append(parts, clause)
		}
	}

	if len(parts) == 1 {
		return parts[0], nil
	}
	return "(" + strings.Join(parts, " AND ") + ")", nil
}

func compileLogical(op string, val interface{}, fieldMap FieldMap) (string, error) {
	arr, ok := val.([]interface{})
	if !ok {
		return "", fmt.Errorf("filter: $and/$or requires an array, got %T", val)
	}
	parts := make([]string, 0, len(arr))
	for _, elem := range arr {
		clause, err := compileNode(elem, fieldMap)
		if err != nil {
			return "", err
		}
		parts = append(parts, clause)
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("filter: empty $and/$or array")
	}
	return "(" + strings.Join(parts, " "+op+" ") + ")", nil
}

func compileFieldExpr(fieldName string, val interface{}, fieldMap FieldMap) (string, error) {
	col, fieldType, err := resolveColumn(fieldName, fieldMap)
	if err != nil {
		return "", err
	}

	ops, ok := val.(map[string]interface{})
	if !ok {
		return buildEq(col, val, fieldType)
	}

	parts := make([]string, 0, len(ops))
	for op, opVal := range ops {
		clause, err := buildOp(col, op, opVal, fieldType)
		if err != nil {
			return "", err
		}
		parts = append(parts, clause)
	}
	if len(parts) == 1 {
		return parts[0], nil
	}
	return "(" + strings.Join(parts, " AND ") + ")", nil
}

func buildOp(col, op string, val interface{}, fieldType string) (string, error) {
	switch op {
	case "$eq":
		return buildEq(col, val, fieldType)
	case "$ne":
		literal, err := formatValue(val, fieldType)
		if err != nil {
			return "", err
		}
		return col + " != " + literal, nil
	case "$gt":
		literal, err := formatValue(val, fieldType)
		if err != nil {
			return "", err
		}
		return col + " > " + literal, nil
	case "$gte":
		literal, err := formatValue(val, fieldType)
		if err != nil {
			return "", err
		}
		return col + " >= " + literal, nil
	case "$lt":
		literal, err := formatValue(val, fieldType)
		if err != nil {
			return "", err
		}
		return col + " < " + literal, nil
	case "$lte":
		literal, err := formatValue(val, fieldType)
		if err != nil {
			return "", err
		}
		return col + " <= " + literal, nil
	case "$in":
		return buildList(col, val, fieldType, "IN")
	case "$nin":
		return buildList(col, val, fieldType, "NOT IN")
	case "$exists":
		exists, ok := val.(bool)
		if !ok {
			return "", fmt.Errorf("filter: $exists requires a boolean")
		}
		if exists {
			return col + " IS NOT NULL", nil
		}
		return col + " IS NULL", nil
	case "$like":
		if fieldType != "str" && fieldType != "timestamp" {
			return "", fmt.Errorf("filter: $like is only valid on str fields")
		}
		pattern, err := requireString(val, "$like")
		if err != nil {
			return "", err
		}
		return col + " LIKE " + quoteSQLString(pattern), nil
	case "$ilike":
		if fieldType != "str" {
			return "", fmt.Errorf("filter: $ilike is only valid on str fields")
		}
		pattern, err := requireString(val, "$ilike")
		if err != nil {
			return "", err
		}
		return col + " ILIKE " + quoteSQLString(pattern), nil
	case "$contains":
		if fieldType != "arr" {
			return "", fmt.Errorf("filter: $contains is only valid on arr fields")
		}
		item, err := requireString(val, "$contains")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("array_has(%s, %s)", col, quoteSQLString(item)), nil
	case "$containsAny":
		if fieldType != "arr" {
			return "", fmt.Errorf("filter: $containsAny is only valid on arr fields")
		}
		return buildArrayHas("array_has_any", col, val)
	case "$containsAll":
		if fieldType != "arr" {
			return "", fmt.Errorf("filter: $containsAll is only valid on arr fields")
		}
		return buildArrayHas("array_has_all", col, val)
	default:
		return "", fmt.Errorf("filter: unknown operator %q", op)
	}
}

func buildEq(col string, val interface{}, fieldType string) (string, error) {
	literal, err := formatValue(val, fieldType)
	if err != nil {
		return "", err
	}
	return col + " = " + literal, nil
}

func buildList(col string, val interface{}, fieldType, sqlOp string) (string, error) {
	arr, ok := val.([]interface{})
	if !ok {
		return "", fmt.Errorf("filter: $in/$nin requires an array")
	}
	parts := make([]string, 0, len(arr))
	for _, item := range arr {
		literal, err := formatValue(item, fieldType)
		if err != nil {
			return "", err
		}
		parts = append(parts, literal)
	}
	return col + " " + sqlOp + " (" + strings.Join(parts, ", ") + ")", nil
}

func buildArrayHas(fn, col string, val interface{}) (string, error) {
	arr, ok := val.([]interface{})
	if !ok {
		return "", fmt.Errorf("filter: %s requires an array value", fn)
	}
	parts := make([]string, 0, len(arr))
	for _, item := range arr {
		str, err := requireString(item, fn)
		if err != nil {
			return "", err
		}
		parts = append(parts, quoteSQLString(str))
	}
	return fmt.Sprintf("%s(%s, make_array(%s))", fn, col, strings.Join(parts, ", ")), nil
}
