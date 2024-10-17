package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"

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

	SMDRelpathService    = "/service"
	SMDRelpathComponents = "/State/Components"
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

// GetStatus is a wrapper function around SMDClient.GetData that takes an
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

// GetComponentsAll is a wrapper function around SMDClient.GetData that queries
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

// PostComponents is a wrapper function around OchamiClient.PostData that takes
// a ComponentSlice and a token, puts the token in the request headers as an
// authorization bearer, marshalls compSlice as JSON and sets it as the request
// body, then basses it to Ochami.PostData.
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
