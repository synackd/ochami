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

// bssHistoryCmd represents the bss-history command
var bssHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "Fetch the endpoint history of BSS",
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

		// If no ID flags are specified, get all boot parameters
		qstr := ""
		if cmd.Flag("xname").Changed || cmd.Flag("endpoint").Changed {
			values := url.Values{}
			if cmd.Flag("xname").Changed {
				x, err := cmd.Flags().GetString("xname")
				if err != nil {
					log.Logger.Error().Err(err).Msg("unable to fetch xname")
					os.Exit(1)
				}
				values.Add("name", x)
			}
			if cmd.Flag("endpoint").Changed {
				e, err := cmd.Flags().GetString("endpoint")
				if err != nil {
					log.Logger.Error().Err(err).Msg("unable to fetch endpoint")
					os.Exit(1)
				}
				values.Add("endpoint", e)
			}
			qstr = values.Encode()
		}

		// Send request
		httpEnv, err := bssClient.GetEndpointHistory(qstr)
		if err != nil {
			if errors.Is(err, client.UnsuccessfulHTTPError) {
				log.Logger.Error().Err(err).Msg("BSS endpoint history request yielded unsuccessful HTTP response")
			} else {
				log.Logger.Error().Err(err).Msg("failed to request endpoint history from BSS")
			}
			os.Exit(1)
		}
		fmt.Println(string(httpEnv.Body))
	},
}

func init() {
	bssHistoryCmd.Flags().String("xname", "", "filter by xname")
	bssHistoryCmd.Flags().String("endpoint", "", "filter by endpoint")
	bssCmd.AddCommand(bssHistoryCmd)
}
