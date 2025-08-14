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

// groupMemberGetCmd represents the "smd group member get" command
var groupMemberGetCmd = &cobra.Command{
	Use:   "get <group_label>",
	Args:  cobra.ExactArgs(1),
	Short: "Get members of a group",
	Long: `Get members of a group.

See ochami-smd(1) for more details.`,
	Example: `  ochami smd group member get compute`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create client to use for requests
		smdClient := smdGetClient(cmd)

		// Handle token for this command
		handleToken(cmd)

		// Send request
		httpEnv, err := smdClient.GetGroupMembers(args[0], token)
		if err != nil {
			if errors.Is(err, client.UnsuccessfulHTTPError) {
				log.Logger.Error().Err(err).Msg("SMD group member request yielded unsuccessful HTTP response")
			} else {
				log.Logger.Error().Err(err).Msg("failed to request group members from SMD")
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
	groupMemberGetCmd.Flags().VarP(&formatOutput, "format-output", "F", "format of output printed to standard output (json,json-pretty,yaml)")

	groupMemberGetCmd.RegisterFlagCompletionFunc("format-output", completionFormatData)

	groupMemberCmd.AddCommand(groupMemberGetCmd)
}
