// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package discover

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/OpenCHAMI/ochami/internal/cli"

	// Subcomands
	static_cmd "github.com/OpenCHAMI/ochami/cmd/discover/static"
)

func NewCmd() *cobra.Command {
	// discoverCmd represents the discover command
	var discoverCmd = &cobra.Command{
		Use:   "discover",
		Args:  cobra.NoArgs,
		Short: "Perform static or dynamic discovery of nodes",
		Run: func(cmd *cobra.Command, args []string) {
			// Check that all required args are passed
			if len(args) == 0 {
				cli.PrintUsageHandleError(cmd)
				os.Exit(0)
			}
		},
	}

	// Add subcommands
	discoverCmd.AddCommand(
		static_cmd.NewCmd(),
	)

	return discoverCmd
}
