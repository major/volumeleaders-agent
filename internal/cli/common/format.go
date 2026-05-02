package common

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// SelectJSONFields converts items to JSON maps and selects only the requested
// fields, preserving the requested field order for downstream output.
func SelectJSONFields[T any](items []T, fields []string) ([]map[string]json.RawMessage, error) {
	selected := make([]map[string]json.RawMessage, 0, len(items))
	for _, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			return nil, fmt.Errorf("marshal item for field selection: %w", err)
		}

		var allFields map[string]json.RawMessage
		if err := json.Unmarshal(data, &allFields); err != nil {
			return nil, fmt.Errorf("decode item for field selection: %w", err)
		}

		row := make(map[string]json.RawMessage, len(fields))
		for _, field := range fields {
			row[field] = allFields[field]
		}
		selected = append(selected, row)
	}
	return selected, nil
}

// ParseJSONFieldList parses a comma-separated JSON field list and validates it
// against the JSON tags of T.
func ParseJSONFieldList[T any](value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}

	validFields := JSONFieldNames[T]()
	valid := make(map[string]struct{}, len(validFields))
	for _, field := range validFields {
		valid[field] = struct{}{}
	}

	fields := make([]string, 0, strings.Count(value, ",")+1)
	seen := make(map[string]struct{})
	for field := range strings.SplitSeq(value, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if _, ok := valid[field]; !ok {
			return nil, fmt.Errorf("invalid field %q; valid fields: %s", field, strings.Join(validFields, ","))
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		fields = append(fields, field)
	}
	return fields, nil
}

// OutputFields returns defaultFields, every JSON field, or a parsed field list
// depending on the user's field selection value.
func OutputFields[T any](value string, defaultFields []string) ([]string, error) {
	if value == "" {
		return defaultFields, nil
	}
	if value == "all" {
		return JSONFieldNamesInOrder[T](), nil
	}
	return ParseJSONFieldList[T](value)
}

// JSONFieldNames returns sorted JSON tag names for T.
func JSONFieldNames[T any]() []string {
	fields := JSONFieldNamesInOrder[T]()
	slices.Sort(fields)
	return fields
}

// JSONFieldNamesInOrder returns JSON tag names for T in struct field order.
func JSONFieldNamesInOrder[T any]() []string {
	var zero T
	typeOf := reflect.TypeOf(zero)
	for typeOf.Kind() == reflect.Pointer {
		typeOf = typeOf.Elem()
	}

	fields := make([]string, 0, typeOf.NumField())
	for field := range typeOf.Fields() {
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		switch name {
		case "", "-":
			continue
		default:
			fields = append(fields, name)
		}
	}
	return fields
}

// ParseOutputFormat normalizes and validates a command output format value.
func ParseOutputFormat(value string) (OutputFormat, error) {
	format := OutputFormat(strings.ToLower(strings.TrimSpace(value)))
	if format == "" {
		return OutputFormatJSON, nil
	}

	switch format {
	case OutputFormatJSON, OutputFormatCSV, OutputFormatTSV:
		return format, nil
	default:
		return "", fmt.Errorf("invalid format %q; valid formats: json,csv,tsv", value)
	}
}
