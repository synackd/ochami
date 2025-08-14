// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/OpenCHAMI/ochami/internal/log"
	"github.com/OpenCHAMI/ochami/pkg/client"
)

// cloudInitDefaultsGetCmd represents the "cloud-init defaults get" command
var cloudInitDefaultsGetCmd = &cobra.Command{
	Use:   "get",
	Args:  cobra.NoArgs,
	Short: "Get cloud-init default meta-data for a cluster",
	Long: `Get cloud-init default meta-data for a cluster.

See ochami-cloud-init(1) for more details.`,
	Example: `  ochami cloud-init defaults get`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create client to use for requests
		cloudInitClient := cloudInitGetClient(cmd)

		// Handle token for this command
		handleToken(cmd)

		// Get data
		henv, err := cloudInitClient.GetDefaults(token)
		if err != nil {
			log.Logger.Error().Err(err).Msgf("failed to get defaults")
			logHelpError(cmd)
			os.Exit(1)
		}

		// Print in desired format
		if outBytes, err := client.FormatBody(henv.Body, formatOutput); err != nil {
			log.Logger.Error().Err(err).Msg("failed to format output")
			logHelpError(cmd)
			os.Exit(1)
		} else {
			fmt.Print(string(outBytes))
		}
	},
}

func init() {
	cloudInitDefaultsGetCmd.Flags().VarP(&formatOutput, "format-output", "F", "format of output printed to standard output")

	cloudInitDefaultsGetCmd.RegisterFlagCompletionFunc("format-output", completionFormatData)

	cloudInitDefaultsCmd.AddCommand(cloudInitDefaultsGetCmd)
}
