// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package iface

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/OpenCHAMI/ochami/internal/cli"
)

func NewCmd() *cobra.Command {
	// ifaceCmd represents the "smd iface" command
	var ifaceCmd = &cobra.Command{
		Use:   "iface",
		Args:  cobra.NoArgs,
		Short: "Manage ethernet interfaces",
		Long: `Manage ethernet interfaces. This is a metacommand. Commands under this one
interact with the State Management Database (SMD).

See ochami-smd(1) for more details.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cli.PrintUsageHandleError(cmd)
				os.Exit(0)
			}
		},
	}

	// Add subcommands
	ifaceCmd.AddCommand(
		newCmdIfaceAdd(),
		newCmdIfaceDelete(),
		newCmdIfaceGet(),
	)

	return ifaceCmd
}
