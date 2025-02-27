// This source code is licensed under the license found in the LICENSE file at
// the root directory of this source tree.
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/OpenCHAMI/ochami/internal/log"
	"github.com/OpenCHAMI/ochami/pkg/client"
	"github.com/OpenCHAMI/ochami/pkg/client/smd"
	"github.com/OpenCHAMI/ochami/pkg/discover"
	"github.com/spf13/cobra"
)

// discoverCmd represents the discover command
var discoverCmd = &cobra.Command{
	Use:   "discover -f <payload_file> [--payload-format <format>] [--overwrite]",
	Args:  cobra.NoArgs,
	Short: "Populate SMD with data",
	Long: `Populate SMD with data. Currently, this command performs "fake" discovery,
whereby data from a payload file is used to create the SMD structures.
In this way, the command does not perform dynamic discovery like Magellan,
but statically populates SMD using a file. If - is used as the argument to
-f, the payload data is read from standard input.

The format of the payload file is an array of node specifications. In YAML,
each node entry would look something like:

nodes:
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
		smdClient, err := smd.NewClient(smdBaseURI, insecure)
		if err != nil {
			log.Logger.Error().Err(err).Msg("error creating new SMD client")
			os.Exit(1)
		}

		// Check if a CA certificate was passed and load it into client if valid
		useCACert(smdClient.OchamiClient)

		if cmd.Flag("overwrite").Changed {
			log.Logger.Warn().Msg("--overwrite passed; overwriting any existing data")
		}

		// Read data from payload file
		nodes := discover.NodeList{}
		handlePayload(cmd, &nodes)
		log.Logger.Debug().Msgf("read %d nodes", len(nodes.Nodes))
		log.Logger.Debug().Msgf("nodes: %s", nodes)

		// Put together payload for different endpoints
		log.Logger.Debug().Msg("generating redfish structures to send to SMD")
		comps, rfes, ifaces, err := discover.DiscoveryInfoV2(smdBaseURI, nodes)
		if err != nil {
			log.Logger.Error().Err(err).Msg("failed to construct structures to send to SMD")
			os.Exit(1)
		}
		log.Logger.Debug().Msgf("generated redfish structures: %v", rfes.RedfishEndpoints)

		// Send Component requests
		// NOTE: These are sent *before* the RedfishEndpoints so the
		// user-specified NIDs get used instead of the SMD-generated
		// ones. The NIDs generated by SMD assume starting at 1 and
		// increment up in the order added.
		compErrorsOccurred := false
		if cmd.Flag("overwrite").Changed {
			// Send a PUT if --overwrite specified to overwrite any existing components
			_, errs, err := smdClient.PutComponents(comps, token)
			if err != nil {
				log.Logger.Error().Err(err).Msg("failed to add/overwrite components in SMD")
				compErrorsOccurred = true
			}
			for _, err := range errs {
				if err != nil {
					var errMsg string
					if errors.Is(err, client.UnsuccessfulHTTPError) {
						errMsg = "SMD component request yielded unsuccessful HTTP response"
					} else {
						errMsg = "failed to add/overwrite component in SMD"
					}
					log.Logger.Error().Err(err).Msg(errMsg)
					compErrorsOccurred = true
				}
			}

			// The SMD Components API does not modify the NID for
			// PUTs. Thus, we explicitly do it with a PATCH to a
			// specific endpoint that does it.
			if _, err := smdClient.PatchComponentsNID(comps, token); err != nil {
				log.Logger.Error().Err(err).Msg("failed to update NIDs for components in SMD")
				compErrorsOccurred = true
			}
		} else {
			// Otherwise send a normal POST
			_, err = smdClient.PostComponents(comps, token)
			if err != nil {
				var errMsg string
				if errors.Is(err, client.UnsuccessfulHTTPError) {
					errMsg = "SMD component request yielded unsuccessful HTTP response"
				} else {
					errMsg = "failed to add components to SMD"
				}
				log.Logger.Error().Err(err).Msg(errMsg)
				compErrorsOccurred = true
			}
		}

		// Send RedfishEndpoint requests
		var (
			rfeErrorsOccurred bool = false
			rfeHenvs          []client.HTTPEnvelope
			rfeErrs           []error
			rfeErr            error
		)
		if cmd.Flag("overwrite").Changed {
			// SMD's RedfishEndpoint API for PUT behaves more like
			// PATCH. In other words, the RedfishEndpoint must exist
			// _first_ before PUTting. This means that, to get
			// normal PUT behavior, we have to first try to POST,
			// then, if 409 is returned, try to PUT.
			for _, rfe := range rfes.RedfishEndpoints {
				// Attempt to POST the redfish endpoint
				rfeListWrapper := smd.RedfishEndpointSliceV2{
					RedfishEndpoints: []smd.RedfishEndpointV2{rfe},
				}
				rfeHenvs, rfeErrs, rfeErr = smdClient.PostRedfishEndpointsV2(rfeListWrapper, token)

				if rfeErr != nil {
					// An error in the function occurred,
					// err for this redfish endpoint and
					// move on.
					log.Logger.Error().Err(rfeErr).Msg("failed to add redfish endpoint to SMD")
					rfeErrorsOccurred = true
					continue
				}

				if rfeErrs[0] != nil {
					// An HTTP error occurred
					if errors.Is(rfeErrs[0], client.UnsuccessfulHTTPError) {
						if rfeHenvs[0].StatusCode == 409 {
							// RFE exists, PUT it
							log.Logger.Info().Msgf("redfish endpoint %s exists, attempting to update it", rfe.ID)
							_, putErrs, putErr := smdClient.PutRedfishEndpointsV2(rfeListWrapper, token)
							if putErr != nil {
								log.Logger.Error().Err(putErr).Msg("failed to update existing redfish endpoint in SMD")
								rfeErrorsOccurred = true
								continue
							}
							if putErrs[0] != nil {
								var errMsg string
								if errors.Is(putErrs[0], client.UnsuccessfulHTTPError) {
									errMsg = "SMD redfish endpoint PUT yielded unsuccessful HTTP response"
								} else {
									errMsg = "failed to update existing redfish endpoint in SMD"
								}
								log.Logger.Error().Err(putErrs[0]).Msg(errMsg)
								rfeErrorsOccurred = true
								continue
							}
						} else {
							// Some other HTTP error occurred, err
							log.Logger.Error().Err(rfeErrs[0]).Msg("SMD redfish endpoint POST yielded non-409 (duplicate) failure")
							rfeErrorsOccurred = true
							continue
						}
					} else {
						log.Logger.Error().Err(rfeErrs[0]).Msg("failed to add redfish endpoint to SMD")
						rfeErrorsOccurred = true
						continue
					}
				}
			}
		} else {
			// --overwrite was not passed, perform regular POST.
			_, rfeErrs, rfeErr = smdClient.PostRedfishEndpointsV2(rfes, token)
			if rfeErr != nil {
				log.Logger.Error().Err(rfeErr).Msg("failed to add redfish endpoints to SMD")
				rfeErrorsOccurred = true
			}
			for _, err := range rfeErrs {
				if err != nil {
					var errMsg string
					if errors.Is(err, client.UnsuccessfulHTTPError) {
						errMsg = "SMD redfish endpoint request yielded unsuccessful HTTP response"
					} else {
						if cmd.Flag("overwrite").Changed {
							errMsg = "failed to add/overwrite redfish endpoint in SMD"
						} else {
							errMsg = "failed to add redfish endpoint to SMD"
						}
					}
					log.Logger.Error().Err(err).Msg(errMsg)
					rfeErrorsOccurred = true
				}
			}
		}

		// Send EthernetInterface requests
		var (
			ifaceErrorsOccurred bool = false
			ifaceHenvs          []client.HTTPEnvelope
			ifaceErrs           []error
			ifaceErr            error
		)
		if cmd.Flag("overwrite").Changed {
			// SMD's EthernetInterface API does not allow the PUT
			// method. Instead, we loop over each ethernet interface
			// to add and attempt a POST. Iff a 409 is returned for
			// that interface, a PATCH is attempted. Otherwise, an
			// error has occurred.
			for _, iface := range ifaces {
				// Attempt to POST the ethernet interface
				ifaceListWrapper := []smd.EthernetInterface{iface}
				ifaceHenvs, ifaceErrs, ifaceErr = smdClient.PostEthernetInterfaces(ifaceListWrapper, token)

				if ifaceErr != nil {
					// An error in the function occurred, err for
					// this interface and move on.
					log.Logger.Error().Err(ifaceErr).Msg("failed to add ethernet interface to SMD")
					ifaceErrorsOccurred = true
					continue
				}

				if ifaceErrs[0] != nil {
					// An HTTP error occurred
					if errors.Is(ifaceErrs[0], client.UnsuccessfulHTTPError) {
						if ifaceHenvs[0].StatusCode == 409 {
							// Ethernet interface exists, patch it
							log.Logger.Info().Msgf("ethernet interface with MAC address %s exists, attempting to update it", iface.MACAddress)
							_, patchErrs, patchErr := smdClient.PatchEthernetInterfaces(ifaceListWrapper, token)
							if patchErr != nil {
								log.Logger.Error().Err(patchErr).Msg("failed to update existing ethernet interface in SMD")
								ifaceErrorsOccurred = true
								continue
							}
							if patchErrs[0] != nil {
								var errMsg string
								if errors.Is(patchErrs[0], client.UnsuccessfulHTTPError) {
									errMsg = "SMD ethernet interface PATCH yielded unsuccessful HTTP response"
								} else {
									errMsg = "failed to update existing ethernet interface in SMD"
								}
								log.Logger.Error().Err(patchErrs[0]).Msg(errMsg)
								ifaceErrorsOccurred = true
								continue
							}
						} else {
							// Some other HTTP error occurred, err
							log.Logger.Error().Err(ifaceErrs[0]).Msg("SMD ethernet interface POST yield non-409 (duplicate) failure")
							ifaceErrorsOccurred = true
							continue
						}
					} else {
						log.Logger.Error().Err(ifaceErrs[0]).Msg("failed to add ethernet interface to SMD")
						ifaceErrorsOccurred = true
						continue
					}
				}
			}
		} else {
			// --overwrite was not passed, perform regular POST.
			_, ifaceErrs, ifaceErr = smdClient.PostEthernetInterfaces(ifaces, token)
			if ifaceErr != nil {
				log.Logger.Error().Err(ifaceErr).Msg("failed to add ethernet interfaces to SMD")
				ifaceErrorsOccurred = true
			}
			for _, err := range ifaceErrs {
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
		}

		// Put together list of groups to add and which components to add to those groups
		groupsToAdd := make(map[string]smd.Group)
		for _, node := range nodes.Nodes {
			if node.Group != "" {
				if g, ok := groupsToAdd[node.Group]; !ok {
					newGroup := smd.Group{
						Label:       node.Group,
						Description: fmt.Sprintf("The %s group", node.Group),
					}
					newGroup.Members.IDs = []string{node.Xname}
					groupsToAdd[node.Group] = newGroup
				} else {
					g.Members.IDs = append(g.Members.IDs, node.Xname)
					groupsToAdd[node.Group] = g
				}
			}
		}
		groupList := make([]smd.Group, len(groupsToAdd))
		var idx = 0
		for _, g := range groupsToAdd {
			groupList[idx] = g
			idx++
		}

		// Add groups and components to those groups
		var (
			groupErrorsOccurred bool = false
			groupHenvs          []client.HTTPEnvelope
			groupErrs           []error
			groupErr            error
		)
		if cmd.Flag("overwrite").Changed {
			// SMD's groups API does not allow the PUT method.
			// Instead, we loop over each group to add and attempt a
			// POST. Iff a 409 is returned for that interface, a
			// PATCH is attempted. Otherwise, an error has occurred.
			for _, group := range groupList {
				// Attempt to POST the group
				groupListWrapper := []smd.Group{group}
				groupHenvs, groupErrs, groupErr = smdClient.PostGroups(groupListWrapper, token)

				if groupErr != nil {
					// An error in the function occurred,
					// err for this group and move on.
					log.Logger.Error().Err(groupErr).Msg("failed to add group to SMD")
					groupErrorsOccurred = true
					continue
				}

				if groupErrs[0] != nil {
					// An HTTP error occurred
					if errors.Is(groupErrs[0], client.UnsuccessfulHTTPError) {
						if groupHenvs[0].StatusCode == 409 {
							// Group exists, patch it
							log.Logger.Info().Msgf("group %s exists, attempting to update it", group.Label)
							_, patchErrs, patchErr := smdClient.PatchGroups(groupListWrapper, token)
							if patchErr != nil {
								log.Logger.Error().Err(patchErr).Msg("failed to update existing group in SMD")
								groupErrorsOccurred = true
								continue
							}
							if patchErrs[0] != nil {
								var errMsg string
								if errors.Is(patchErrs[0], client.UnsuccessfulHTTPError) {
									errMsg = "SMD group PATCH yielded unsuccessful HTTP response"
								} else {
									errMsg = "failed to update existing group in SMD"
								}
								log.Logger.Error().Err(patchErrs[0]).Msg(errMsg)
								groupErrorsOccurred = true
								continue
							}
						} else {
							// Some other HTTP error occurred, err
							log.Logger.Error().Err(groupErrs[0]).Msg("SMD group POST yielded non-409 (duplicate) failure")
							groupErrorsOccurred = true
							continue
						}
					} else {
						log.Logger.Error().Err(groupErrs[0]).Msg("failed to add group to SMD")
						groupErrorsOccurred = true
						continue
					}
				}
			}
		} else {
			_, groupErrs, groupErr = smdClient.PostGroups(groupList, token)
			if groupErr != nil {
				log.Logger.Error().Err(groupErr).Msg("failed to add groups to SMD")
				groupErrorsOccurred = true
			}
			for _, err := range groupErrs {
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
		}

		// Notify user if any request errors occurred
		exitStatus := 0
		if compErrorsOccurred {
			log.Logger.Warn().Msg("component requests completed with errors")
			exitStatus = 1
		}
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
	discoverCmd.Flags().StringP("payload-format", "F", defaultPayloadFormat, "format of payload file (yaml,json) passed with --payload")
	discoverCmd.Flags().Bool("overwrite", false, "overwrite any existing information instead of failing")

	discoverCmd.MarkFlagRequired("payload")

	rootCmd.AddCommand(discoverCmd)
}
