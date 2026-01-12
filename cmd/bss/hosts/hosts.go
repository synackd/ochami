// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package hosts

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/OpenCHAMI/ochami/internal/cli"
)

func NewCmd() *cobra.Command {
	// hostsCmd represents the "bss hosts" command
	var hostsCmd = &cobra.Command{
		Use:   "hosts",
		Args:  cobra.NoArgs,
		Short: "Work with hosts in BSS",
		Long: `Work with hosts in BSS.

See ochami-bss(1) for more details.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cli.PrintUsageHandleError(cmd)
				os.Exit(0)
			}
		},
	}

	// Add subcommands
	hostsCmd.AddCommand(
		newCmdHostsGet(),
	)

	return hostsCmd
}
