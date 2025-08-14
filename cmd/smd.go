// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/OpenCHAMI/ochami/internal/log"
	"github.com/OpenCHAMI/ochami/pkg/client/smd"
)

// smdGetClient sets up the SMD client with the SMD base URI and certificates
// (if necessary) and returns it. This function is used by each subcommand.
func smdGetClient(cmd *cobra.Command) *smd.SMDClient {
	// Without a base URI, we cannot do anything
	smdBaseURI, err := getBaseURISMD(cmd)
	if err != nil {
		log.Logger.Error().Err(err).Msg("failed to get base URI for SMD")
		logHelpError(cmd)
		os.Exit(1)
	}

	// Create client to make request to SMD
	smdClient, err := smd.NewClient(smdBaseURI, insecure)
	if err != nil {
		log.Logger.Error().Err(err).Msg("error creating new SMD client")
		logHelpError(cmd)
		os.Exit(1)
	}

	// Check if a CA certificate was passed and load it into client if valid
	useCACert(smdClient.OchamiClient)

	return smdClient
}

// smdCmd represents the bss command
var smdCmd = &cobra.Command{
	Use:   "smd",
	Args:  cobra.NoArgs,
	Short: "Communicate with the State Management Database (SMD)",
	Long: `Communicate with the State Management Database (SMD).

See ochami-smd(1) for more details.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			printUsageHandleError(cmd)
			os.Exit(0)
		}
	},
}

func init() {
	smdCmd.PersistentFlags().String("uri", "", "absolute base URI or relative base path of SMD")
	rootCmd.AddCommand(smdCmd)
}
