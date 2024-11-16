// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/OpenCHAMI/ochami/internal/client"
	"github.com/OpenCHAMI/ochami/internal/log"
)

// componentGetCmd represents the smd-component-get command
var componentGetCmd = &cobra.Command{
	Use:   "get",
	Args:  cobra.NoArgs,
	Short: "Get all components or component identified by an xname or node ID",
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

		var httpEnv client.HTTPEnvelope
		if cmd.Flag("xname").Changed {
			// This endpoint requires authentication, so a token is needed
			setTokenFromEnvVar(cmd)
			checkToken(cmd)

			httpEnv, err = smdClient.GetComponentsXname(cmd.Flag("xname").Value.String(), token)
		} else if cmd.Flag("nid").Changed {
			// This endpoint requires authentication, so a token is needed
			setTokenFromEnvVar(cmd)
			checkToken(cmd)

			var nid int32
			nid, err = cmd.Flags().GetInt32("nid")
			if err != nil {
				log.Logger.Error().Err(err).Msg("error getting nid from flag")
				os.Exit(1)
			}
			httpEnv, err = smdClient.GetComponentsNid(nid, token)
		} else {
			httpEnv, err = smdClient.GetComponentsAll()
		}
		if err != nil {
			if errors.Is(err, client.UnsuccessfulHTTPError) {
				log.Logger.Error().Err(err).Msg("SMD component request yielded unsuccessful HTTP response")
			} else {
				log.Logger.Error().Err(err).Msg("failed to request components from SMD")
			}
			os.Exit(1)
		}
		fmt.Println(string(httpEnv.Body))
	},
}

func init() {
	componentGetCmd.Flags().StringP("xname", "x", "", "xname whose Component to fetch")
	componentGetCmd.Flags().Int32P("nid", "n", 0, "node ID whose Component to fetch")

	componentGetCmd.MarkFlagsMutuallyExclusive("xname", "nid")

	componentCmd.AddCommand(componentGetCmd)
}