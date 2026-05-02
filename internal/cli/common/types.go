package common

import "github.com/leodido/structcli"

func init() {
	structcli.RegisterEnum[OutputFormat](map[OutputFormat][]string{
		OutputFormatJSON: {"json"},
		OutputFormatCSV:  {"csv"},
		OutputFormatTSV:  {"tsv"},
	})
	structcli.RegisterEnum[OrderDirection](map[OrderDirection][]string{
		OrderDirectionASC:  {"asc"},
		OrderDirectionDESC: {"desc"},
	})
}

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

// OrderDirection identifies the supported DataTables sort directions.
type OrderDirection string

const (
	// OrderDirectionASC sorts results in ascending order.
	OrderDirectionASC OrderDirection = "asc"
	// OrderDirectionDESC sorts results in descending order.
	OrderDirectionDESC OrderDirection = "desc"
)

// DataTableOptions carries common DataTables request options.
type DataTableOptions struct {
	Start    int
	Length   int
	OrderCol int
	OrderDir OrderDirection
	Filters  map[string]string
	Fields   []string
}

// PaginationPageSize is the number of records fetched per page when the user
// requests all results (--length -1).
const PaginationPageSize = 1000
