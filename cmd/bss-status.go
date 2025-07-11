// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/OpenCHAMI/ochami/internal/log"
	"github.com/OpenCHAMI/ochami/pkg/client"
)

// bssStatusCmd represents the bss-status command
var bssStatusCmd = &cobra.Command{
	Use:   "status",
	Args:  cobra.NoArgs,
	Short: "Get status of the Boot Script Service (BSS)",
	Long: `Get status of the Boot Script Service (BSS).

See ochami-bss(1) for more details.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create client to use for requests
		bssClient := bssGetClient(cmd, false)

		// Determine which component to get status for and send request
		var httpEnv client.HTTPEnvelope
		var err error
		if cmd.Flag("all").Changed {
			httpEnv, err = bssClient.GetStatus("all")
		} else if cmd.Flag("storage").Changed {
			httpEnv, err = bssClient.GetStatus("storage")
		} else if cmd.Flag("smd").Changed {
			httpEnv, err = bssClient.GetStatus("smd")
		} else if cmd.Flag("version").Changed {
			httpEnv, err = bssClient.GetStatus("version")
		} else {
			httpEnv, err = bssClient.GetStatus("")
		}
		if err != nil {
			if errors.Is(err, client.UnsuccessfulHTTPError) {
				log.Logger.Error().Err(err).Msg("BSS status request yielded unsuccessful HTTP response")
			} else {
				log.Logger.Error().Err(err).Msg("failed to get BSS status")
			}
			logHelpError(cmd)
			os.Exit(1)
		}

		// Print output
		if outBytes, err := client.FormatBody(httpEnv.Body, formatOutput); err != nil {
			log.Logger.Error().Err(err).Msg("failed to format output")
			logHelpError(cmd)
			os.Exit(1)
		} else {
			fmt.Print(string(outBytes))
		}
	},
}

func init() {
	bssStatusCmd.Flags().Bool("all", false, "print all status data from BSS")
	bssStatusCmd.Flags().Bool("storage", false, "print status of storage backend from BSS")
	bssStatusCmd.Flags().Bool("smd", false, "print status of BSS connection to SMD")
	bssStatusCmd.Flags().Bool("version", false, "print version of BSS")
	bssStatusCmd.Flags().VarP(&formatOutput, "format-output", "F", "format of output printed to standard output (json,json-pretty,yaml)")

	bssStatusCmd.RegisterFlagCompletionFunc("format-output", completionFormatData)
	bssStatusCmd.MarkFlagsMutuallyExclusive("all", "storage", "smd", "version")

	bssCmd.AddCommand(bssStatusCmd)
}
