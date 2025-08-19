// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"github.com/spf13/cobra"
)

// pcsServiceCmd represents the "pcs transitions" command
var pcsServiceCmd = &cobra.Command{
	Use:   "service",
	Args:  cobra.NoArgs,
	Short: "Manage and check PCS itself",
	Long: `Manage and check PCS itself. This is a metacommand.

See ochami-pcs(1) for more details.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			printUsageHandleError(cmd)
		}
	},
}

func init() {
	pcsCmd.AddCommand(pcsServiceCmd)
}
