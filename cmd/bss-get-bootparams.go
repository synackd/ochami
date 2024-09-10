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
	"os"

	"github.com/spf13/cobra"
	"github.com/synackd/ochami/internal/client"
	"github.com/synackd/ochami/internal/log"
)

// bssGetBootparamsCmd represents the bootparams command
var bssGetBootparamsCmd = &cobra.Command{
	Use:   "bootparams",
	Short: "Get boot parameters for one or all nodes",
	Run: func(cmd *cobra.Command, args []string) {
		// If base URI is empty, we cannot do anything
		if baseURI == "" {
			log.Logger.Error().Msg("no base URI set")
			if err := cmd.Usage(); err != nil {
				log.Logger.Error().Err(err).Msg("failed to print usage")
			}
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

		// If no args specified, get all boot parameters
		if len(args) == 0 {
			bssClient, err := client.NewBSSClient(baseURI)
			if err != nil {
				log.Logger.Error().Err(err).Msg("error creating new BSS client")
				os.Exit(1)
			}

			data, err := bssClient.GetData("/bootparameters", token, nil)
			if err != nil {
				if errors.Is(err, client.UnsuccessfulHTTPError) {
					log.Logger.Error().Err(err).Msg("BSS boot parameter request yielded unsuccessful HTTP response")
				} else {
					log.Logger.Error().Err(err).Msg("failed to request boot parameters from BSS")
				}
				os.Exit(1)
			}

			fmt.Println(data)
		}
	},
}

func init() {
	bssGetCmd.AddCommand(bssGetBootparamsCmd)
}
