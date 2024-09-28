// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/synackd/ochami/internal/client"
	"github.com/synackd/ochami/internal/log"
)

// smdStatusCmd represents the smd-status command
var smdStatusCmd = &cobra.Command{
	Use:   "status",
	Args:  cobra.NoArgs,
	Short: "Get status of SMD service",
	Run: func(cmd *cobra.Command, args []string) {
		// Without a base URI, we cannot do anything
		smdBaseURI, err := getBaseURI(cmd)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to get base URI for SMD")
			os.Exit(1)
		}

		// Create client to make request to SMD
		smdClient, err := client.NewSMDClient(smdBaseURI, insecure)
		if err != nil {
			log.Logger.Error().Err(err).Msg("error creating new SMD client")
			os.Exit(1)
		}

		// Check if a CA certificate was passed and load it into client if valid
		useCACert(smdClient.OchamiClient)

		// Determine which component to get status for and send request
		var httpEnv client.HTTPEnvelope
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
			os.Exit(1)
		}
		fmt.Println(string(httpEnv.Body))
	},
}

func init() {
	smdStatusCmd.Flags().Bool("all", false, "print all status data from SMD")

	smdCmd.AddCommand(smdStatusCmd)
}
