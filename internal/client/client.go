package client

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/synackd/ochami/internal/version"
)

// OchamiClient is an *http.Client that contains metadata for OpenCHAMI services
// being communicated with.
type OchamiClient struct {
	*http.Client
	BaseURI  *url.URL // Base URL for OpenCHAMI services (e.g. https://foobar.openchami.cluster)
	BasePath string   // Base path for the service (e.g. /boot/v1 for BSS)
}

var userAgent = "ochami/" + version.Version

// defaultClient creates an http.DefaultClient for its OchamiClient and
// configures it to not try to verify TLS certificates.
func (oc *OchamiClient) defaultClient() {
	oc.Client = http.DefaultClient
	oc.Client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			// Default client does not verify server certificate
			InsecureSkipVerify: true,
		},
	}
}

// NewOchamiClient takes a baseURI and basePath and returns a pointer to a new
// OchamiClient. If an error occurs parsing baseURI, it is returned. baseURI is
// the base URI of the OpenCHAMI services (e.g.
// https://foobar.openchami.cluster) and basePath is the endpoint prefix that is
// service-dependent (e.g. for BSS it could be "/boot/v1").
func NewOchamiClient(baseURI, basePath string) (*OchamiClient, error) {
	u, err := url.Parse(baseURI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URI: %v", err)
	}
	oc := &OchamiClient{
		BaseURI: u,
		BasePath: basePath,
	}
	oc.defaultClient()
	return oc, err
}

// MakeRequest is a convenience function that, using an OchamiClient as the HTTP
// client, sends an HTTP request to the passed uri including optional headers
// and body, and uses the passed HTTP method.
func (oc *OchamiClient) MakeRequest(method, uri string, headers *HTTPHeaders, body HTTPBody) (*http.Response, HTTPBody, error) {
	// Create request using function args
	req, err := http.NewRequest(method, uri, bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create new HTTP request: %v", err)
	}

	// Create empty headers if headers pointer is nil so range works
	if headers == nil {
		headers = NewHTTPHeaders()
	}

	// Add headers, including user agent
	req.Header.Add("User-Agent", userAgent)
	for k, v := range *headers {
		req.Header.Add(k, v)
	}

	// Execute HTTP request
	res, err := oc.Client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute HTTP request: %v", err)
	}

	// Read response
	resBody, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read HTTP response body: %v", err)
	}

	return res, resBody, err
}
