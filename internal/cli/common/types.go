package common

import (
	"fmt"

	"github.com/leodido/structcli"
)

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
	structcli.RegisterEnum[TriStateFilter](map[TriStateFilter][]string{
		TriStateAll:     {"-1"},
		TriStateExclude: {"0"},
		TriStateOnly:    {"1"},
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

// Set implements pflag.Value for OutputFormat.
func (v *OutputFormat) Set(value string) error {
	switch OutputFormat(value) {
	case OutputFormatJSON, OutputFormatCSV, OutputFormatTSV:
		*v = OutputFormat(value)
		return nil
	default:
		return fmt.Errorf("invalid value %q, expected one of json, csv, tsv", value)
	}
}

// String implements pflag.Value for OutputFormat.
func (v OutputFormat) String() string {
	return string(v)
}

// Type implements pflag.Value for OutputFormat.
func (v OutputFormat) Type() string {
	return "string"
}

// OrderDirection identifies the supported DataTables sort directions.
type OrderDirection string

const (
	// OrderDirectionASC sorts results in ascending order.
	OrderDirectionASC OrderDirection = "asc"
	// OrderDirectionDESC sorts results in descending order.
	OrderDirectionDESC OrderDirection = "desc"
)

// Set implements pflag.Value for OrderDirection.
func (v *OrderDirection) Set(value string) error {
	switch OrderDirection(value) {
	case OrderDirectionASC, OrderDirectionDESC:
		*v = OrderDirection(value)
		return nil
	default:
		return fmt.Errorf("invalid value %q, expected one of asc, desc", value)
	}
}

// String implements pflag.Value for OrderDirection.
func (v OrderDirection) String() string {
	return string(v)
}

// Type implements pflag.Value for OrderDirection.
func (v OrderDirection) Type() string {
	return "string"
}

// TriStateFilter identifies toggle-style filters used by the VolumeLeaders API.
// A value of -1 leaves the filter unselected, 0 excludes matching rows, and 1
// returns only matching rows. It is string-backed so structcli can expose the
// supported values as JSON Schema enums while the CLI still accepts the same
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

// Set implements pflag.Value for TriStateFilter.
func (v *TriStateFilter) Set(value string) error {
	switch TriStateFilter(value) {
	case TriStateAll, TriStateExclude, TriStateOnly:
		*v = TriStateFilter(value)
		return nil
	default:
		return fmt.Errorf("invalid value %q, expected one of -1, 0, 1", value)
	}
}

// String implements pflag.Value for TriStateFilter.
func (v TriStateFilter) String() string {
	return string(v)
}

// Type implements pflag.Value for TriStateFilter.
func (v TriStateFilter) Type() string {
	return "string"
}

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

// DataTableRequestConfig names the request assembly fields that most commands
// pass through to DataTableOptions. Keeping the construction in one shape makes
// endpoint handlers easier to scan when they mainly differ by filters.
type DataTableRequestConfig struct {
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
