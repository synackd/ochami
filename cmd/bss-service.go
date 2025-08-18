// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// bssServiceCmd represents the "bss service" command
var bssServiceCmd = &cobra.Command{
	Use:   "service",
	Args:  cobra.NoArgs,
	Short: "Manage and check BSS itself",
	Long: `Manage and check BSS itself. This is a metacommand.

See ochami-bss(1) for more details.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			printUsageHandleError(cmd)
			os.Exit(0)
		}
	},
}

func init() {
	bssCmd.AddCommand(bssServiceCmd)
}
