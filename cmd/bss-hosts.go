// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/synackd/ochami/internal/log"
)

// bssHostsCmd represents the hosts command
var bssHostsCmd = &cobra.Command{
	Use:   "hosts",
	Args:  cobra.NoArgs,
	Short: "Work with hosts in BSS",
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
	bssCmd.AddCommand(bssHostsCmd)
}
