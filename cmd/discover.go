// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/OpenCHAMI/ochami/internal/client"
	"github.com/OpenCHAMI/ochami/internal/discover"
	"github.com/OpenCHAMI/ochami/internal/log"
)

// discoverCmd represents the discover command
var discoverCmd = &cobra.Command{
	Use:   "discover -f <payload_file> [--payload-format <format>]",
	Args:  cobra.NoArgs,
	Short: "Populate SMD with data",
	Long: `Populate SMD with data. Currently, this command performs "fake" discovery,
whereby data from a payload file is used to create the SMD structures.
In this way, the command does not perform dynamic discovery like Magellan,
but statically populates SMD using a file.

The format of the payload file is an array of node specifications. In YAML,
each node entry would look something like:

- name: node01
  nid: 1
  xname: x1000c1s7b0n0
  bmc_mac: de:ca:fc:0f:ee:ee
  bmc_ip: 172.16.0.101
  group: compute
  interfaces:
  - mac_addr: de:ad:be:ee:ee:f1
    ip_addrs:
    - name: internal
      ip_addr: 172.16.0.1
  - mac_addr: de:ad:be:ee:ee:f2
    ip_addrs:
    - name: external
      ip_addr: 10.15.3.100
  - mac_addr: 02:00:00:91:31:b3
    ip_addrs:
    - name: HSN
      ip_addr: 192.168.0.1

`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check that all required args are passed
		if len(args) == 0 && !cmd.Flag("payload").Changed {
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

		// Read data from payload file
		nodes := discover.NodeList{}
		dFile := cmd.Flag("payload").Value.String()
		dFormat := cmd.Flag("payload-format").Value.String()
		err = client.ReadPayload(dFile, dFormat, &nodes)
		if err != nil {
			log.Logger.Error().Err(err).Msg("unable to read payload for request")
			os.Exit(1)
		}
		log.Logger.Debug().Msgf("read %d nodes", len(nodes))
		log.Logger.Debug().Msgf("nodes: %s", nodes)

		// Put together payload for different endpoints
		log.Logger.Debug().Msg("generating redfish structures to send to SMD")
		rfes, ifaces, err := discover.DiscoveryInfoV2(smdBaseURI, nodes)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to construct structures to send to SMD")
			os.Exit(1)
		}
		log.Logger.Debug().Msgf("generated redfish structures: %v", rfes.RedfishEndpoints)

		// Send RedfishEndpoint requests
		rfeErrorsOccurred := false
		_, errs, err := smdClient.PostRedfishEndpointsV2(rfes, token)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to add redfish endpoints to SMD")
			rfeErrorsOccurred = true
		}
		for _, err := range errs {
			if err != nil {
				var errMsg string
				if errors.Is(err, client.UnsuccessfulHTTPError) {
					errMsg = "SMD redfish endpoint request yielded unsuccessful HTTP response"
				} else {
					errMsg = "failed to add redfish endpoint to SMD"
				}
				log.Logger.Error().Err(err).Msg(errMsg)
				rfeErrorsOccurred = true
			}
		}

		// Send EthernetInterface requests
		ifaceErrorsOccurred := false
		_, errs, err = smdClient.PostEthernetInterfaces(ifaces, token)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to add ethernet interfaces to SMD")
			ifaceErrorsOccurred = true
		}
		for _, err := range errs {
			if err != nil {
				var errMsg string
				if errors.Is(err, client.UnsuccessfulHTTPError) {
					errMsg = "SMD ethernet interface request yielded unsuccessful HTTP response"
				} else {
					errMsg = "failed to add ethernet interface to SMD"
				}
				log.Logger.Error().Err(err).Msg(errMsg)
				ifaceErrorsOccurred = true
			}
		}

		// Put together list of groups to add and which components to add to those groups
		groupsToAdd := make(map[string]client.Group)
		for _, node := range nodes {
			if node.Group != "" {
				if g, ok := groupsToAdd[node.Group]; !ok {
					newGroup := client.Group{
						Label:       node.Group,
						Description: fmt.Sprintf("The %s group", node.Group),
					}
					newGroup.Members.IDs = []string{node.Xname}
					groupsToAdd[node.Group] = newGroup
				} else {
					g.Members.IDs = append(g.Members.IDs, node.Xname)
				}
			}
		}
		groupList := make([]client.Group, len(groupsToAdd))
		var idx = 0
		for _, g := range groupsToAdd {
			groupList[idx] = g
			idx++
		}

		// Add groups and components to those groups
		groupErrorsOccurred := false
		_, errs, err = smdClient.PostGroups(groupList, token)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to add groups to SMD")
			groupErrorsOccurred = true
		}
		for _, err := range errs {
			if err != nil {
				var errMsg string
				if errors.Is(err, client.UnsuccessfulHTTPError) {
					errMsg = "SMD groups request yielded unsuccessful HTTP response"
				} else {
					errMsg = "failed to add groups to SMD"
				}
				log.Logger.Error().Err(err).Msg(errMsg)
				groupErrorsOccurred = true
			}
		}

		// Notify user if any request errors occurred
		exitStatus := 0
		if rfeErrorsOccurred {
			log.Logger.Warn().Msg("redfish endpoint requests completed with errors")
			exitStatus = 1
		}
		if ifaceErrorsOccurred {
			log.Logger.Warn().Msg("ethernet interface requests completed with errors")
			exitStatus = 1
		}
		if groupErrorsOccurred {
			log.Logger.Warn().Msg("group requests completed with errors")
			exitStatus = 1
		}
		os.Exit(exitStatus)
	},
}

func init() {
	discoverCmd.Flags().StringP("payload", "f", "", "file containing the request payload; JSON format unless --payload-format specified")
	discoverCmd.Flags().String("payload-format", defaultPayloadFormat, "format of payload file (yaml,json) passed with --payload")

	discoverCmd.MarkFlagRequired("payload")

	rootCmd.AddCommand(discoverCmd)
}
