// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/OpenCHAMI/ochami/internal/log"
	"github.com/OpenCHAMI/ochami/pkg/client/bss"
)

// bssGetClient sets up the BSS client with the BSS base URI and certificates
// (if necessary) and returns it. If tokenRequired is true, it will ensure that
// the token is set and valid and load it. This function is used by each
// subcommand.
func bssGetClient(cmd *cobra.Command, tokenRequired bool) *bss.BSSClient {
	// Without a base URI, we cannot do anything
	bssBaseURI, err := getBaseURIBSS(cmd)
	if err != nil {
		log.Logger.Error().Err(err).Msg("failed to get base URI for BSS")
		logHelpError(cmd)
		os.Exit(1)
	}

	// Make sure token is set/valid, if required
	if tokenRequired {
		setTokenFromEnvVar(cmd)
		checkToken(cmd)
	}

	// Create client to make request to BSS
	bssClient, err := bss.NewClient(bssBaseURI, insecure)
	if err != nil {
		log.Logger.Error().Err(err).Msg("error creating new BSS client")
		logHelpError(cmd)
		os.Exit(1)
	}

	// Check if a CA certificate was passed and load it into client if valid
	useCACert(bssClient.OchamiClient)

	return bssClient
}

// bssCmd represents the bss command
var bssCmd = &cobra.Command{
	Use:   "bss",
	Args:  cobra.NoArgs,
	Short: "Communicate with the Boot Script Service (BSS)",
	Long: `Communicate with the Boot Script Service (BSS).

See ochami-bss(1) for more details.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			printUsageHandleError(cmd)
			os.Exit(0)
		}
	},
}

func init() {
	bssCmd.PersistentFlags().String("uri", "", "absolute base URI or relative base path of BSS")
	rootCmd.AddCommand(bssCmd)
}
