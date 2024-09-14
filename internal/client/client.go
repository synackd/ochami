package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/synackd/ochami/internal/version"
)

// OchamiClient is an *http.Client that contains metadata for OpenCHAMI services
// being communicated with.
type OchamiClient struct {
	*http.Client
	BaseURI  *url.URL // Base URL for OpenCHAMI services (e.g. https://foobar.openchami.cluster)
	BasePath string   // Base path for the service (e.g. /boot/v1 for BSS)
}

var (
	userAgent = "ochami/" + version.Version

	// TLS timeout configuration
	tlsHandshakeTimeout   = 120 * time.Second
	responseHeaderTimeout = 120 * time.Second
)

// defaultClient creates an http.DefaultClient for its OchamiClient.
func (oc *OchamiClient) defaultClient() {
	oc.Client = http.DefaultClient
}

// defaultClientInsecure creates an http.DefaultClient for its OchamiClient and
// configures it to not try to verify TLS certificates.
func (oc *OchamiClient) defaultClientInsecure() {
	oc.Client = http.DefaultClient
	oc.Client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			// This default client does not verify server certificate
			InsecureSkipVerify: true,
		},
	}
}

// NewOchamiClient takes a baseURI and basePath and returns a pointer to a new
// OchamiClient. If an error occurs parsing baseURI, it is returned. baseURI is
// the base URI of the OpenCHAMI services (e.g.
// https://foobar.openchami.cluster) and basePath is the endpoint prefix that is
// service-dependent (e.g. for BSS it could be "/boot/v1"). If insecure is true,
// the client will not verify TLS certificates.
func NewOchamiClient(baseURI, basePath string, insecure bool) (*OchamiClient, error) {
	u, err := url.Parse(baseURI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URI: %v", err)
	}
	oc := &OchamiClient{
		BaseURI:  u,
		BasePath: basePath,
	}
	if insecure {
		oc.defaultClientInsecure()
	} else {
		oc.defaultClient()
	}
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

// UseCACert takes a path to a CA certificate bundle in PEM format and sets it
// as the OchamiClient's certificate authority certificate to verify the
// certificates of connections to TLS-enabled HTTP URIs (HTTPS).
func (oc *OchamiClient) UseCACert(caCertPath string) error {
	cacert, err := os.ReadFile(caCertPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %v", caCertPath, err)
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(cacert)

	if oc == nil {
		return fmt.Errorf("client is nil")
	}

	(*oc).Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:            certPool,
			InsecureSkipVerify: false,
		},
		DisableKeepAlives:     true,
		TLSHandshakeTimeout:   tlsHandshakeTimeout,
		ResponseHeaderTimeout: responseHeaderTimeout,
	}

	return nil
}
