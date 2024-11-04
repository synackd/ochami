// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/synackd/ochami/internal/log"
)

// cloudInitCmd represents the cloud-init command
var cloudInitCmd = &cobra.Command{
	Use:   "cloud-init",
	Args:  cobra.NoArgs,
	Short: "Interact with the cloud-init service",
	Long:  `Interact with the cloud-init service. This is a metacommand.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			err := cmd.Usage()
			if err != nil {
				log.Logger.Error().Err(err).Msg("failed to print usage")
				os.Exit(1)
			}
			os.Exit(0)
		}
	},
}

func init() {
	rootCmd.AddCommand(cloudInitCmd)
}
