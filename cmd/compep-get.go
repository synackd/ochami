// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/OpenCHAMI/ochami/internal/client"
	"github.com/OpenCHAMI/ochami/internal/log"
)

// compepGetCmd represents the compep-get command
var compepGetCmd = &cobra.Command{
	Use:   "get [<xname>...]",
	Short: "Get all component endpoints or one identified by an xname",
	Run: func(cmd *cobra.Command, args []string) {
		// Without a base URI, we cannot do anything
		smdBaseURI, err := getBaseURI(cmd)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to get base URI for SMD")
			os.Exit(1)
		}

		// This endpoint requires authentication, so a token is needed
		setTokenFromEnvVar(cmd)
		checkToken(cmd)

		// Create client to make request to SMD
		smdClient, err := client.NewSMDClient(smdBaseURI, insecure)
		if err != nil {
			log.Logger.Error().Err(err).Msg("error creating new SMD client")
			os.Exit(1)
		}

		// Check if a CA certificate was passed and load it into client if valid
		useCACert(smdClient.OchamiClient)

		var httpEnv client.HTTPEnvelope
		if len(args) == 0 {
			// Get all ComponentEndpoints if no args passed
			httpEnv, err = smdClient.GetComponentEndpointsAll(token)
			if err != nil {
				if errors.Is(err, client.UnsuccessfulHTTPError) {
					log.Logger.Error().Err(err).Msg("SMD component endpoimt request yielded unsuccessful HTTP response")
				} else {
					log.Logger.Error().Err(err).Msg("failed to request component endpoints from SMD")
				}
				os.Exit(1)
			}
			fmt.Println(string(httpEnv.Body))
		} else {
			httpEnvs, errs, err := smdClient.GetComponentEndpoints(token, args...)
			if err != nil {
				log.Logger.Error().Err(err).Msg("failed to get component endpoints from SMD")
				os.Exit(1)
			}
			// Since smdClient.GetComponentEndpoints does the deletion iteratively, we need to
			// deal with each error that might have occurred.
			var errorsOccurred = false
			for _, e := range errs {
				if errors.Is(e, client.UnsuccessfulHTTPError) {
					errorsOccurred = true
					log.Logger.Error().Err(e).Msg("SMD redfish endpoint deletion yielded unsuccessful HTTP response")
				} else if e != nil {
					errorsOccurred = true
					log.Logger.Error().Err(e).Msg("failed to delete redfish endpoint")
				}
			}

			// Put selected ComponentEndpoints into array and marshal
			type compEp struct {
				ComponentEndpoints []interface{} `json:"ComponentEndpoints"`
			}
			var ceArr []interface{}
			for i, h := range httpEnvs {
				if errs[i] == nil {
					var ce interface{}
					err := json.Unmarshal(h.Body, &ce)
					if err != nil {
						log.Logger.Warn().Err(err).Msg("failed to unmarshal component endpoint")
						continue
					}
					ceArr = append(ceArr, ce)
				}
			}

			// Warn the user if any errors occurred during deletion iterations
			if errorsOccurred {
				log.Logger.Warn().Msg("SMD redfish endpoint deletion completed with errors")
				os.Exit(1)
			}

			ces := compEp{ComponentEndpoints: ceArr}
			cesStr, err := json.Marshal(ces)
			if err != nil {
				log.Logger.Error().Err(err).Msg("failed to unmarshal list of component endpoints")
				os.Exit(1)
			}
			fmt.Println(string(cesStr))
		}
	},
}

func init() {
	compepCmd.AddCommand(compepGetCmd)
}
