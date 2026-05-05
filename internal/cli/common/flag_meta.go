package common

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Annotation keys for flag metadata stored via cobra flag annotations.
// These match the keys used by structcli so that structcli's JSONSchema and
// help-topic generators can read annotations set by our helpers during the
// migration period.
const (
	flagGroupAnnotation    = "leodido/structcli/flag-groups"
	flagEnumAnnotation     = "leodido/structcli/flag-enum"
	flagRequiredAnnotation = "cobra_annotation_bash_comp_one_required_flag"
)

// ValidatableOptions is implemented by option structs that require
// cross-field validation after flag parsing.
type ValidatableOptions interface {
	Validate(context.Context) []error
}

// FlagGroup returns the group annotation for f, or "" if none is set.
func FlagGroup(f *pflag.Flag) string {
	return firstAnnotation(f, flagGroupAnnotation)
}

// FlagEnum returns the comma-separated enum annotation for f, or "" if none is set.
func FlagEnum(f *pflag.Flag) string {
	return firstAnnotation(f, flagEnumAnnotation)
}

// IsFlagRequired reports whether f has been marked as required.
func IsFlagRequired(f *pflag.Flag) bool {
	return firstAnnotation(f, flagRequiredAnnotation) == "true"
}

// MarkFlagRequired marks a flag as required on cmd. The annotation is read by
// WrapValidation for enforcement and by cobra for shell completion hints.
func MarkFlagRequired(cmd *cobra.Command, flagName string) {
	_ = cmd.MarkFlagRequired(flagName)
}

// AnnotateFlagGroup sets the flag group annotation on a registered flag.
func AnnotateFlagGroup(cmd *cobra.Command, flagName, group string) {
	annotateFlag(cmd, flagName, flagGroupAnnotation, group)
}

// AnnotateFlagEnum sets the allowed values annotation on a registered flag.
func AnnotateFlagEnum(cmd *cobra.Command, flagName string, values []string) {
	_ = cmd.Flags().SetAnnotation(flagName, flagEnumAnnotation, values)
}

// WrapValidation installs a PreRunE hook on cmd that:
//  1. Skips all checks when --jsonschema is requested (schema bypass).
//  2. Reports missing required flags.
//  3. Calls opts.Validate() if opts implements ValidatableOptions.
//
// Any existing PreRunE is chained after validation passes.
func WrapValidation(cmd *cobra.Command, opts any) {
	existing := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if schemaRequested(cmd) {
			return nil
		}
		if err := missingRequiredFlags(cmd); err != nil {
			return err
		}
		if v, ok := opts.(ValidatableOptions); ok {
			for _, e := range v.Validate(cmd.Context()) {
				if e != nil {
					return e
				}
			}
		}
		if existing != nil {
			return existing(cmd, args)
		}
		return nil
	}
}

// schemaRequested reports whether --jsonschema was passed on the root command.
func schemaRequested(cmd *cobra.Command) bool {
	f := cmd.Root().Flags().Lookup("jsonschema")
	return f != nil && f.Changed
}

// missingRequiredFlags returns an error listing required flags that were not set.
// The error format matches cobra.Command.ValidateRequiredFlags for consistency.
func missingRequiredFlags(cmd *cobra.Command) error {
	var missing []string
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if IsFlagRequired(f) && !f.Changed {
			missing = append(missing, f.Name)
		}
	})
	if len(missing) > 0 {
		return fmt.Errorf(`required flag(s) "%s" not set`, strings.Join(missing, `", "`))
	}
	return nil
}
