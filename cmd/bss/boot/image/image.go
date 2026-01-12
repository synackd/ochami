// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package image

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/OpenCHAMI/ochami/internal/cli"
)

func NewCmd() *cobra.Command {
	// bssBootImageCmd represents the "bss boot image" command
	var bssBootImageCmd = &cobra.Command{
		Use:   "image",
		Args:  cobra.NoArgs,
		Short: "Get and set boot image for nodes",
		Long: `Get and set boot image for nodes. This is a metacommand.

See ochami-bss(1) for more details.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cli.PrintUsageHandleError(cmd)
				os.Exit(0)
			}
		},
	}

	// Add subcommands
	bssBootImageCmd.AddCommand(
		newCmdBootImageSet(),
	)

	return bssBootImageCmd
}
