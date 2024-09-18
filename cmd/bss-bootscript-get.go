// Copyright © 2024 Triad National Security, LLC. All rights reserved.
//
// This program was produced under U.S. Government contract 89233218CNA000001
// for Los Alamos National Laboratory (LANL), which is operated by Triad
// National Security, LLC for the U.S. Department of Energy/National Nuclear
// Security Administration. All rights in the program are reserved by Triad
// National Security, LLC, and the U.S. Department of Energy/National Nuclear
// Security Administration. The Government is granted for itself and others
// acting on its behalf a nonexclusive, paid-up, irrevocable worldwide license
// in this material to reproduce, prepare derivative works, distribute copies to
// the public, perform publicly and display publicly, and to permit others to do
// so.
package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"
	"github.com/synackd/ochami/internal/client"
	"github.com/synackd/ochami/internal/log"
)

// bssBootScriptGetCmd represents the get command
var bssBootScriptGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get iPXE boot script for a node",
	Run: func(cmd *cobra.Command, args []string) {
		// Without a base URI, we cannot do anything
		bssBaseURI, err := getBaseURI(cmd)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to get base URI for BSS")
			os.Exit(1)
		}

		// Create client to make request to BSS
		bssClient, err := client.NewBSSClient(bssBaseURI, insecure)
		if err != nil {
			log.Logger.Error().Err(err).Msg("error creating new BSS client")
			os.Exit(1)
		}

		// Check if a CA certificate was passed and load it into client if valid
		useCACert(bssClient.OchamiClient)

		// Structure representing the boot script query string
		values := url.Values{}

		// At least one of these required
		if cmd.Flag("xname").Changed {
			s, err := cmd.Flags().GetStringSlice("xname")
			if err != nil {
				log.Logger.Error().Err(err).Msg("unable to fetch xname list")
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
				os.Exit(1)
			}
			for _, n := range s {
				values.Add("nid", fmt.Sprintf("%d", n))
			}
		}

		// These are optional
		if cmd.Flag("retry").Changed {
			s, err := cmd.Flags().GetInt("retry")
			if err != nil {
				log.Logger.Error().Err(err).Msg("unable to fetch number of retries")
				os.Exit(1)
			}
			values.Add("retry", fmt.Sprintf("%d", s))
		}
		if cmd.Flag("arch").Changed {
			s, err := cmd.Flags().GetString("arch")
			if err != nil {
				log.Logger.Error().Err(err).Msg("unable to fetch arch")
				os.Exit(1)
			}
			values.Add("arch", s)
		}
		if cmd.Flag("timestamp").Changed {
			s, err := cmd.Flags().GetInt("timestamp")
			if err != nil {
				log.Logger.Error().Err(err).Msg("unable to fetch timestamp")
				os.Exit(1)
			}
			values.Add("timestamp", fmt.Sprintf("%d", s))
		}
		qstr := values.Encode()

		httpEnv, err := bssClient.GetBootScript(qstr)
		if err != nil {
			if errors.Is(err, client.UnsuccessfulHTTPError) {
				log.Logger.Error().Err(err).Msg("BSS boot script request yielded unsuccessful HTTP response")
			} else {
				log.Logger.Error().Err(err).Msg("failed to request boot script from BSS")
			}
			os.Exit(1)
		}
		fmt.Println(string(httpEnv.Body))
	},
}

func init() {
	bssBootScriptGetCmd.Flags().StringSliceP("xname", "x", []string{}, "one or more xnames whose boot script to get")
	bssBootScriptGetCmd.Flags().StringSliceP("mac", "m", []string{}, "one or more MAC addresses whose boot script to get")
	bssBootScriptGetCmd.Flags().Int32SliceP("nid", "n", []int32{}, "one or more node IDs whose boot script to get")
	bssBootScriptGetCmd.Flags().Int("retry", 0, "number of times to retry fetching boot script on failed boot")
	bssBootScriptGetCmd.Flags().String("arch", "", "architecture value from iPXE variable ${buildarch}")
	bssBootScriptGetCmd.Flags().Int("timestamp", 0, "timestamp in seconds since Unix epoch for when SMD state needs to be updated by")

	bssBootScriptGetCmd.MarkFlagsOneRequired("xname", "mac", "nid")

	bssBootScriptCmd.AddCommand(bssBootScriptGetCmd)
}
