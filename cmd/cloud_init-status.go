// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// cloudInitServiceCmd represents the "cloud-init service" command
var cloudInitServiceCmd = &cobra.Command{
	Use:   "service",
	Args:  cobra.NoArgs,
	Short: "Manage and check cloud-init itself",
	Long: `Manage and check cloud-init itself. This is a metacommand.

See ochami-cloud-init(1) for more details.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			printUsageHandleError(cmd)
			os.Exit(0)
		}
	},
}

func init() {
	cloudInitCmd.AddCommand(cloudInitServiceCmd)
}
