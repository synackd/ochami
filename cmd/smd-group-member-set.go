// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/OpenCHAMI/ochami/internal/log"
	"github.com/OpenCHAMI/ochami/pkg/client"
)

// groupMemberSetCmd represents the smd-group-member-set command
var groupMemberSetCmd = &cobra.Command{
	Use:   "set <group_label> <component>...",
	Args:  cobra.MinimumNArgs(2),
	Short: "Set group membership list to a list of components",
	Long: `Set group membership list to a list of components. The components specified
in the list are set as the only members of the group. If a component
specified is already in the group, it remains in the group. If a
component specified is not already in te group, it is added to the
group. If a component is in the group but not specified, it is
removed from the group.

See ochami-smd(1) for more details.`,
	Example: `  ochami smd group member set compute x1000c1s7b1n0 x1000c1s7b2n0`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create client to use for requests
		smdClient := smdGetClient(cmd, true)

		// Send off request
		_, err := smdClient.PutGroupMembers(token, args[0], args[1:]...)
		if err != nil {
			if errors.Is(err, client.UnsuccessfulHTTPError) {
				log.Logger.Error().Err(err).Msgf("SMD group member request for group %s yielded unsuccessful HTTP response", args[0])
			} else {
				log.Logger.Error().Err(err).Msgf("failed to set group membership for group %s in SMD", args[0])
			}
			logHelpError(cmd)
			os.Exit(1)
		}
	},
}

func init() {
	groupMemberCmd.AddCommand(groupMemberSetCmd)
}
