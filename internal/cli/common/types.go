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

// OrderDirection identifies the supported DataTables sort directions.
type OrderDirection string

const (
	// OrderDirectionASC sorts results in ascending order.
	OrderDirectionASC OrderDirection = "asc"
	// OrderDirectionDESC sorts results in descending order.
	OrderDirectionDESC OrderDirection = "desc"
)

// TriStateFilter identifies toggle-style filters used by the VolumeLeaders API.
// A value of -1 leaves the filter unselected, 0 excludes matching rows, and 1
// returns only matching rows. It is string-backed so the CLI accepts the same
// numeric-looking arguments users pass today.
type TriStateFilter string

const (
	// TriStateAll leaves the filter unselected.
	TriStateAll TriStateFilter = "-1"
	// TriStateExclude excludes rows matching the filter.
	TriStateExclude TriStateFilter = "0"
	// TriStateOnly returns only rows matching the filter.
	TriStateOnly TriStateFilter = "1"
)

// Int returns the integer value expected by internal DataTables filter code.
func (value TriStateFilter) Int() int {
	switch value {
	case TriStateOnly:
		return 1
	case TriStateExclude:
		return 0
	default:
		return -1
	}
}

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
