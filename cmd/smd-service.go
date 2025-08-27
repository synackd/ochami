// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// smdServiceCmd represents the "smd service" command
var smdServiceCmd = &cobra.Command{
	Use:   "service",
	Args:  cobra.NoArgs,
	Short: "Check/Manage the State Management Database (SMD)",
	Long: `Check/Manage the State Management Database (SMD). This is a metacommand.

See ochami-smd(1) for more details.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			printUsageHandleError(cmd)
			os.Exit(0)
		}
	},
}

func init() {
	smdCmd.AddCommand(smdServiceCmd)
}
