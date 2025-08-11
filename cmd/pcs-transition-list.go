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
	"github.com/OpenCHAMI/ochami/pkg/format"
)

// pcsTransitionListCmd represents the "pcs transition list" command
var pcsTransitionListCmd = &cobra.Command{
	Use:   "list",
	Args:  cobra.NoArgs,
	Short: "List active PCS transitions",
	Long: `List active PCS transitions.

See ochami-pcs(1) for more details.`,
	Example: `  # List transitions
  ochami pcs transition list`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create client to use for requests
		pcsClient := pcsGetClient(cmd)

		// Handle token for this command
		handleToken(cmd)

		// Get transitions
		transitionsHttpEnv, err := pcsClient.GetTransitions(token)
		if err != nil {
			if errors.Is(err, client.UnsuccessfulHTTPError) {
				log.Logger.Error().Err(err).Msg("PCS transitions request yielded unsuccessful HTTP response")
			} else {
				log.Logger.Error().Err(err).Msg("failed to list PCS transitions")
			}
			logHelpError(cmd)
			os.Exit(1)
		}

		var output interface{}
		err = json.Unmarshal(transitionsHttpEnv.Body, &output)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to unmarshal transitions")
			logHelpError(cmd)
			os.Exit(1)
		}

		// Print output
		if outBytes, err := format.MarshalData(output, formatOutput); err != nil {
			log.Logger.Error().Err(err).Msg("failed to format output")
			logHelpError(cmd)
			os.Exit(1)
		} else {
			fmt.Println(string(outBytes))
		}
	},
}

func init() {
	pcsTransitionListCmd.Flags().VarP(&formatOutput, "format-output", "F", "format of output printed to standard output (json,json-pretty,yaml)")

	pcsTransitionListCmd.RegisterFlagCompletionFunc("format-output", completionFormatData)

	pcsTransitionCmd.AddCommand(pcsTransitionListCmd)
}
