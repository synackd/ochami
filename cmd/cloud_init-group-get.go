// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/OpenCHAMI/cloud-init/pkg/cistore"
	"github.com/spf13/cobra"

	"github.com/OpenCHAMI/ochami/internal/log"
	"github.com/OpenCHAMI/ochami/pkg/client"
	"github.com/OpenCHAMI/ochami/pkg/client/ci"
)

// cloudInitGetGroupData returns a slice of cloud-init group data for the
// requested groups. If an error occurs, the program exits.
func cloudInitGetGroupData(cmd *cobra.Command, args []string) (groupSlice []cistore.GroupData) {
	// Create client to use for requests
	cloudInitClient := cloudInitGetClient(cmd, true)

	// Get data
	if len(args) == 0 {
		// No args passed, get all group data at once
		henvs, errs, err := cloudInitClient.GetGroups(token)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to get all groups from cloud-init")
			logHelpError(cmd)
			os.Exit(1)
		}
		if errs[0] != nil {
			if errors.Is(errs[0], client.UnsuccessfulHTTPError) {
				log.Logger.Error().Err(errs[0]).Msg("cloud-init group request yielded unsuccessful HTTP response")
			} else {
				log.Logger.Error().Err(errs[0]).Msg("failed to cloud-init groups")
			}
			logHelpError(cmd)
			os.Exit(1)
		}

		// Group data is formatted as a map keyed on the name,
		// which is a bit awkward since the name appears twice
		// and is hard to iterate through.
		//
		// Convert group map into group slice.
		var groupMap map[string]cistore.GroupData
		if err := json.Unmarshal(henvs[0].Body, &groupMap); err != nil {
			log.Logger.Error().Err(err).Msg("failed to unmarshal all groups")
			logHelpError(cmd)
			os.Exit(1)
		}
		groupSlice = ci.CIGroupDataMapToSlice(groupMap)
	} else {
		// One or more arguments (group IDs) provided, get data
		// for just those groups.
		henvs, errs, err := cloudInitClient.GetGroups(token, args...)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to get cloud-init groups")
			logHelpError(cmd)
			os.Exit(1)
		}
		// Since the requests are done iteratively, we need to
		// deal with each error that might have occurred.
		var errorsOccurred = false
		for _, err := range errs {
			if err != nil {
				if errors.Is(err, client.UnsuccessfulHTTPError) {
					log.Logger.Error().Err(err).Msg("cloud-init group request yielded unsuccessful HTTP response")
				} else {
					log.Logger.Error().Err(err).Msg("failed to get cloud-init groups")
				}
				errorsOccurred = true
			}
		}
		if errorsOccurred {
			log.Logger.Warn().Msg("cloud-init group retrieval completed with errors")
			logHelpError(cmd)
			os.Exit(1)
		}

		// Collect group data into JSON array
		errorsOccurred = false
		for _, henv := range henvs {
			var ciGroup cistore.GroupData
			if err := json.Unmarshal(henv.Body, &ciGroup); err != nil {
				log.Logger.Error().Err(err).Msg("failed to unmarshal HTTP body into group")
				errorsOccurred = true
			} else {
				groupSlice = append(groupSlice, ciGroup)
			}
		}
		if errorsOccurred {
			log.Logger.Warn().Msg("not all group data was collected due to errors")
			logHelpError(cmd)
			os.Exit(1)
		}
	}
	return
}

// cloudInitGroupGetCmd represents the "cloud-init group get" command
var cloudInitGroupGetCmd = &cobra.Command{
	Use:   "get",
	Args:  cobra.NoArgs,
	Short: "Get group data for all or a subset of cloud-init groups",
	Long: `Get group data for all or a subset of cloud-init groups.

See ochami-cloud-init(1) for more details.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			printUsageHandleError(cmd)
			os.Exit(0)
		}
	},
}

// cloudInitGroupGetConfigCmd represents the "cloud-init group get config" command
var cloudInitGroupGetConfigCmd = &cobra.Command{
	Use:   "config [<group_name>...]",
	Short: "Get cloud-init config from cloud-init server for one or more groups",
	Long: `Get cloud-init config from cloud-init server for one or more groups.

See ochami-cloud-init(1) for more details.`,
	Example: `  # Get just the cloud-init configuration
  ochami cloud-init group get config
  ochami cloud-init group get config compute`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get all data for specified (or unspecified) groups
		groupSlice := cloudInitGetGroupData(cmd, args)

		// Extract cloud-config for each group
		type configGroup struct {
			Name     string                 `json:"name" yaml:"name"`
			Data     map[string]interface{} `json:"meta-data" yaml:"meta-data"`
			Content  []byte                 `json:"content" yaml:"content"`
			Encoding string                 `json:"encoding" enums:"base64,plain"`
		}
		var configSlice []configGroup
		for _, config := range groupSlice {
			if len(config.File.Content) == 0 {
				log.Logger.Warn().Msgf("cloud-config for group %s was empty, not printing", config.Name)
				continue
			}
			newCfg := configGroup{
				Name:     config.Name,
				Data:     config.Data,
				Content:  config.File.Content,
				Encoding: config.File.Encoding,
			}

			// Base64 decode any base64-decoded cloud configs
			ccf := cistore.CloudConfigFile{
				Content:  newCfg.Content,
				Encoding: newCfg.Encoding,
			}
			if cBytes, err := ci.DecodeCloudConfig(ccf); err != nil {
				log.Logger.Error().Err(err).Msgf("failed to decode cloud-config for %s", newCfg.Name)
				logHelpError(cmd)
				os.Exit(1)
			} else {
				newCfg.Content = cBytes
				newCfg.Encoding = "plain"
			}

			configSlice = append(configSlice, newCfg)
		}

		// Print cloud-init config(s)
		for cidx, cfg := range configSlice {
			if ciHeaderWhen == CIFlagHeaderNever {
				fmt.Println(string(configSlice[cidx].Content))
			} else if ciHeaderWhen == CIFlagHeaderAlways {
				fmt.Printf("--- (%d/%d) group=%s\n", cidx+1, len(configSlice), cfg.Name)
				fmt.Println(string(configSlice[cidx].Content))
				fmt.Println()
			} else {
				if len(configSlice) == 1 {
					fmt.Println(string(configSlice[cidx].Content))
				} else {
					fmt.Printf("--- (%d/%d) group=%s\n", cidx+1, len(configSlice), cfg.Name)
					fmt.Println(string(configSlice[cidx].Content))
				}
			}
		}
	},
}

// cloudInitGroupGetMetaDataCmd represents the "cloud-init group get meta-data" command
var cloudInitGroupGetMetadataCmd = &cobra.Command{
	Use:   "meta-data [<group_name>...]",
	Short: "Get meta-data from cloud-init server for one or more groups",
	Long: `Get meta-data from cloud-init server for one or more groups.

See ochami-cloud-init(1) for more details.`,
	Example: `  # Get just the meta-data
  ochami cloud-init group get meta-data
  ochami cloud-init group get meta-data compute`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get all data for specified (or unspecified) groups
		groupSlice := cloudInitGetGroupData(cmd, args)

		// Extract meta-data for each group
		type mdGroup struct {
			Name string                 `json:"name" yaml:"name"`
			Data map[string]interface{} `json:"meta-data" yaml:"meta-data"`
		}
		var mdSlice []mdGroup
		for _, group := range groupSlice {
			newGr := mdGroup{
				Name: group.Name,
				Data: group.Data,
			}
			mdSlice = append(mdSlice, newGr)
		}

		// Marshal data into JSON so it can be reformatted into
		// desired output format.
		groupSliceBytes, err := json.Marshal(mdSlice)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to marshal group list into JSON")
			logHelpError(cmd)
			os.Exit(1)
		}

		// Print in desired format
		if outBytes, err := client.FormatBody(groupSliceBytes, formatOutput); err != nil {
			log.Logger.Error().Err(err).Msg("failed to format output")
			logHelpError(cmd)
			os.Exit(1)
		} else {
			fmt.Print(string(outBytes))
		}
	},
}

// cloudInitGroupGetRawCmd represents the "cloud-init group get raw" command
var cloudInitGroupGetRawCmd = &cobra.Command{
	Use:   "raw [<group_name>...]",
	Short: "Get raw data from cloud-init server for one or more groups",
	Long: `Get raw data from cloud-init server for one or more groups.

See ochami-cloud-init(1) for more details.`,
	Example: `  # Get raw information about group from cloud-init server
  ochami cloud-init group get raw
  ochami cloud-init group get raw compute`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get all data for specified (or unspecified) groups
		groupSlice := cloudInitGetGroupData(cmd, args)

		// Marshal data into JSON so it can be reformatted into
		// desired output format.
		groupSliceBytes, err := json.Marshal(groupSlice)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to marshal group list into JSON")
			logHelpError(cmd)
			os.Exit(1)
		}

		// Print in desired format
		if outBytes, err := client.FormatBody(groupSliceBytes, formatOutput); err != nil {
			log.Logger.Error().Err(err).Msg("failed to format output")
			logHelpError(cmd)
			os.Exit(1)
		} else {
			fmt.Print(string(outBytes))
		}
	},
}

func init() {
	// Add config subcommand
	cloudInitGroupGetConfigCmd.Flags().Var(&ciHeaderWhen, "headers", "when to print headers above cloud-configs (always,multiple,never")
	cloudInitGroupGetConfigCmd.RegisterFlagCompletionFunc("headers", cloudInitCompletionHeaderWhen)
	cloudInitGroupGetCmd.AddCommand(cloudInitGroupGetConfigCmd)

	// Add meta-data subcommand
	cloudInitGroupGetMetadataCmd.PersistentFlags().VarP(&formatOutput, "format-output", "F", "format of output printed to standard output")
	cloudInitGroupGetMetadataCmd.RegisterFlagCompletionFunc("format-output", completionFormatData)
	cloudInitGroupGetCmd.AddCommand(cloudInitGroupGetMetadataCmd)

	// Add raw subcommand
	cloudInitGroupGetRawCmd.PersistentFlags().VarP(&formatOutput, "format-output", "F", "format of output printed to standard output")
	cloudInitGroupGetRawCmd.RegisterFlagCompletionFunc("format-output", completionFormatData)
	cloudInitGroupGetCmd.AddCommand(cloudInitGroupGetRawCmd)

	// Add get command
	cloudInitGroupCmd.AddCommand(cloudInitGroupGetCmd)
}
