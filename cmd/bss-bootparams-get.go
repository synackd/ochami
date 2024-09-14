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

// bssBootparamsGetCmd represents the bootparams command
var bssBootparamsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get boot parameters for one or all nodes",
	Example: `  ochami bss bootparams get
  ochami bss bootparams get --mac 00:de:ad:be:ef:00
  ochami bss bootparams get --mac 00:de:ad:be:ef:00,00:c0:ff:ee:00:00
  ochami bss bootparams get --mac 00:de:ad:be:ef:00 --mac 00:c0:ff:ee:00:00`,
	Run: func(cmd *cobra.Command, args []string) {
		// Without a base URI, we cannot do anything
		bssBaseURI, err := getBaseURI(cmd)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to get base URI for BSS")
			os.Exit(1)
		}

		// This endpoint requires authentication, so a token is needed
		// TODO: Check token validity/expiration
		if token == "" {
			log.Logger.Error().Msg("no token set")
			if err := cmd.Usage(); err != nil {
				log.Logger.Error().Err(err).Msg("failed to print usage")
			}
			os.Exit(1)
		}

		// Create client to make request to BSS
		bssClient, err := client.NewBSSClient(bssBaseURI, insecure)
		if err != nil {
			log.Logger.Error().Err(err).Msg("error creating new BSS client")
			os.Exit(1)
		}

		// Check if a CA certificate was passed and load it into client if valid
		if cacertPath != "" {
			log.Logger.Debug().Msgf("Attempting to use CA certificate at %s", cacertPath)
			err = bssClient.UseCACert(cacertPath)
			if err != nil {
				log.Logger.Error().Err(err).Msgf("failed to load CA certificate %s: %v", cacertPath)
				os.Exit(1)
			}
		}

		// If no ID flags are specified, get all boot parameters
		qstr := ""
		if cmd.Flag("xname").Changed ||
			cmd.Flag("mac").Changed ||
			cmd.Flag("nid").Changed {
			values := url.Values{}
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
				s, err := cmd.Flags().GetIntSlice("nid")
				if err != nil {
					log.Logger.Error().Err(err).Msg("unable to fetch nid list")
					os.Exit(1)
				}
				for _, n := range s {
					values.Add("nid", fmt.Sprintf("%d", n))
				}
			}
			qstr = values.Encode()
		}
		httpEnv, err := bssClient.GetBootParams(qstr, token)
		if err != nil {
			if errors.Is(err, client.UnsuccessfulHTTPError) {
				log.Logger.Error().Err(err).Msg("BSS boot parameter request yielded unsuccessful HTTP response")
			} else {
				log.Logger.Error().Err(err).Msg("failed to request boot parameters from BSS")
			}
			os.Exit(1)
		}
		fmt.Println(string(httpEnv.Body))
	},
}

func init() {
	bssBootparamsGetCmd.Flags().StringSliceP("xname", "x", []string{}, "one or more xnames whose boot parameters to get")
	bssBootparamsGetCmd.Flags().StringSliceP("mac", "m", []string{}, "one or more MAC addresses whose boot parameters to get")
	bssBootparamsGetCmd.Flags().IntSliceP("nid", "n", []int{}, "one or more node IDs whose boot parameters to get")
	bssBootparamsCmd.AddCommand(bssBootparamsGetCmd)
}
