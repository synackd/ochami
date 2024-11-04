// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"github.com/synackd/ochami/internal/client"
	"github.com/synackd/ochami/internal/log"
)

// groupUpdateCmd represents the update command
var groupUpdateCmd = &cobra.Command{
	Use:   "update [-f <payload_file>] | ([--description <description>] [--tag <tag>]... <group_label>)",
	Short: "Update the description and/or tags of a group",
	Long: `Update the description and/or tags of a group. At least one of --description
or --tag must be specified. Alternatively, pass -f to pass a file
(optionally specifying --payload-format, JSON by default), but the
rules above still apply for the payload.

This command sends a PATCH to SMD. An access token is required.`,
	Example: `  ochami group update --description "New description for compute" compute
  ochami group update --tag existing_tag --tag new_tag compute
  ochami group update --tag existing_tag,new_tag compute
  ochami group update --tag existing_tag,new_tag -d "New description for compute" compute
  ochami group update -f payload.json
  ochami group update -f payload.yaml --payload-format yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		// cmd.LocalFlags().NFlag() doesn't seem to work, so we check every flag
		if len(args) == 0 && !cmd.Flag("description").Changed && !cmd.Flag("tag").Changed {
			err := cmd.Usage()
			if err != nil {
				log.Logger.Error().Err(err).Msg("failed to print usage")
				os.Exit(1)
			}
			os.Exit(0)
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
		smdClient, err := client.NewSMDClient(smdBaseURI, insecure)
		if err != nil {
			log.Logger.Error().Err(err).Msg("error creating new SMD client")
			os.Exit(1)
		}

		// Check if a CA certificate was passed and load it into client if valid
		useCACert(smdClient.OchamiClient)

		// The group list we will send
		var groups []client.Group

		// Read payload from file first, allowing overwrites from flags
		if cmd.Flag("payload").Changed {
			dFile := cmd.Flag("payload").Value.String()
			dFormat := cmd.Flag("payload-format").Value.String()
			err := client.ReadPayload(dFile, dFormat, &groups)
			if err != nil {
				log.Logger.Error().Err(err).Msg("unable to read payload for request")
				os.Exit(1)
			}
		} else {
			// ...otherwise use CLI options/args
			group := client.Group{Label: args[0]}
			if cmd.Flag("description").Changed {
				if group.Description, err = cmd.Flags().GetString("description"); err != nil {
					log.Logger.Error().Err(err).Msg("unable to fetch description")
					os.Exit(1)
				}
			}
			if cmd.Flag("tag").Changed {
				if group.Tags, err = cmd.Flags().GetStringSlice("tag"); err != nil {
					log.Logger.Error().Err(err).Msg("unable to fetch tags")
					os.Exit(1)
				}
			}
			groups = append(groups, group)
		}

		// Send 'em off
		_, errs, err := smdClient.PatchGroups(groups, token)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to patch group in SMD")
			os.Exit(1)
		}
		// Since smdClient.PatchGroups does the edition iteratively, we need to deal with
		// each error that might have occurred.
		var errorsOccurred = false
		for _, err := range errs {
			if err != nil {
				if errors.Is(err, client.UnsuccessfulHTTPError) {
					log.Logger.Error().Err(err).Msg("SMD group request yielded unsuccessful HTTP response")
				} else {
					log.Logger.Error().Err(err).Msg("failed to update group(s) to SMD")
				}
				errorsOccurred = true
			}
		}
		if errorsOccurred {
			log.Logger.Warn().Msg("SMD group update completed with errors")
			os.Exit(1)
		}
	},
}

func init() {
	groupUpdateCmd.Flags().StringP("description", "d", "", "short description to update group with")
	groupUpdateCmd.Flags().StringSlice("tag", []string{}, "one or more tags to set for group")
	groupUpdateCmd.Flags().StringP("payload", "f", "", "file containing the request payload; JSON format unless --payload-format specified")
	groupUpdateCmd.Flags().String("payload-format", defaultPayloadFormat, "format of payload file (yaml,json) passed with --payload")

	groupUpdateCmd.MarkFlagsOneRequired("description", "tag", "payload")

	groupCmd.AddCommand(groupUpdateCmd)
}