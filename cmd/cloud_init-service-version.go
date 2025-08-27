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

// cloudInitServiceVersionCmd represents the "cloud-init service status" command
var cloudInitServiceVersionCmd = &cobra.Command{
	Use:   "version",
	Args:  cobra.NoArgs,
	Short: "Print version of the cloud-init metadata service",
	Long: `Print version of the cloud-init metadata service.

See ochami-cloud-init(1) for more details.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create client to use for requests
		cloudInitClient := cloudInitGetClient(cmd)

		henv, err := cloudInitClient.GetVersion()
		if err != nil {
			if errors.Is(err, client.UnsuccessfulHTTPError) {
				log.Logger.Error().Err(err).Msg("cloud-init version request yielded unsuccessful HTTP response")
			} else {
				log.Logger.Error().Err(err).Msg("failed to get cloud-init version")
			}
			logHelpError(cmd)
			os.Exit(1)
		}

		if outBytes, err := client.FormatBody(henv.Body, formatOutput); err != nil {
			log.Logger.Error().Err(err).Msg("failed to format output")
			logHelpError(cmd)
			os.Exit(1)
		} else {
			fmt.Print(string(outBytes))
		}
	},
}

func init() {
	cloudInitServiceVersionCmd.Flags().VarP(&formatOutput, "format-output", "F", "format of output printed to standard output (json,json-pretty,yaml)")

	cloudInitServiceVersionCmd.RegisterFlagCompletionFunc("format-output", completionFormatData)

	cloudInitServiceCmd.AddCommand(cloudInitServiceVersionCmd)
}
