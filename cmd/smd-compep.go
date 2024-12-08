// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"os"

	"github.com/OpenCHAMI/ochami/internal/log"
	"github.com/spf13/cobra"
)

// compepCmd represents the smd-compep command
var compepCmd = &cobra.Command{
	Use:   "compep",
	Args:  cobra.NoArgs,
	Short: "Manage component endpoints",
	Long: `Manage component endpoints. This is a metacommand. Commands under this one
interact with the State Management Database (SMD).`,
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
	smdCmd.AddCommand(compepCmd)
}
