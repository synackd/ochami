// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"
	"github.com/OpenCHAMI/ochami/internal/client"
	"github.com/OpenCHAMI/ochami/internal/log"
)

// groupGetCmd represents the smd-group-get command
var groupGetCmd = &cobra.Command{
	Use:   "get",
	Args:  cobra.NoArgs,
	Short: "Get all groups or group(s) identified by name and/or tag",
	Example: `  ochami smd group get
  ochami smd group get --name group1
  ochami smd group get --tag group1_tag
  ochami smd group get --name group1,group2
  ochami smd group get --name group1 --name group2
  ochami smd group get --name group1,group2 --tag tag1,tag2
  ochami smd group get --name group1 --name group2 --tag tag1 --tag tag2`,
	Run: func(cmd *cobra.Command, args []string) {
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
		smdClient, err := client.NewSMDClient(smdBaseURI, insecure)
		if err != nil {
			log.Logger.Error().Err(err).Msg("error creating new SMD client")
			os.Exit(1)
		}

		// Check if a CA certificate was passed and load it into client if valid
		useCACert(smdClient.OchamiClient)

		// If no ID flags are specified, get all groups
		qstr := ""
		if cmd.Flag("name").Changed || cmd.Flag("tag").Changed {
			values := url.Values{}
			if cmd.Flag("name").Changed {
				s, err := cmd.Flags().GetStringSlice("name")
				if err != nil {
					log.Logger.Error().Err(err).Msg("unable to fetch name list")
					os.Exit(1)
				}
				for _, n := range s {
					values.Add("group", n)
				}
			}
			if cmd.Flag("tag").Changed {
				s, err := cmd.Flags().GetStringSlice("tag")
				if err != nil {
					log.Logger.Error().Err(err).Msg("unable to fetch tag list")
					os.Exit(1)
				}
				for _, t := range s {
					values.Add("tag", t)
				}
			}
			qstr = values.Encode()
		}
		httpEnv, err := smdClient.GetGroups(qstr, token)
		if err != nil {
			if errors.Is(err, client.UnsuccessfulHTTPError) {
				log.Logger.Error().Err(err).Msg("SMD group request yielded unsuccessful HTTP response")
			} else {
				log.Logger.Error().Err(err).Msg("failed to request groups from SMD")
			}
			os.Exit(1)
		}
		fmt.Println(string(httpEnv.Body))
	},
}

func init() {
	groupGetCmd.Flags().StringSlice("name", []string{}, "filter groups by name")
	groupGetCmd.Flags().StringSlice("tag", []string{}, "filter groups by tag")
	groupCmd.AddCommand(groupGetCmd)
}
