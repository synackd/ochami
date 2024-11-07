// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/OpenCHAMI/ochami/internal/log"
)

// bootParamsCmd represents the boot-params command
var bootParamsCmd = &cobra.Command{
	Use:   "params",
	Args:  cobra.NoArgs,
	Short: "Work with boot parameters for components",
	Long: `Work with boot parameters for components, including kernel URI, initrd URI,
and kernel command line arguments. This is a metacommand.`,
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
	bootCmd.AddCommand(bootParamsCmd)
}
