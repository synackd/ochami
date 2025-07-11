// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/OpenCHAMI/ochami/internal/log"
	"github.com/OpenCHAMI/ochami/pkg/client/ci"
)

type CIFlagHeaderWhen string

const (
	CIFlagHeaderAlways   = "always"
	CIFlagHeaderMultiple = "multiple"
	CIFlagHeaderNever    = "never"
)

var (
	CIFlagHeaderWhenHelp = map[string]string{
		string(CIFlagHeaderAlways):   "Always print headers, even if singular output",
		string(CIFlagHeaderMultiple): "Only print headers if multiple items in output",
		string(CIFlagHeaderNever):    "Never print headers",
	}
	ciHeaderWhen CIFlagHeaderWhen = CIFlagHeaderMultiple
)

func (cfhw CIFlagHeaderWhen) String() string {
	return string(cfhw)
}

func (cfhw *CIFlagHeaderWhen) Set(v string) error {
	switch CIFlagHeaderWhen(v) {
	case CIFlagHeaderAlways,
		CIFlagHeaderMultiple,
		CIFlagHeaderNever:
		*cfhw = CIFlagHeaderWhen(v)
		return nil
	default:
		return fmt.Errorf("must be one of %v", []CIFlagHeaderWhen{
			CIFlagHeaderAlways,
			CIFlagHeaderMultiple,
			CIFlagHeaderNever,
		})
	}
}

func (cfhw CIFlagHeaderWhen) Type() string {
	return "CIFlagHeaderWhen"
}

func cloudInitCompletionHeaderWhen(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var helpSlice []string
	for k, v := range CIFlagHeaderWhenHelp {
		helpSlice = append(helpSlice, fmt.Sprintf("%s\t%s", k, v))
	}
	return helpSlice, cobra.ShellCompDirectiveDefault
}

// cloudInitGetClient sets up the cloud-init client with the cloud-init base URI
// and certificates (if necessary) and returns it. If tokenRequired is true,
// it will ensure that the token is set and valid and load it. This function is
// used by each subcommand.
func cloudInitGetClient(cmd *cobra.Command, tokenRequired bool) *ci.CloudInitClient {
	// Without a base URI, we cannot do anything
	cloudInitbaseURI, err := getBaseURICloudInit(cmd)
	if err != nil {
		log.Logger.Error().Err(err).Msg("failed to get base URI for cloud-init")
		logHelpError(cmd)
		os.Exit(1)
	}

	// Make sure token is set/valid, if required
	if tokenRequired {
		setTokenFromEnvVar(cmd)
		checkToken(cmd)
	}

	// Create client to make request to cloud-init
	cloudInitClient, err := ci.NewClient(cloudInitbaseURI, insecure)
	if err != nil {
		log.Logger.Error().Err(err).Msg("error creating new cloud-init client")
		logHelpError(cmd)
		os.Exit(1)
	}

	// Check if a CA certificate was passed and load it into client if valid
	useCACert(cloudInitClient.OchamiClient)

	return cloudInitClient
}

// cloudInitCmd represents the "cloud-init" command
var cloudInitCmd = &cobra.Command{
	Use:   "cloud-init",
	Args:  cobra.NoArgs,
	Short: "Interact with the cloud-init service",
	Long: `Interact with the cloud-init service. This is a metacommand.

See ochami-cloud-init(1) for more details.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			printUsageHandleError(cmd)
			os.Exit(0)
		}
	},
}

func init() {
	cloudInitCmd.PersistentFlags().String("uri", "", "absolute base URI or relative base path of cloud-init")
	rootCmd.AddCommand(cloudInitCmd)
}
