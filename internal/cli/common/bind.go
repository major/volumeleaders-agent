package common

import (
	"fmt"

	"github.com/leodido/structcli"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// BindOrPanic binds CLI flags to opts using structcli.Bind and panics if binding fails.
// This is expected to only fail during development (misconfigured flag definitions).
func BindOrPanic(cmd *cobra.Command, opts any, name string) {
	if err := structcli.Bind(cmd, opts); err != nil {
		panic(fmt.Sprintf("structcli.Bind %s: %v", name, err))
	}
}

// annotateFlag sets a single annotation value on a registered flag.
func annotateFlag(cmd *cobra.Command, flagName, key, value string) {
	_ = cmd.Flags().SetAnnotation(flagName, key, []string{value})
}

// firstAnnotation returns the first annotation value for key on f, or "" if absent.
func firstAnnotation(f *pflag.Flag, key string) string {
	if f.Annotations != nil {
		if vals, ok := f.Annotations[key]; ok && len(vals) > 0 {
			return vals[0]
		}
	}
	return ""
}
