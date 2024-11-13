// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/OpenCHAMI/ochami/internal/log"
)

// bootScriptCmd represents the bss-boot-script command
var bootScriptCmd = &cobra.Command{
	Use:   "script",
	Args:  cobra.NoArgs,
	Short: "Work with boot scripts for components",
	Long:  `Work with boot scripts for components. This is a metacommand.`,
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
	bootCmd.AddCommand(bootScriptCmd)
}
