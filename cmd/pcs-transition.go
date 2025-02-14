// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"github.com/spf13/cobra"
)

// pcsTransitionCmd represents the "pcs transitions" command
var pcsTransitionCmd = &cobra.Command{
	Use:   "transition",
	Args:  cobra.NoArgs,
	Short: "Manage PCS transitions",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			printUsageHandleError(cmd)
		}
	},
}

func init() {
	pcsCmd.AddCommand(pcsTransitionCmd)
}
