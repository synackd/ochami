package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"

	"github.com/openchami/schemas/schemas/csm"
	"github.com/synackd/ochami/internal/log"
)

// SMDClient is an OchamiClient that has its BasePath set configured to the one
// that BSS uses.
type SMDClient struct {
	*OchamiClient
}

const (
	serviceNameSMD = "SMD"
	basePathSMD    = "/hsm/v2"

	SMDRelpathService            = "/service"
	SMDRelpathComponents         = "/State/Components"
	SMDRelpathEthernetInterfaces = "/Inventory/EthernetInterfaces"
	SMDRelpathRedfishEndpoints   = "/Inventory/RedfishEndpoints"
)

// Component is a minimal subset of SMD's Component struct that contains only
// what is necessary for sending a valid Component request to SMD.
type Component struct {
	ID      string `json:"ID"`
	State   string `json:"State"`
	Enabled bool   `json:"Enabled"`
	Role    string `json:"Role"`
	Arch    string `json:"Arch"`
	NID     int64  `json:"NID"`
}

// ComponentSlice is a convenience data structure to make marshalling Component
// requests easier.
type ComponentSlice struct {
	Components []Component `json:"Components"`
}

// EthernetInterface is a minimal subset of SMD's EthernetInterface struct that
// contains only what is necessary for sending a valid EthernetInterface request
// to SMD.
type EthernetInterface struct {
	ID          string       `json:"ID"`
	ComponentID string       `json:"ComponentID"`
	Description string       `json:"Description"`
	MACAddress  string       `json:"MACAddress"`
	IPAddresses []EthernetIP `json:"IPAddresses"`
}

type EthernetIP struct {
	IPAddress string `json:"IPAddress"`
	Network   string `json:"Network"`
}

// RedfishEndpoint slice is a convenience data structure to make marshalling
// RedfishEndpoint requests easier.
type RedfishEndpointSlice struct {
	RedfishEndpoints []csm.RedfishEndpoint `json:"RedfishEndpoints"`
}

// NewSMDClient takes a baseURI and basePath and returns a pointer to a new
// SMDClient. If an error occurred creating the embedded OchamiClient, it is
// returned. If insecure is true, TLS certificates will not be verified.
func NewSMDClient(baseURI string, insecure bool) (*SMDClient, error) {
	oc, err := NewOchamiClient(serviceNameSMD, baseURI, basePathSMD, insecure)
	if err != nil {
		return nil, fmt.Errorf("failed to create OchamiClient for %s: %w", serviceNameSMD, err)
	}
	sc := &SMDClient{
		OchamiClient: oc,
	}

	return sc, err
}

// GetStatus is a wrapper function around OchamiClient.GetData that takes an
// optional component and uses it to determine which subpath of the SMD /service
// endpoint to query. If empty, the /service/ready endpoint is queried.
// Otherwise:
//
// "all" -> "/service/values"
func (sc *SMDClient) GetStatus(component string) (HTTPEnvelope, error) {
	var (
		henv              HTTPEnvelope
		err               error
		smdStatusEndpoint string
	)
	switch component {
	case "":
		smdStatusEndpoint = path.Join(SMDRelpathService, "ready")
	case "all":
		smdStatusEndpoint = path.Join(SMDRelpathService, "values")
	default:
		return henv, fmt.Errorf("GetStatus(): unknown status component: %s", component)
	}

	henv, err = sc.GetData(smdStatusEndpoint, "", nil)
	if err != nil {
		err = fmt.Errorf("GetStatus(): error getting SMD all status: %w", err)
	}

	return henv, err
}

// GetComponentsAll is a wrapper function around OchamiClient.GetData that queries
// /State/Components.
func (sc *SMDClient) GetComponentsAll() (HTTPEnvelope, error) {
	henv, err := sc.GetData(SMDRelpathComponents, "", nil)
	if err != nil {
		err = fmt.Errorf("GetComponentsAll(): error getting components: %w", err)
	}

	return henv, err
}

// GetComponentsXname is like GetComponentsAll except that it takes a token and
// queries /State/Components/{xname}.
func (sc *SMDClient) GetComponentsXname(xname, token string) (HTTPEnvelope, error) {
	var henv HTTPEnvelope
	finalEP := SMDRelpathComponents + "/" + xname
	headers := NewHTTPHeaders()
	if token != "" {
		if err := headers.SetAuthorization(token); err != nil {
			return henv, fmt.Errorf("GetComponentsXname(): error setting token in HTTP headers")
		}
	}
	henv, err := sc.GetData(finalEP, "", headers)
	if err != nil {
		err = fmt.Errorf("GetComponentsXname(): error getting component for xname %q: %w", xname, err)
	}

	return henv, err
}

// GetComponentsNid is like GetComponentsAll except that it takes a token and
// queries /State/Components/ByNID/{nid}.
func (sc *SMDClient) GetComponentsNid(nid int32, token string) (HTTPEnvelope, error) {
	var henv HTTPEnvelope
	finalEP := SMDRelpathComponents + "/ByNID/" + fmt.Sprint(nid)
	headers := NewHTTPHeaders()
	if token != "" {
		if err := headers.SetAuthorization(token); err != nil {
			return henv, fmt.Errorf("GetComponentsNid(): error setting token in HTTP headers")
		}
	}
	henv, err := sc.GetData(finalEP, "", headers)
	if err != nil {
		err = fmt.Errorf("GetComponentsNid(): error getting component for NID %d: %w", nid, err)
	}

	return henv, err
}

// GetRedfishEndpoints is a wrapper around OchamiClient.GetData that takes an
// optional query string (without the "?") and a token. It sets token as the
// authorization bearer in the headers and passes the query string and headers
// to OchamiClient.GetData, using the SMD RedfishEndpoints API endpoint.
func (sc *SMDClient) GetRedfishEndpoints(query, token string) (HTTPEnvelope, error) {
	var (
		henv    HTTPEnvelope
		headers *HTTPHeaders
		err     error
	)
	headers = NewHTTPHeaders()
	if token != "" {
		if err = headers.SetAuthorization(token); err != nil {
			return henv, fmt.Errorf("GetRedfishEndpoints(): error setting token in HTTP headers")
		}
	}
	henv, err = sc.GetData(SMDRelpathRedfishEndpoints, query, headers)
	if err != nil {
		err = fmt.Errorf("GetRedfishEndpoints(): error getting redfish endpoints: %w", err)
	}

	return henv, err
}

// GetEthernetInterfaces is a wrapper around OchamiClient.GetData that takes a
// query string and passes it to OchamiClient.GetData using SMD's ethernet
// interfaces endpoint.
func (sc *SMDClient) GetEthernetInterfaces(query string) (HTTPEnvelope, error) {
	henv, err := sc.GetData(SMDRelpathEthernetInterfaces, query, nil)
	if err != nil {
		err = fmt.Errorf("GetEthernetInterfaces(): error getting ethernet interfaces: %w", err)
	}

	return henv, err
}

// GetEthernetInterfacesByID is a wrapper around OchamiClient.GetData that takes
// an ethernet interface ID, token, and a flag indicating if the ethernet
// interface itself should be retrieved or a list of its IPs. It passes these to
// OchamiClient.GetData, setting the token as the authorization bearer in the
// request headers.
func (sc *SMDClient) GetEthernetInterfaceByID(id, token string, getIPs bool) (HTTPEnvelope, error) {
	var (
		ep      string
		err     error
		henv    HTTPEnvelope
		headers *HTTPHeaders
	)
	headers = NewHTTPHeaders()
	if token != "" {
		if err = headers.SetAuthorization(token); err != nil {
			return henv, fmt.Errorf("GetRedfishEndpoints(): error setting token in HTTP headers")
		}
	}
	if getIPs {
		if ep, err = url.JoinPath(SMDRelpathEthernetInterfaces, id); err != nil {
			return henv, fmt.Errorf("GetEthernetInterfacesByID(): failed to join ethernet path (%s) with id (%s): %w", SMDRelpathEthernetInterfaces, id, err)
		}
		if ep, err = url.JoinPath(ep, "IPAddresses"); err != nil {
			return henv, fmt.Errorf("GetEthernetInterfacesByID(): failed to join endpoint %s with \"IPAddresses\": %w", ep, err)
		}
	} else {
		ep, err = url.JoinPath(SMDRelpathEthernetInterfaces, id)
	}
	return sc.GetData(ep, "", headers)
}

// PostComponents is a wrapper function around OchamiClient.PostData that takes
// a ComponentSlice and a token, puts the token in the request headers as an
// authorization bearer, marshalls compSlice as JSON and sets it as the request
// body, then passes it to Ochami.PostData.
func (sc *SMDClient) PostComponents(compSlice ComponentSlice, token string) (HTTPEnvelope, error) {
	var (
		henv    HTTPEnvelope
		headers *HTTPHeaders
		body    HTTPBody
		err     error
	)
	if body, err = json.Marshal(compSlice); err != nil {
		return henv, fmt.Errorf("PostComponents(): failed to marshal ComponentArray: %w", err)
	}
	headers = NewHTTPHeaders()
	if token != "" {
		if err := headers.SetAuthorization(token); err != nil {
			return henv, fmt.Errorf("PostComponents(): error setting token in HTTP headers")
		}
	}
	henv, err = sc.PostData(SMDRelpathComponents, "", headers, body)
	if err != nil {
		err = fmt.Errorf("PostComponents(): failed to POST component(s) to SMD: %w", err)
	}

	return henv, err
}

// PostRedfishEndpoints is a wrapper function around OchamiClient.PostData that
// takes a RedfishEndpointSlice and a token, puts the token in the request
// headers as an authorization bearer, and iteratively calls
// OchamiClient.PostData using each RedfishEndpoint in the slice.
func (sc *SMDClient) PostRedfishEndpoints(rfes RedfishEndpointSlice, token string) ([]HTTPEnvelope, []error, error) {
	var (
		errors  []error
		henvs   []HTTPEnvelope
		headers *HTTPHeaders
	)
	headers = NewHTTPHeaders()
	if token != "" {
		if err := headers.SetAuthorization(token); err != nil {
			return nil, []error{}, fmt.Errorf("PostRedfishEndpoints(): error setting token in HTTP headers")
		}
	}
	for _, rfe := range rfes.RedfishEndpoints {
		var body HTTPBody
		var err error
		if body, err = json.Marshal(rfe); err != nil {
			newErr := fmt.Errorf("PostRedfishEndpoints(): failed to marshal RedfishEndpoint: %w", err)
			errors = append(errors, newErr)
			henvs = append(henvs, HTTPEnvelope{})
			continue
		}
		henv, err := sc.PostData(SMDRelpathRedfishEndpoints, "", headers, body)
		henvs = append(henvs, henv)
		if err != nil {
			newErr := fmt.Errorf("PostRedfishEndpoints(): failed to POST redfish endpoint to SMD: %w", err)
			log.Logger.Debug().Err(err).Msg("failed to add redfish endpoint")
			errors = append(errors, newErr)
			continue
		}
		log.Logger.Debug().Msgf("successfully added component %s", rfe.ID)
		errors = append(errors, nil)
	}

	return henvs, errors, nil
}

// PostEthernetInterfaces is a wrapper function around OchamiClient.PostData
// that takes a slice of EthernetInterfaces and a token, puts the token in the
// request headers as an authorization bearer, and iteratively calls
// OchamiClient.PostData using each EthernetInterface in the slice.
func (sc *SMDClient) PostEthernetInterfaces(eis []EthernetInterface, token string) ([]HTTPEnvelope, []error, error) {
	var (
		errors  []error
		henvs   []HTTPEnvelope
		headers *HTTPHeaders
	)
	headers = NewHTTPHeaders()
	if token != "" {
		if err := headers.SetAuthorization(token); err != nil {
			return nil, []error{}, fmt.Errorf("PostEthernetInterfaces(): error setting token in HTTP headers")
		}
	}
	for _, ei := range eis {
		var body HTTPBody
		var err error
		if body, err = json.Marshal(ei); err != nil {
			newErr := fmt.Errorf("PostEthernetInterfaces(): failed to marshal EthernetInterface: %w", err)
			errors = append(errors, newErr)
			henvs = append(henvs, HTTPEnvelope{})
			continue
		}
		henv, err := sc.PostData(SMDRelpathEthernetInterfaces, "", headers, body)
		henvs = append(henvs, henv)
		if err != nil {
			newErr := fmt.Errorf("PostEthernetInterfaces(): failed to POST ethernet interface(s) to SMD: %w", err)
			log.Logger.Debug().Err(newErr).Msg("failed to add ethernet interface")
			errors = append(errors, newErr)
			continue
		}
		log.Logger.Debug().Msgf("successfully added ethernet interface for component %s", ei.ComponentID)
		errors = append(errors, nil)
	}

	return henvs, errors, nil
}

// DeleteComponents takes a token and xnames and iteratively calls
// OchamiClient.DeleteData for each xname. This is necessary because SMD only
// allows deleting one xname at a time. A slice of HTTPEnvelopes is returned
// containing one HTTPEnvelope per deletion, as well as an error slice
// containing errors corresponding to each deletion. The indexes of these should
// correspond. If an error in the function itself occurred, a separate error is
// returned. This is to distinguish HTTP request errors from control flow
// errors.
func (sc *SMDClient) DeleteComponents(token string, xnames ...string) ([]HTTPEnvelope, []error, error) {
	headers := NewHTTPHeaders()
	if token != "" {
		if err := headers.SetAuthorization(token); err != nil {
			return nil, []error{}, fmt.Errorf("DeleteComponents(): error setting token in HTTP headers")
		}
	}
	var errors []error
	var henvs []HTTPEnvelope
	for _, xname := range xnames {
		xnamePath, err := url.JoinPath(SMDRelpathComponents, xname)
		if err != nil {
			newErr := fmt.Errorf("DeleteComponents(): failed join component path (%s) with xname (%s): %w", SMDRelpathComponents, xname, err)
			henvs = append(henvs, HTTPEnvelope{})
			errors = append(errors, newErr)
			continue
		}
		henv, err := sc.DeleteData(xnamePath, "", headers, nil)
		henvs = append(henvs, henv)
		if err != nil {
			newErr := fmt.Errorf("DeleteComponents(): failed to DELETE component %s in SMD: %w", xname, err)
			log.Logger.Debug().Err(err).Msgf("failed to delete component %s", xname)
			errors = append(errors, newErr)
			continue
		}
		log.Logger.Debug().Msgf("successfully deleted component %s", xname)
		errors = append(errors, nil)
	}

	return henvs, errors, nil
}

// DeleteComponentsAll is a wrapper function around OchamiClient.DeleteData that
// takes a token, puts it in the request headers as an authorization bearer, and
// sends it in a DELETE request to the SMD components endpoint. This should
// delete all components SMD knows about if the token is authorized.
func (sc *SMDClient) DeleteComponentsAll(token string) (HTTPEnvelope, error) {
	var (
		henv    HTTPEnvelope
		headers *HTTPHeaders
		err     error
	)

	headers = NewHTTPHeaders()
	if token != "" {
		if err := headers.SetAuthorization(token); err != nil {
			return henv, fmt.Errorf("DeleteComponentsAll(): error setting token in HTTP headers")
		}
	}
	henv, err = sc.DeleteData(SMDRelpathComponents, "", headers, nil)
	if err != nil {
		err = fmt.Errorf("DeleteComponentsAll(): failed to DELETE component(s) to SMD: %w", err)
	}

	return henv, err
}

// DeleteRedfishEndpoints takes a token and xnames and iteratively calls
// OchamiClient.DeleteData for each xname. This is necessary because SMD only
// allows deleting one xname at a time. A slice of HTTPEnvelopes is returned
// containing one HTTPEnvelope per deletion, as well as an error slice
// containing errors corresponding to each deletion. The indexes of these should
// correspond. If an error in the function itself occurred, a separate error is
// returned. This is to distinguish HTTP request errors from control flow
// errors.
func (sc *SMDClient) DeleteRedfishEndpoints(token string, xnames ...string) ([]HTTPEnvelope, []error, error) {
	headers := NewHTTPHeaders()
	if token != "" {
		if err := headers.SetAuthorization(token); err != nil {
			return nil, []error{}, fmt.Errorf("DeleteRedfishEndpoints(): error setting token in HTTP headers")
		}
	}
	var errors []error
	var henvs []HTTPEnvelope
	for _, xname := range xnames {
		xnamePath, err := url.JoinPath(SMDRelpathRedfishEndpoints, xname)
		if err != nil {
			newErr := fmt.Errorf("DeleteRedfishEndpoints(): failed join component path (%s) with xname (%s): %w", SMDRelpathRedfishEndpoints, xname, err)
			henvs = append(henvs, HTTPEnvelope{})
			errors = append(errors, newErr)
			continue
		}
		henv, err := sc.DeleteData(xnamePath, "", headers, nil)
		henvs = append(henvs, henv)
		if err != nil {
			newErr := fmt.Errorf("DeleteRedfishEndpoints(): failed to DELETE component %s in SMD: %w", xname, err)
			log.Logger.Debug().Err(err).Msgf("failed to delete component %s", xname)
			errors = append(errors, newErr)
			continue
		}
		log.Logger.Debug().Msgf("successfully deleted component %s", xname)
		errors = append(errors, nil)
	}

	return henvs, errors, nil
}

// DeleteRedfishEndpointsAll is a wrapper function around
// OchamiClient.DeleteData that takes a token, puts it in the request headers as
// an authorization bearer, and sends it in a DELETE request to the SMD redfish
// endpoints endpoint. This should delete all redfish endpoints SMD knows about
// if the token is authorized.
func (sc *SMDClient) DeleteRedfishEndpointsAll(token string) (HTTPEnvelope, error) {
	var (
		henv    HTTPEnvelope
		headers *HTTPHeaders
		err     error
	)

	headers = NewHTTPHeaders()
	if token != "" {
		if err := headers.SetAuthorization(token); err != nil {
			return henv, fmt.Errorf("DeleteRedfishEndpointsAll(): error setting token in HTTP headers")
		}
	}
	henv, err = sc.DeleteData(SMDRelpathRedfishEndpoints, "", headers, nil)
	if err != nil {
		err = fmt.Errorf("DeleteRedfishEndpointsAll(): failed to DELETE redfish endpoint(s) to SMD: %w", err)
	}

	return henv, err
}

// DeleteEthernetInterfaces takes a token and one or more ethernet interface IDs
// and iteratively calls OchamiClient.DeleteData for each ID. This is necessary
// because SMD only allows deleting one ethernet interface at a time. A slice of
// HTTPEnvelopes is returned containing one HTTPEnvelope per deletion, as well
// as an error slice containing errors corresponding to each deletion. The
// indexes of these should correspond. If an error in the function itself
// occurred, a separate error is returned. This is to distinguish HTTP request
// errors from control flow errors.
func (sc *SMDClient) DeleteEthernetInterfaces(token string, eIds ...string) ([]HTTPEnvelope, []error, error) {
	headers := NewHTTPHeaders()
	if token != "" {
		if err := headers.SetAuthorization(token); err != nil {
			return nil, []error{}, fmt.Errorf("DeleteEthernetInterfaces(): error setting token in HTTP headers")
		}
	}
	var errors []error
	var henvs []HTTPEnvelope
	for _, eId := range eIds {
		eIdPath, err := url.JoinPath(SMDRelpathEthernetInterfaces, eId)
		if err != nil {
			newErr := fmt.Errorf("DeleteEthernetInterfaces(): failed join component path (%s) with ethernet interface %s: %w", SMDRelpathEthernetInterfaces, eId, err)
			henvs = append(henvs, HTTPEnvelope{})
			errors = append(errors, newErr)
			continue
		}
		henv, err := sc.DeleteData(eIdPath, "", headers, nil)
		henvs = append(henvs, henv)
		if err != nil {
			newErr := fmt.Errorf("DeleteEthernetInterfaces(): failed to DELETE ethernet interface %s in SMD: %w", eId, err)
			log.Logger.Debug().Err(err).Msgf("failed to delete ethernet interface %s", eId)
			errors = append(errors, newErr)
			continue
		}
		log.Logger.Debug().Msgf("successfully deleted ethernet interface %s", eId)
		errors = append(errors, nil)
	}

	return henvs, errors, nil
}

// DeleteEthernetInterfacesAll is a wrapper function around
// OchamiClient.DeleteData that takes a token, puts it in the request headers as
// an authorization bearer, and sends it in a DELETE request to the SMD ethernet
// interfaces endpoint. This should delete all ethernet interfaces SMD knows
// about if the token is authorized.
func (sc *SMDClient) DeleteEthernetInterfacesAll(token string) (HTTPEnvelope, error) {
	var (
		henv    HTTPEnvelope
		headers *HTTPHeaders
		err     error
	)

	headers = NewHTTPHeaders()
	if token != "" {
		if err := headers.SetAuthorization(token); err != nil {
			return henv, fmt.Errorf("DeleteEthernetInterfacesAll(): error setting token in HTTP headers")
		}
	}
	henv, err = sc.DeleteData(SMDRelpathEthernetInterfaces, "", headers, nil)
	if err != nil {
		err = fmt.Errorf("DeleteEthernetInterfacesAll(): failed to DELETE ethernet interface(s) to SMD: %w", err)
	}

	return henv, err
}
