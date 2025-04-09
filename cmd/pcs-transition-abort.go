// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/OpenCHAMI/ochami/internal/log"
	"github.com/OpenCHAMI/ochami/pkg/client"
	"github.com/OpenCHAMI/ochami/pkg/format"
	"github.com/spf13/cobra"
)

// pcsTransitionAbortCmd represents the "pcs transition abort" command
var pcsTransitionAbortCmd = &cobra.Command{
	Use:   "abort <transition_id>",
	Args:  cobra.ExactArgs(1),
	Short: "Abort a PCS transition",
	Long: `Abort a PCS transition.

See ochami-pcs(1) for more details.`,
	Example: `  # Abort a transition
  ochami pcs transition abort 8f252166-c53c-435e-8354-e69649537a0f`,
	Run: func(cmd *cobra.Command, args []string) {
		transitionID := args[0]

		// Create client to use for requests
		pcsClient := pcsGetClient(cmd, true)

		// Abort the transition
		transitionHttpEnv, err := pcsClient.DeleteTransition(transitionID, token)
		if err != nil {
			if errors.Is(err, client.UnsuccessfulHTTPError) {
				log.Logger.Error().Err(err).Msg("PCS transition abort request yielded unsuccessful HTTP response")
			} else {
				log.Logger.Error().Err(err).Msg("failed to abort PCS transition")
			}
			logHelpError(cmd)
			os.Exit(1)
		}

		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to abort transition")
			logHelpError(cmd)
			os.Exit(1)
		}

		var output interface{}
		err = json.Unmarshal(transitionHttpEnv.Body, &output)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to unmarshal abort transitions response")
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
	pcsTransitionAbortCmd.Flags().VarP(&formatOutput, "format-output", "F", "format of output printed to standard output (json,json-pretty,yaml)")

	pcsTransitionAbortCmd.RegisterFlagCompletionFunc("format-output", completionFormatData)

	pcsTransitionCmd.AddCommand(pcsTransitionAbortCmd)
}
