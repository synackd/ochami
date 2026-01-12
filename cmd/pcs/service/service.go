// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package service

import (
	"github.com/spf13/cobra"

	"github.com/OpenCHAMI/ochami/internal/cli"
)

func NewCmd() *cobra.Command {
	// serviceCmd represents the "pcs transitions" command
	var serviceCmd = &cobra.Command{
		Use:   "service",
		Args:  cobra.NoArgs,
		Short: "Manage and check PCS itself",
		Long: `Manage and check PCS itself. This is a metacommand.

See ochami-pcs(1) for more details.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cli.PrintUsageHandleError(cmd)
			}
		},
	}

	// Add subcommands
	serviceCmd.AddCommand(
		newCmdServiceStatus(),
	)

	return serviceCmd
}
