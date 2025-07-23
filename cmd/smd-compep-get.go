// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/OpenCHAMI/ochami/internal/log"
	"github.com/OpenCHAMI/ochami/pkg/client"
)

// compepGetCmd represents the "smd compep get" command
var compepGetCmd = &cobra.Command{
	Use:   "get [<xname>...]",
	Short: "Get all component endpoints or a subset, identified by xname",
	Long: `Get all component endpoints or a subset, identified by xname.

See ochami-smd(1) for more details.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create client to use for requests
		smdClient := smdGetClient(cmd, true)

		var httpEnv client.HTTPEnvelope
		var err error
		if len(args) == 0 {
			// Get all ComponentEndpoints if no args passed
			httpEnv, err = smdClient.GetComponentEndpointsAll(token)
			if err != nil {
				if errors.Is(err, client.UnsuccessfulHTTPError) {
					log.Logger.Error().Err(err).Msg("SMD component endpoimt request yielded unsuccessful HTTP response")
				} else {
					log.Logger.Error().Err(err).Msg("failed to request component endpoints from SMD")
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
		} else {
			httpEnvs, errs, err := smdClient.GetComponentEndpoints(token, args...)
			if err != nil {
				log.Logger.Error().Err(err).Msg("failed to get component endpoints from SMD")
				logHelpError(cmd)
				os.Exit(1)
			}
			// Since smdClient.GetComponentEndpoints does the deletion iteratively, we need to
			// deal with each error that might have occurred.
			var errorsOccurred = false
			for _, e := range errs {
				if err != nil {
					if errors.Is(e, client.UnsuccessfulHTTPError) {
						log.Logger.Error().Err(e).Msg("SMD redfish endpoint deletion yielded unsuccessful HTTP response")
					} else {
						log.Logger.Error().Err(e).Msg("failed to delete redfish endpoint")
					}
					errorsOccurred = true
				}
			}

			// Put selected ComponentEndpoints into array and marshal
			type compEp struct {
				ComponentEndpoints []interface{} `json:"ComponentEndpoints" yaml:"ComponentEndpoints"`
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
				logHelpError(cmd)
				log.Logger.Warn().Msg("SMD redfish endpoint deletion completed with errors")
				os.Exit(1)
			}

			ces := compEp{ComponentEndpoints: ceArr}
			cesBytes, err := json.Marshal(ces)
			if err != nil {
				log.Logger.Error().Err(err).Msg("failed to unmarshal list of component endpoints")
				logHelpError(cmd)
				os.Exit(1)
			}

			// Print output
			if outBytes, err := client.FormatBody(cesBytes, formatOutput); err != nil {
				log.Logger.Error().Err(err).Msg("failed to format output")
				logHelpError(cmd)
				os.Exit(1)
			} else {
				fmt.Print(string(outBytes))
			}
		}
	},
}

func init() {
	compepGetCmd.Flags().VarP(&formatOutput, "format-output", "F", "format of output printed to standard output (json,json-pretty,yaml)")

	compepGetCmd.RegisterFlagCompletionFunc("format-output", completionFormatData)

	compepCmd.AddCommand(compepGetCmd)
}
