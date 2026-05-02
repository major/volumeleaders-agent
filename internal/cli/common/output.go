package common

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

// PrintDataTablesResult writes DataTables-backed command results in the
// requested format, optionally narrowing rows to the requested JSON fields.
func PrintDataTablesResult[T any](w io.Writer, ctx context.Context, result []T, fields []string, format OutputFormat) error {
	if format == OutputFormatJSON && len(fields) == 0 {
		return PrintJSON(w, ctx, result)
	}

	selectedFields := fields
	if len(selectedFields) == 0 {
		selectedFields = JSONFieldNamesInOrder[T]()
	}

	selected, err := SelectJSONFields(result, selectedFields)
	if err != nil {
		return err
	}

	switch format {
	case OutputFormatJSON:
		return PrintJSON(w, ctx, selected)
	case OutputFormatCSV:
		return PrintDelimited(w, selected, selectedFields, ',')
	case OutputFormatTSV:
		return PrintDelimited(w, selected, selectedFields, '\t')
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

// PrintJSON marshals v to JSON and writes it to w. When the pretty flag is set
// in context, the output is indented for human readability.
func PrintJSON(w io.Writer, ctx context.Context, v any) error {
	var (
		output []byte
		err    error
	)
	if pretty, _ := ctx.Value(PrettyJSONKey).(bool); pretty {
		output, err = json.MarshalIndent(v, "", "  ")
	} else {
		output, err = json.Marshal(v)
	}
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	if _, err := w.Write(output); err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}
	if _, err := io.WriteString(w, "\n"); err != nil {
		return fmt.Errorf("write JSON newline: %w", err)
	}
	return nil
}

// PrintDelimited writes rows as delimited text with fields as the header row.
func PrintDelimited(w io.Writer, rows []map[string]json.RawMessage, fields []string, comma rune) error {
	writer := csv.NewWriter(w)
	writer.Comma = comma

	if err := writer.Write(fields); err != nil {
		return fmt.Errorf("write delimited header: %w", err)
	}
	for _, row := range rows {
		record, err := delimitedRecord(row, fields)
		if err != nil {
			return err
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("write delimited row: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("flush delimited output: %w", err)
	}
	return nil
}

func delimitedRecord(row map[string]json.RawMessage, fields []string) ([]string, error) {
	record := make([]string, 0, len(fields))
	for _, field := range fields {
		value, err := delimitedValue(row[field])
		if err != nil {
			return nil, fmt.Errorf("format field %q: %w", field, err)
		}
		record = append(record, value)
	}
	return record, nil
}

func delimitedValue(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "", nil
	}
	if isJSONNumber(raw) {
		return string(raw), nil
	}

	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil
	}

	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", fmt.Errorf("decode JSON value: %w", err)
	}
	switch typed := value.(type) {
	case bool:
		return strconv.FormatBool(typed), nil
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64), nil
	default:
		data, err := json.Marshal(typed)
		if err != nil {
			return "", fmt.Errorf("marshal complex JSON value: %w", err)
		}
		return string(data), nil
	}
}

func isJSONNumber(raw json.RawMessage) bool {
	return raw[0] == '-' || raw[0] >= '0' && raw[0] <= '9'
}
