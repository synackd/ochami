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

// pcsTransition show Cmd represents the "pcs transition show" command
var pcsTransitionShowCmd = &cobra.Command{
	Use:   "show <transition_id>",
	Args:  cobra.ExactArgs(1),
	Short: "Show details of a PCS transition",
	Long: `Show details of a PCS transition.

See ochami-pcs(1) for more details.`,
	Example: `  # Show a transition
  ochami pcs transition show 8f252166-c53c-435e-8354-e69649537a0f`,
	Run: func(cmd *cobra.Command, args []string) {
		transitionID := args[0]

		// Create client to use for requests
		pcsClient := pcsGetClient(cmd)

		// Handle token for this command
		handleToken(cmd)

		// Get transition
		transitionHttpEnv, err := pcsClient.GetTransition(transitionID, token)
		if err != nil {
			if errors.Is(err, client.UnsuccessfulHTTPError) {
				log.Logger.Error().Err(err).Msg("PCS transitions request yielded unsuccessful HTTP response")
			} else {
				log.Logger.Error().Err(err).Msg("failed to get PCS transition")
			}
			logHelpError(cmd)
			os.Exit(1)
		}

		// Unmarshal output
		var output interface{}
		err = json.Unmarshal(transitionHttpEnv.Body, &output)
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
	pcsTransitionShowCmd.Flags().VarP(&formatOutput, "format-output", "F", "format of output printed to standard output (json,json-pretty,yaml)")

	pcsTransitionShowCmd.RegisterFlagCompletionFunc("format-output", completionFormatData)

	pcsTransitionCmd.AddCommand(pcsTransitionShowCmd)
}
