package client

import (
	"fmt"
	"path"
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
