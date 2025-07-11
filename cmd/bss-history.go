// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"

	"github.com/OpenCHAMI/ochami/internal/log"
	"github.com/OpenCHAMI/ochami/pkg/client"
)

// bssHistoryCmd represents the bss-history command
var bssHistoryCmd = &cobra.Command{
	Use:   "history",
	Args:  cobra.NoArgs,
	Short: "Fetch the endpoint history of BSS",
	Long: `Fetch the endpoint history of BSS.

See ochami-bss(1) for more details.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create client to use for requests
		bssClient := bssGetClient(cmd, false)

		// If no ID flags are specified, get all boot parameters
		qstr := ""
		if cmd.Flag("xname").Changed || cmd.Flag("endpoint").Changed {
			values := url.Values{}
			if cmd.Flag("xname").Changed {
				x, err := cmd.Flags().GetString("xname")
				if err != nil {
					log.Logger.Error().Err(err).Msg("unable to fetch xname")
					logHelpError(cmd)
					os.Exit(1)
				}
				values.Add("name", x)
			}
			if cmd.Flag("endpoint").Changed {
				e, err := cmd.Flags().GetString("endpoint")
				if err != nil {
					log.Logger.Error().Err(err).Msg("unable to fetch endpoint")
					logHelpError(cmd)
					os.Exit(1)
				}
				values.Add("endpoint", e)
			}
			qstr = values.Encode()
		}

		// Send request
		httpEnv, err := bssClient.GetEndpointHistory(qstr)
		if err != nil {
			if errors.Is(err, client.UnsuccessfulHTTPError) {
				log.Logger.Error().Err(err).Msg("BSS endpoint history request yielded unsuccessful HTTP response")
			} else {
				log.Logger.Error().Err(err).Msg("failed to request endpoint history from BSS")
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
			fmt.Printf(string(outBytes))
		}
	},
}

func init() {
	bssHistoryCmd.Flags().String("xname", "", "filter by xname")
	bssHistoryCmd.Flags().String("endpoint", "", "filter by endpoint")
	bssHistoryCmd.Flags().VarP(&formatOutput, "format-output", "F", "format of output printed to standard output (json,json-pretty,yaml)")

	bssHistoryCmd.RegisterFlagCompletionFunc("format-output", completionFormatData)

	bssCmd.AddCommand(bssHistoryCmd)
}
