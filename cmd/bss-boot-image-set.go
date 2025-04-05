// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/OpenCHAMI/bss/pkg/bssTypes"
	"github.com/OpenCHAMI/ochami/internal/log"
	"github.com/OpenCHAMI/ochami/pkg/client"
	"github.com/OpenCHAMI/ochami/pkg/client/bss"
	"github.com/OpenCHAMI/ochami/pkg/format"
	"github.com/spf13/cobra"
	kargs "github.com/synackd/go-kargs"
)

// bootImageSetCmd represents the "bss boot image set" command
var bootImageSetCmd = &cobra.Command{
	Use:   "set (-x <xname>[,...] | -m <mac>[,...] | -n <nid>[,...]) <image>",
	Args:  cobra.ExactArgs(1),
	Short: "Set root= kernel command line for one or more nodes, overwriting any previously set",
	Long: `Set root= kernel command line for one or more nodes, overwriting any previously set.
At least one of --xname, --mac, or --nid is required to tell ochami which
components need modification.

An access token is required.

See ochami-bss(1) for more details.`,
	Example: `  # Set nodes to boot live image
  ochami bss boot image set --mac 00:de:ad:be:ef:00,de:ca:fc:0f:fe:ee live:https://172.16.0.254/image.squashfs`,
	Run: func(cmd *cobra.Command, args []string) {
		// Without a base URI, we cannot do anything
		bssBaseURI, err := getBaseURIBSS(cmd)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to get base URI for BSS")
			logHelpError(cmd)
			os.Exit(1)
		}

		// This endpoint requires authentication, so a token is needed
		setTokenFromEnvVar(cmd)
		checkToken(cmd)

		// Create client to make request to BSS
		bssClient, err := bss.NewClient(bssBaseURI, insecure)
		if err != nil {
			log.Logger.Error().Err(err).Msg("error creating new BSS client")
			logHelpError(cmd)
			os.Exit(1)
		}

		// Check if a CA certificate was passed and load it into client if valid
		useCACert(bssClient.OchamiClient)

		// Get current kernel command line args
		values := url.Values{}
		if cmd.Flag("xname").Changed {
			s, err := cmd.Flags().GetStringSlice("xname")
			if err != nil {
				log.Logger.Error().Err(err).Msg("unable to fetch xname list")
				logHelpError(cmd)
				os.Exit(1)
			}
			for _, x := range s {
				values.Add("name", x)
			}
		}
		if cmd.Flag("mac").Changed {
			s, err := cmd.Flags().GetStringSlice("mac")
			if err != nil {
				log.Logger.Error().Err(err).Msg("unable to fetch mac list")
				logHelpError(cmd)
				os.Exit(1)
			}
			for _, m := range s {
				values.Add("mac", m)
			}
		}
		if cmd.Flag("nid").Changed {
			s, err := cmd.Flags().GetInt32Slice("nid")
			if err != nil {
				log.Logger.Error().Err(err).Msg("unable to fetch nid list")
				logHelpError(cmd)
				os.Exit(1)
			}
			for _, n := range s {
				values.Add("nid", fmt.Sprintf("%d", n))
			}
		}
		qstr := values.Encode()
		httpEnv, err := bssClient.GetBootParams(qstr, token)
		if err != nil {
			if errors.Is(err, client.UnsuccessfulHTTPError) {
				log.Logger.Error().Err(err).Msg("BSS boot parameter request yielded unsuccessful HTTP response")
			} else {
				log.Logger.Error().Err(err).Msg("failed to request boot parameters from BSS")
			}
			logHelpError(cmd)
			os.Exit(1)
		}
		var bp bssTypes.BootParams
		if err := format.UnmarshalData(httpEnv.Body, &bp, format.DataFormatJson); err != nil {
			log.Logger.Error().Err(err).Msg("failed to unmarshal boot params")
			logHelpError(cmd)
			os.Exit(1)
		}

		// Edit parameters for nodes
		k := kargs.NewKargs([]byte(bp.Params))
		k.SetKarg("root", args[0])
		bp.Params = k.String()

		// Send modified params back to BSS
		_, err = bssClient.PutBootParams(bp, token)
		if err != nil {
			if errors.Is(err, client.UnsuccessfulHTTPError) {
				log.Logger.Error().Err(err).Msg("BSS boot parameter PUT request yielded unsuccessful HTTP response")
			} else {
				log.Logger.Error().Err(err).Msg("failed to set boot parameters in BSS")
			}
			logHelpError(cmd)
			os.Exit(1)
		}
	},
}

func init() {
	bootImageSetCmd.Flags().StringSliceP("xname", "x", []string{}, "one or more xnames whose boot parameters to set")
	bootImageSetCmd.Flags().StringSliceP("mac", "m", []string{}, "one or more MAC addresses whose boot parameters to set")
	bootImageSetCmd.Flags().Int32SliceP("nid", "n", []int32{}, "one or more node IDs whose boot parameters to set")

	bootImageSetCmd.MarkFlagsOneRequired("xname", "mac", "nid")

	bootImageCmd.AddCommand(bootImageSetCmd)
}
