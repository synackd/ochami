// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"

	"github.com/OpenCHAMI/ochami/internal/log"
	"github.com/OpenCHAMI/ochami/pkg/client"
)

// groupGetCmd represents the smd-group-get command
var groupGetCmd = &cobra.Command{
	Use:   "get",
	Args:  cobra.NoArgs,
	Short: "Get all groups or group(s) identified by name and/or tag",
	Long: `Get all groups or group(s) identified by name and/or tag.

See ochami-smd(1) for more details.`,
	Example: `  ochami smd group get
  ochami smd group get --name group1
  ochami smd group get --tag group1_tag
  ochami smd group get --name group1,group2
  ochami smd group get --name group1 --name group2
  ochami smd group get --name group1,group2 --tag tag1,tag2
  ochami smd group get --name group1 --name group2 --tag tag1 --tag tag2`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create client to use for requests
		smdClient := smdGetClient(cmd, true)

		// If no ID flags are specified, get all groups
		qstr := ""
		if cmd.Flag("name").Changed || cmd.Flag("tag").Changed {
			values := url.Values{}
			if cmd.Flag("name").Changed {
				s, err := cmd.Flags().GetStringSlice("name")
				if err != nil {
					log.Logger.Error().Err(err).Msg("unable to fetch name list")
					logHelpError(cmd)
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
					logHelpError(cmd)
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
			logHelpError(cmd)
			os.Exit(1)
		}

		// Print output
		if outBytes, err := client.FormatBody(httpEnv.Body, formatOutput); err != nil {
			log.Logger.Error().Err(err).Msg("failed to format output")
			logHelpError(cmd)
			os.Exit(1)
		} else {
			fmt.Print(string(outBytes))
		}
	},
}

func init() {
	groupGetCmd.Flags().StringSlice("name", []string{}, "filter groups by name")
	groupGetCmd.Flags().StringSlice("tag", []string{}, "filter groups by tag")
	groupGetCmd.Flags().VarP(&formatOutput, "format-output", "F", "format of output printed to standard output (json,json-pretty,yaml)")

	groupGetCmd.RegisterFlagCompletionFunc("format-output", completionFormatData)

	groupCmd.AddCommand(groupGetCmd)
}
