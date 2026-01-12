// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package compep

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/OpenCHAMI/ochami/internal/cli"
)

func NewCmd() *cobra.Command {
	// compepCmd represents the "smd compep" command
	var compepCmd = &cobra.Command{
		Use:   "compep",
		Args:  cobra.NoArgs,
		Short: "Manage component endpoints",
		Long: `Manage component endpoints. This is a metacommand. Commands under this one
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
	compepCmd.AddCommand(
		newCmdCompepDelete(),
		newCmdCompepGet(),
	)

	return compepCmd
}
