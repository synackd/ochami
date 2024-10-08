// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"errors"
	"os"

	"github.com/OpenCHAMI/bss/pkg/bssTypes"
	"github.com/spf13/cobra"
	"github.com/synackd/ochami/internal/client"
	"github.com/synackd/ochami/internal/log"
)

// bootParamsDeleteCmd represents the bss-bootparams-delete command
var bootParamsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete boot parameters for one or more components",
	Long: `Delete boot parameters for one or more components. At least one of --kernel,
--initrd, --params, --xname, --mac, or --nid must be specified. This command can delete
boot parameters by config (kernel URI, initrd URI, or kernel command line) or by component
(--xname, --mac, or --nid). The user will be asked for confirmation before deletion unless
--force is passed. Alternatively, pass -f to pass a file (optionally specifying --payload-format,
JSON by default), but the rules above still apply for the payload.

This command sends a DELETE to BSS. An access token is required.`,
	Example: `  ochami boot params delete --kernel https://example.com/kernel
  ochami boot params delete --kernel https://example.com/kernel --initrd https://example.com/initrd
  ochami boot params delete -f payload.json
  ochami boot params delete -f payload.yaml --format-input yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		// cmd.LocalFlags().NFlag() doesn't seem to work, so we check every flag
		if len(args) == 0 &&
			!cmd.Flag("xname").Changed && !cmd.Flag("nid").Changed && !cmd.Flag("mac").Changed &&
			!cmd.Flag("kernel").Changed && !cmd.Flag("initrd").Changed && !cmd.Flag("payload").Changed {
			err := cmd.Usage()
			if err != nil {
				log.Logger.Error().Err(err).Msg("failed to print usage")
				os.Exit(1)
			}
			os.Exit(0)
		}

		// Without a base URI, we cannot do anything
		bssBaseURI, err := getBaseURI(cmd)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to get base URI for BSS")
			os.Exit(1)
		}

		// This endpoint requires authentication, so a token is needed
		setTokenFromEnvVar(cmd)
		checkToken(cmd)

		// Create client to make request to BSS
		bssClient, err := client.NewBSSClient(bssBaseURI, insecure)
		if err != nil {
			log.Logger.Error().Err(err).Msg("error creating new BSS client")
			os.Exit(1)
		}

		// Check if a CA certificate was passed and load it into client if valid
		useCACert(bssClient.OchamiClient)

		// The BSS BootParams struct we will send
		bp := bssTypes.BootParams{}

		// Read payload from file first, allowing overwrites from flags
		if cmd.Flag("payload").Changed {
			dFile := cmd.Flag("payload").Value.String()
			dFormat := cmd.Flag("payload-format").Value.String()
			err := client.ReadPayload(dFile, dFormat, &bp)
			if err != nil {
				log.Logger.Error().Err(err).Msg("unable to read payload for request")
				os.Exit(1)
			}
		}

		// Set the hosts the boot parameters are for
		if cmd.Flag("xname").Changed {
			bp.Hosts, err = cmd.Flags().GetStringSlice("xname")
			if err != nil {
				log.Logger.Error().Err(err).Msg("unable to fetch xname list")
				os.Exit(1)
			}
		}
		if cmd.Flag("mac").Changed {
			bp.Macs, err = cmd.Flags().GetStringSlice("mac")
			if err != nil {
				log.Logger.Error().Err(err).Msg("unable to fetch mac list")
				os.Exit(1)
			}
			if err = bp.CheckMacs(); err != nil {
				log.Logger.Error().Err(err).Msg("invalid mac(s)")
				os.Exit(1)
			}
		}
		if cmd.Flag("nid").Changed {
			bp.Nids, err = cmd.Flags().GetInt32Slice("nid")
			if err != nil {
				log.Logger.Error().Err(err).Msg("unable to fetch nid list")
				os.Exit(1)
			}
		}

		// Set the boot parameters
		if cmd.Flag("kernel").Changed {
			bp.Kernel, err = cmd.Flags().GetString("kernel")
			if err != nil {
				log.Logger.Error().Err(err).Msg("unable to fetch kernel uri")
				os.Exit(1)
			}
		}
		if cmd.Flag("initrd").Changed {
			bp.Initrd, err = cmd.Flags().GetString("initrd")
			if err != nil {
				log.Logger.Error().Err(err).Msg("unable to fetch initrd uri")
				os.Exit(1)
			}
		}
		if cmd.Flag("params").Changed {
			bp.Params, err = cmd.Flags().GetString("params")
			if err != nil {
				log.Logger.Error().Err(err).Msg("unable to fetch params")
				os.Exit(1)
			}
		}

		// Ask before attempting deletion unless --force was passed
		if !cmd.Flag("force").Changed {
			log.Logger.Debug().Msg("--force not passed, prompting user to confirm deletion")
			respDelete := loopYesNo("Really delete?")
			if !respDelete {
				log.Logger.Info().Msg("User aborted boot parameter deletion")
				os.Exit(0)
			} else {
				log.Logger.Debug().Msg("User answered affirmatively to delete boot parameters")
			}
		}

		// Send 'em off
		_, err = bssClient.DeleteBootParams(bp, token)
		if err != nil {
			if errors.Is(err, client.UnsuccessfulHTTPError) {
				log.Logger.Error().Err(err).Msg("BSS boot parameter request yielded unsuccessful HTTP response")
			} else {
				log.Logger.Error().Err(err).Msg("failed to set boot parameters in BSS")
			}
			os.Exit(1)
		}
	},
}

func init() {
	bootParamsDeleteCmd.Flags().String("kernel", "", "URI of kernel")
	bootParamsDeleteCmd.Flags().String("initrd", "", "URI of initrd/initramfs")
	bootParamsDeleteCmd.Flags().String("params", "", "kernel parameters")
	bootParamsDeleteCmd.Flags().StringSliceP("xname", "x", []string{}, "one or more xnames whose boot parameters to delete")
	bootParamsDeleteCmd.Flags().StringSliceP("mac", "m", []string{}, "one or more MAC addresses whose boot parameters to delete")
	bootParamsDeleteCmd.Flags().Int32SliceP("nid", "n", []int32{}, "one or more node IDs whose boot parameters to delete")
	bootParamsDeleteCmd.Flags().StringP("payload", "f", "", "file containing the request payload; JSON format unless --payload-format specified")
	bootParamsDeleteCmd.Flags().String("payload-format", defaultPayloadFormat, "format of payload file (yaml,json) passed with --payload")
	bootParamsDeleteCmd.Flags().Bool("force", false, "do not ask before attempting deletion")

	// We can delete either by component or by boot parameters
	bootParamsDeleteCmd.MarkFlagsOneRequired("xname", "mac", "nid", "kernel", "initrd", "params", "payload")

	bootParamsCmd.AddCommand(bootParamsDeleteCmd)
}