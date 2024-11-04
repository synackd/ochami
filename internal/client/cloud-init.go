package client

import (
	"fmt"
	"net/url"
)

// CloudInitClient is an OchamiClient that has its BasePath configured to the
// one that the cloud-init service uses.
type CloudInitClient struct {
	*OchamiClient
}

const (
	serviceNameCloudInit = "cloud-init"
	// cloud-init doesn't have a service prefix and has two separate
	// endpoints. To mitigate this, we treat the service root as '/' and use
	// the relative paths as the service endpoints.
	basePathCloudInit      = "/"
	cloudInitRelpathOpen   = "/cloud-init"
	cloudInitRelpathSecure = "/cloud-init-secure"
)

// NewCloudInitClient takes a baseURI and basePath and returns a pointer to a
// new CloudInitClient. If an error occurred creating the embedded
// OchamiClient, it is returned. If insecure is true, TLS certificates will not
// be verified.
func NewCloudInitClient(baseURI string, insecure bool) (*CloudInitClient, error) {
	oc, err := NewOchamiClient(serviceNameCloudInit, baseURI, basePathCloudInit, insecure)
	if err != nil {
		return nil, fmt.Errorf("failed to create OchamiClient for %s: %w", serviceNameCloudInit, err)
	}
	cic := &CloudInitClient{
		OchamiClient: oc,
	}

	return cic, err
}

// GetConfigs is a wrapper function around OchamiClient.GetData that determines
// whether to use only the cloud-init base path or it appended with an id and
// calls GetData on the endpoint, returning the result. If an error occurs in
// the function or via HTTP, it is returned as well. If id is blank, all configs
// are returned. Otherwise, just the config for the id is returned.
func (cic *CloudInitClient) GetConfigs(id string) (HTTPEnvelope, error) {
	finalEP := cloudInitRelpathOpen
	if id != "" {
		var err error
		finalEP, err = url.JoinPath(cloudInitRelpathOpen, id)
		if err != nil {
			return HTTPEnvelope{}, fmt.Errorf("GetConfigs(): failed to join cloud-init open path (%s) with id %s: %w", cloudInitRelpathOpen, id, err)
		}
	}
	henv, err := cic.GetData(finalEP, "", nil)
	if err != nil {
		err = fmt.Errorf("GetConfigs(): error getting cloud-init configs: %w", err)
	}

	return henv, err
}

// GetConfigsSecure is like GetConfigs except that it uses the secure cloud-init
// endpoint and thus requires a token.
func (cic *CloudInitClient) GetConfigsSecure(id, token string) (HTTPEnvelope, error) {
	headers := NewHTTPHeaders()
	if token != "" {
		if err := headers.SetAuthorization(token); err != nil {
			return HTTPEnvelope{}, fmt.Errorf("GetConfigsSecure(): error setting token in HTTP headers")
		}
	}
	finalEP := cloudInitRelpathSecure
	if id != "" {
		var err error
		finalEP, err = url.JoinPath(cloudInitRelpathSecure, id)
		if err != nil {
			return HTTPEnvelope{}, fmt.Errorf("GetConfigsSecure(): failed to join cloud-init secure path (%s) with id %s: %w", cloudInitRelpathSecure, id, err)
		}
	}
	henv, err := cic.GetData(finalEP, "", headers)
	if err != nil {
		err = fmt.Errorf("GetConfigsSecure(): error getting secure cloud-init configs: %w", err)
	}

	return henv, err
}
