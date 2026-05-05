package common

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	flagGroupAnnotation    = "volumeleaders-agent/group"
	flagEnumAnnotation     = "volumeleaders-agent/enum"
	flagRequiredAnnotation = "volumeleaders-agent/required"
)

// ValidatableOptions describes option structs that perform post-parse checks.
type ValidatableOptions interface {
	Validate(context.Context) []error
}

// BindOrPanic binds CLI flags to opts using Cobra/pflag and panics if binding fails.
// This is expected to only fail during development when flag definitions are invalid.
func BindOrPanic(cmd *cobra.Command, opts any, name string) {
	if err := Bind(cmd, opts); err != nil {
		panic(fmt.Sprintf("bind %s: %v", name, err))
	}
}

// Bind registers pflag flags for every struct field tagged with flag metadata.
func Bind(cmd *cobra.Command, opts any) error {
	value := reflect.ValueOf(opts)
	if value.Kind() != reflect.Pointer || value.IsNil() || value.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("opts must be a non-nil pointer to a struct")
	}
	if err := bindStruct(cmd, value.Elem()); err != nil {
		return err
	}
	wrapValidation(cmd, opts)
	return nil
}

// FlagGroup returns the display group recorded for a flag.
func FlagGroup(flag *pflag.Flag) string {
	return firstAnnotation(flag, flagGroupAnnotation)
}

// FlagEnum returns the supported enum values recorded for a flag.
func FlagEnum(flag *pflag.Flag) []string {
	values := flag.Annotations[flagEnumAnnotation]
	return slices.Clone(values)
}

func bindStruct(cmd *cobra.Command, value reflect.Value) error {
	valueType := value.Type()
	for i := range value.NumField() {
		field := valueType.Field(i)
		fieldValue := value.Field(i)
		if field.Anonymous && fieldValue.Kind() == reflect.Struct {
			if err := bindStruct(cmd, fieldValue); err != nil {
				return err
			}
			continue
		}

		name := field.Tag.Get("flag")
		if name == "" {
			continue
		}
		if !fieldValue.CanSet() {
			return fmt.Errorf("field %s cannot be set", field.Name)
		}
		if cmd.Flags().Lookup(name) != nil {
			return fmt.Errorf("duplicate flag %q", name)
		}

		description := field.Tag.Get("flagdescr")
		shorthand := field.Tag.Get("flagshort")
		if defaultValue := field.Tag.Get("default"); defaultValue != "" {
			defaultSetter, err := newReflectFlagValue(fieldValue)
			if err != nil {
				return fmt.Errorf("flag %q default: %w", name, err)
			}
			if err := defaultSetter.Set(defaultValue); err != nil {
				return fmt.Errorf("flag %q default %q: %w", name, defaultValue, err)
			}
		}
		flagValue, err := newReflectFlagValue(fieldValue)
		if err != nil {
			return fmt.Errorf("flag %q: %w", name, err)
		}
		cmd.Flags().VarP(flagValue, name, shorthand, description)
		flag := cmd.Flags().Lookup(name)
		flag.DefValue = flagValue.String()
		if fieldValue.Kind() == reflect.Bool {
			flag.NoOptDefVal = "true"
		}
		annotateFlag(flag, flagGroupAnnotation, field.Tag.Get("flaggroup"))
		annotateFlag(flag, flagEnumAnnotation, enumValues(fieldValue.Type()))
		if field.Tag.Get("flagrequired") == "true" {
			annotateFlag(flag, flagRequiredAnnotation, "true")
		}
	}
	return nil
}

func annotateFlag(flag *pflag.Flag, key string, values any) {
	if flag.Annotations == nil {
		flag.Annotations = make(map[string][]string)
	}
	switch typed := values.(type) {
	case string:
		if typed != "" {
			flag.Annotations[key] = []string{typed}
		}
	case []string:
		if len(typed) > 0 {
			flag.Annotations[key] = slices.Clone(typed)
		}
	}
}

func wrapValidation(cmd *cobra.Command, opts any) {
	validator, ok := opts.(ValidatableOptions)
	original := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if original != nil {
			if err := original(cmd, args); err != nil {
				return err
			}
		}
		if schemaRequested(cmd) {
			return nil
		}
		if missing := missingRequiredFlags(cmd); len(missing) > 0 {
			return fmt.Errorf("required flag(s) not set: %s", strings.Join(missing, ", "))
		}
		if !ok {
			return nil
		}
		errs := validator.Validate(cmd.Context())
		if len(errs) == 0 {
			return nil
		}
		return errs[0]
	}
}

func schemaRequested(cmd *cobra.Command) bool {
	if flag := cmd.Flag("jsonschema"); flag != nil && flag.Changed {
		return true
	} else if flag != nil && flag.Value.String() != "" {
		return true
	}
	root := cmd.Root()
	if root == nil {
		return false
	}
	flag := root.PersistentFlags().Lookup("jsonschema")
	if flag != nil && flag.Changed {
		return true
	}
	if flag != nil && flag.Value.String() != "" {
		return true
	}
	return slices.ContainsFunc(os.Args, func(arg string) bool {
		return arg == "--jsonschema" || strings.HasPrefix(arg, "--jsonschema=")
	})
}

func missingRequiredFlags(cmd *cobra.Command) []string {
	missing := make([]string, 0)
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if IsFlagRequired(flag) && !flag.Changed {
			missing = append(missing, flag.Name)
		}
	})
	slices.Sort(missing)
	return missing
}

type reflectFlagValue struct {
	value reflect.Value
	enum  []string
}

func newReflectFlagValue(value reflect.Value) (*reflectFlagValue, error) {
	switch value.Kind() {
	case reflect.Bool, reflect.Float32, reflect.Float64, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.String:
		return &reflectFlagValue{value: value, enum: enumValues(value.Type())}, nil
	default:
		return nil, fmt.Errorf("unsupported flag field type %s", value.Type())
	}
}

func (flag *reflectFlagValue) Set(raw string) error {
	if len(flag.enum) > 0 && !slices.Contains(flag.enum, raw) {
		return fmt.Errorf("invalid value %q, expected one of %s", raw, strings.Join(flag.enum, ", "))
	}
	switch flag.value.Kind() {
	case reflect.Bool:
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		flag.value.SetBool(parsed)
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(raw, flag.value.Type().Bits())
		if err != nil {
			return err
		}
		flag.value.SetFloat(parsed)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(raw, 10, flag.value.Type().Bits())
		if err != nil {
			return err
		}
		flag.value.SetInt(parsed)
	case reflect.String:
		flag.value.SetString(raw)
	default:
		return fmt.Errorf("unsupported flag field type %s", flag.value.Type())
	}
	return nil
}

func (flag *reflectFlagValue) String() string {
	if !flag.value.IsValid() {
		return ""
	}
	switch flag.value.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(flag.value.Bool())
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(flag.value.Float(), 'f', -1, flag.value.Type().Bits())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(flag.value.Int(), 10)
	case reflect.String:
		return flag.value.String()
	default:
		return ""
	}
}

func (flag *reflectFlagValue) Type() string {
	switch flag.value.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Float32:
		return "float32"
	case reflect.Float64:
		return "float64"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int"
	default:
		return "string"
	}
}

func (flag *reflectFlagValue) IsBoolFlag() bool {
	return flag.value.Kind() == reflect.Bool
}

func enumValues(valueType reflect.Type) []string {
	key := valueType.PkgPath() + "." + valueType.Name()
	switch key {
	case "github.com/major/volumeleaders-agent/internal/cli/common.OutputFormat":
		return []string{"json", "csv", "tsv"}
	case "github.com/major/volumeleaders-agent/internal/cli/common.OrderDirection":
		return []string{"asc", "desc"}
	case "github.com/major/volumeleaders-agent/internal/cli/common.TriStateFilter":
		return []string{"-1", "0", "1"}
	case "github.com/major/volumeleaders-agent/internal/cli/alert.alertTickerGroup":
		return []string{"AllTickers", "SelectedTickers"}
	case "github.com/major/volumeleaders-agent/internal/cli/trade.tradeSummaryGroup", "github.com/major/volumeleaders-agent/internal/cli/report.reportSummaryGroup":
		return []string{"ticker", "day", "ticker,day", "ticker, day", "ticker day", "ticker-day"}
	case "github.com/major/volumeleaders-agent/internal/cli/watchlist.watchlistSecurityType":
		return []string{"-1", "1", "26", "4"}
	case "github.com/major/volumeleaders-agent/internal/cli/watchlist.watchlistRelativeSize":
		return []string{"0", "5", "10", "25", "50", "100"}
	case "github.com/major/volumeleaders-agent/internal/cli/watchlist.watchlistTradeRank":
		return []string{"-1", "1", "3", "5", "10", "25", "50", "100"}
	default:
		return nil
	}
}

func firstAnnotation(flag *pflag.Flag, key string) string {
	values := flag.Annotations[key]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// IsFlagRequired reports whether a flag was marked required by the binder.
func IsFlagRequired(flag *pflag.Flag) bool {
	return slices.Contains(flag.Annotations[flagRequiredAnnotation], "true")
}

// MarkFlagRequired records required-flag metadata without enabling Cobra's
// built-in required-flag validation, which would block schema generation.
func MarkFlagRequired(cmd *cobra.Command, name string) error {
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		return fmt.Errorf("unknown flag %q", name)
	}
	annotateFlag(flag, flagRequiredAnnotation, "true")
	return nil
}
