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

// smdStatusCmd represents the "smd status" command
var smdStatusCmd = &cobra.Command{
	Use:   "status",
	Args:  cobra.NoArgs,
	Short: "Get status of the State Management Database (SMD)",
	Long: `Get status of the State Management Database (SMD).

See ochami-smd(1) for more details.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create client to use for requests
		smdClient := smdGetClient(cmd)

		// Determine which component to get status for and send request
		var httpEnv client.HTTPEnvelope
		var err error
		if cmd.Flag("all").Changed {
			httpEnv, err = smdClient.GetStatus("all")
		} else {
			httpEnv, err = smdClient.GetStatus("")
		}
		if err != nil {
			if errors.Is(err, client.UnsuccessfulHTTPError) {
				log.Logger.Error().Err(err).Msg("SMD status request yielded unsuccessful HTTP response")
			} else {
				log.Logger.Error().Err(err).Msg("failed to get SMD status")
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
	smdStatusCmd.Flags().Bool("all", false, "print all status data from SMD")
	smdStatusCmd.Flags().VarP(&formatOutput, "format-output", "F", "format of output printed to standard output (json,json-pretty,yaml)")

	smdStatusCmd.RegisterFlagCompletionFunc("format-output", completionFormatData)

	smdCmd.AddCommand(smdStatusCmd)
}
