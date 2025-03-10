// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"errors"
	"os"

	"github.com/OpenCHAMI/ochami/internal/log"
	"github.com/OpenCHAMI/ochami/pkg/client"
	"github.com/OpenCHAMI/ochami/pkg/client/smd"
	"github.com/spf13/cobra"
)

// componentAddCmd represents the smd-component-add command
var componentAddCmd = &cobra.Command{
	Use:   "add -f <payload_file> | (<xname> <node_id>)",
	Short: "Add new component(s)",
	Long: `Add new component(s). A name (xname) and node ID (int64) are required unless
-f is passed to read from a payload file. Specifying -f also is
mutually exclusive with the other flags of this command. If - is
used as the argument to -f, the data is read from standard input.

This command sends a POST to SMD. An access token is required.`,
	Example: `  ochami smd component add x3000c1s7b56n0 56
  ochami smd component add --state Ready --enabled --role Compute --arch X86 x3000c1s7b56n0 56
  ochami smd component add -f payload.json
  ochami smd component add -f payload.yaml --payload-format yaml
  echo '<json_data>' | ochami smd component add -f -
  echo '<yaml_data>' | ochami smd component add -f - --payload-format yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check that all required args are passed
		if len(args) == 0 && !cmd.Flag("payload").Changed {
			err := cmd.Usage()
			if err != nil {
				log.Logger.Error().Err(err).Msg("failed to print usage")
				os.Exit(1)
			}
			os.Exit(0)
		} else if len(args) > 2 {
			log.Logger.Error().Msgf("expected 2 arguments (xname, nid) but got %d: %v", len(args), args)
			os.Exit(1)
		}

		// Without a base URI, we cannot do anything
		smdBaseURI, err := getBaseURI(cmd)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to get base URI for SMD")
			os.Exit(1)
		}

		// This endpoint requires authentication, so a token is needed
		setTokenFromEnvVar(cmd)
		checkToken(cmd)

		// Create client to make request to SMD
		smdClient, err := smd.NewClient(smdBaseURI, insecure)
		if err != nil {
			log.Logger.Error().Err(err).Msg("error creating new SMD client")
			os.Exit(1)
		}

		// Check if a CA certificate was passed and load it into client if valid
		useCACert(smdClient.OchamiClient)

		var compSlice smd.ComponentSlice
		if cmd.Flag("payload").Changed {
			handlePayload(cmd, &compSlice)
		} else {
			// ...otherwise use CLI options
			comp := smd.Component{
				ID:    args[0],
				State: cmd.Flag("state").Value.String(),
				Role:  cmd.Flag("role").Value.String(),
				Arch:  cmd.Flag("arch").Value.String(),
			}
			comp.Enabled, err = cmd.Flags().GetBool("enabled")
			if err != nil {
				log.Logger.Error().Err(err).Msg("failed to retrieve flag 'enabled', defaulting to true")
				comp.Enabled = true
			}

			compSlice.Components = append(compSlice.Components, comp)
		}

		// Send off request
		_, err = smdClient.PostComponents(compSlice, token)
		if err != nil {
			if errors.Is(err, client.UnsuccessfulHTTPError) {
				log.Logger.Error().Err(err).Msg("SMD component request yielded unsuccessful HTTP response")
			} else {
				log.Logger.Error().Err(err).Msg("failed to add component(s) to SMD")
			}
			os.Exit(1)
		}
	},
}

func init() {
	componentAddCmd.Flags().String("state", "Ready", "set readiness state of new component")
	componentAddCmd.Flags().Bool("enabled", true, "set if new component is enabled")
	componentAddCmd.Flags().String("role", "Compute", "role of new component")
	componentAddCmd.Flags().String("arch", "X86", "CPU architecture of new component")
	componentAddCmd.Flags().StringP("payload", "f", "", "file containing the request payload; JSON format unless --payload-format specified")
	componentAddCmd.Flags().StringP("payload-format", "F", defaultPayloadFormat, "format of payload file (yaml,json) passed with --payload")

	componentAddCmd.MarkFlagsMutuallyExclusive("state", "payload")
	componentAddCmd.MarkFlagsMutuallyExclusive("enabled", "payload")
	componentAddCmd.MarkFlagsMutuallyExclusive("role", "payload")
	componentAddCmd.MarkFlagsMutuallyExclusive("arch", "payload")

	componentCmd.AddCommand(componentAddCmd)
}
