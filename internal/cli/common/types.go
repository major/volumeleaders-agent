package common

// ContextKey identifies values stored in command contexts.
type ContextKey int

const (
	// PrettyJSONKey enables indented JSON output when set to true in context.
	PrettyJSONKey ContextKey = 1
	// TestClientKey identifies an injected client value used by command tests.
	TestClientKey ContextKey = 2
)

// OutputFormat identifies the supported command output formats.
type OutputFormat string

const (
	// OutputFormatJSON emits JSON output.
	OutputFormatJSON OutputFormat = "json"
	// OutputFormatCSV emits comma-separated output.
	OutputFormatCSV OutputFormat = "csv"
	// OutputFormatTSV emits tab-separated output.
	OutputFormatTSV OutputFormat = "tsv"
)

// DataTableOptions carries common DataTables request options.
type DataTableOptions struct {
	Start    int
	Length   int
	OrderCol int
	OrderDir string
	Filters  map[string]string
	Fields   []string
}

// PaginationPageSize is the number of records fetched per page when the user
// requests all results (--length -1).
const PaginationPageSize = 1000
