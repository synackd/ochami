// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package node

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/OpenCHAMI/ochami/internal/cli"
)

func NewCmd() *cobra.Command {
	// nodeCmd represents the "cloud-init node" command
	var nodeCmd = &cobra.Command{
		Use:   "node",
		Args:  cobra.NoArgs,
		Short: "Manage cloud-init node-specific config",
		Long: `Manage cloud-init node-specific config.

See ochami-cloud-init(1) for more details.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cli.PrintUsageHandleError(cmd)
				os.Exit(0)
			}
		},
	}

	// Add subcommands
	nodeCmd.AddCommand(
		newCmdNodeGet(),
		newCmdNodeSet(),
	)

	return nodeCmd
}
