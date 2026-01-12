// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package service

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/OpenCHAMI/ochami/internal/cli"
)

func NewCmd() *cobra.Command {
	// serviceCmd represents the "bss service" command
	var serviceCmd = &cobra.Command{
		Use:   "service",
		Args:  cobra.NoArgs,
		Short: "Manage and check BSS itself",
		Long: `Manage and check BSS itself. This is a metacommand.

See ochami-bss(1) for more details.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cli.PrintUsageHandleError(cmd)
				os.Exit(0)
			}
		},
	}

	// Add subcommands
	serviceCmd.AddCommand(
		newCmdServiceStatus(),
		newCmdServiceVersion(),
	)

	return serviceCmd
}
