// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/OpenCHAMI/ochami/internal/log"
	"github.com/OpenCHAMI/ochami/pkg/client/pcs"
)

// pcsGetClient sets up the PCS client with the PCS base URI and certificates
// (if necessary) and returns it. If tokenRequired is true, it will ensure that
// the token is set and valid and load it. This function is used by each
// subcommand.
func pcsGetClient(cmd *cobra.Command, tokenRequired bool) *pcs.PCSClient {
	// Without a base URI, we cannot do anything
	pcsBaseURI, err := getBaseURIPCS(cmd)
	if err != nil {
		log.Logger.Error().Err(err).Msg("failed to get base URI for PCS")
		logHelpError(cmd)
		os.Exit(1)
	}

	// Make sure token is set/valid, if required
	if tokenRequired {
		setTokenFromEnvVar(cmd)
		checkToken(cmd)
	}

	// Create client to make request to PCS
	pcsClient, err := pcs.NewClient(pcsBaseURI, insecure)
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("error creating new PCS client")
	}

	// Check if a CA certificate was passed and load it into client if valid
	useCACert(pcsClient.OchamiClient)

	return pcsClient
}

// pcsCmd represents the pcs command
var pcsCmd = &cobra.Command{
	Use:   "pcs",
	Args:  cobra.NoArgs,
	Short: "Interact with the Power Control Service (PCS)",
	Long: `Interact with the Power Control Service (PCS).

See ochami-pcs(1) for more details.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			printUsageHandleError(cmd)
			os.Exit(0)
		}
	},
}

func init() {
	pcsCmd.PersistentFlags().String("uri", "", "absolute base URI or relative base path of PCS")
	rootCmd.AddCommand(pcsCmd)
}
