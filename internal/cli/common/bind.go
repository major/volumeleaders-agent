package common

import (
	"fmt"

	"github.com/leodido/structcli"
	"github.com/spf13/cobra"
)

// BindOrPanic binds CLI flags to opts using structcli.Bind and panics if binding fails.
// This is expected to only fail during development (misconfigured flag definitions).
func BindOrPanic(cmd *cobra.Command, opts any, name string) {
	if err := structcli.Bind(cmd, opts); err != nil {
		panic(fmt.Sprintf("structcli.Bind %s: %v", name, err))
	}
}
