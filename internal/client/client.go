package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/synackd/ochami/internal/log"
	"github.com/synackd/ochami/internal/version"
)

// OchamiClient is an *http.Client that contains metadata for OpenCHAMI services
// being communicated with.
type OchamiClient struct {
	*http.Client
	BaseURI     *url.URL // Base URL for OpenCHAMI services (e.g. https://foobar.openchami.cluster)
	BasePath    string   // Base path for the service (e.g. /boot/v1 for BSS)
	ServiceName string   // Name of service being contacted (e.g. BSS)
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
func NewOchamiClient(serviceName, baseURI, basePath string, insecure bool) (*OchamiClient, error) {
	u, err := url.Parse(baseURI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URI: %w", err)
	}
	oc := &OchamiClient{
		BaseURI:     u,
		BasePath:    basePath,
		ServiceName: serviceName,
	}
	if insecure {
		oc.defaultClientInsecure()
	} else {
		oc.defaultClient()
	}
	return oc, err
}

// GetURI takes an endpoint and joins it with the OchamiClient's BaseURI and
// BasePath to form the final URI to be used for a request. If query is
// specified, it is used as a raw query string and appended onto the URL
// without URL encoding. query should not contain the initial '?'.
func (oc *OchamiClient) GetURI(endpoint, query string) (string, error) {
	uri, err := url.Parse(oc.BaseURI.String())
	if err != nil {
		return "", fmt.Errorf("failed to parse base URI %s: %w", oc.BaseURI, err)
	}
	uri.Path = path.Join(uri.Path, oc.BasePath, endpoint)
	if query != "" {
		uri.RawQuery = query
	}
	return uri.String(), err
}

// GetData is a wrapper around MakeOchamiRequest that sends a GET request to
// endpoint, using an optional token and optional headers, and returns an
// HTTPEnvelope containg the response metadata and the data received in the
// response along with a nil error. If the HTTP response code is unsuccessful
// (i.e. not 2XX), then the returned error will contain an
// UnsuccessfulHTTPError. Otherwise, the error that occurred is returned. query
// is the raw query string (without the '?') to be added to the URI. It should
// already be URL-encoded, e.g. generated using url.Values' Encode() function.
func (oc *OchamiClient) GetData(endpoint, query string, headers *HTTPHeaders) (HTTPEnvelope, error) {
	var he HTTPEnvelope

	res, err := oc.MakeOchamiRequest(http.MethodGet, endpoint, query, headers, nil)
	if err != nil {
		return he, fmt.Errorf("error making GET request to %s: %w", oc.ServiceName, err)
	}
	if res != nil {
		he, err := NewHTTPEnvelopeFromResponse(res)
		if err != nil {
			return he, fmt.Errorf("could not create HTTP envelope from GET response: %w", err)
		}
		return he, he.CheckResponse()
	}
	return he, fmt.Errorf("%s GET response was empty", oc.ServiceName)
}

// PostData is a wrapper around MakeOchamiRequest that sends a POST request to
// endpoint, using an optional token, optional headers, a body, and returns an
// HTTPEnvelope containg the response metadata and the data received in the
// response along with a nil error. If the HTTP response code is unsuccessful
// (i.e. not 2XX), then the returned error will contain an
// UnsuccessfulHTTPError. Otherwise, the error that occurred is returned. query
// is the raw query string (without the '?') to be added to the URI. It should
// already be URL-encoded, e.g. generated using url.Values' Encode() function.
func (oc *OchamiClient) PostData(endpoint, query string, headers *HTTPHeaders, body HTTPBody) (HTTPEnvelope, error) {
	var he HTTPEnvelope

	res, err := oc.MakeOchamiRequest(http.MethodPost, endpoint, query, headers, body)
	if err != nil {
		return he, fmt.Errorf("error making POST request to %s, %w", oc.ServiceName, err)
	}
	if res != nil {
		he, err := NewHTTPEnvelopeFromResponse(res)
		if err != nil {
			return he, fmt.Errorf("could not create HTTP envelope from POST response: %w", err)
		}
		return he, he.CheckResponse()
	}
	return he, fmt.Errorf("%s POST response was empty", oc.ServiceName)
}

// PutData is a wrapper around MakeOchamiRequest that sends a PUT request to
// endpoint, using an optional token, optional headers, a body, and returns an
// HTTPEnvelope containg the response metadata and the data received in the
// response along with a nil error. If the HTTP response code is unsuccessful
// (i.e. not 2XX), then the returned error will contain an
// UnsuccessfulHTTPError. Otherwise, the error that occurred is returned. query
// is the raw query string (without the '?') to be added to the URI. It should
// already be URL-encoded, e.g. generated using url.Values' Encode() function.
func (oc *OchamiClient) PutData(endpoint, query string, headers *HTTPHeaders, body HTTPBody) (HTTPEnvelope, error) {
	var he HTTPEnvelope

	res, err := oc.MakeOchamiRequest(http.MethodPut, endpoint, query, headers, body)
	if err != nil {
		return he, fmt.Errorf("error making PUT request to %s, %w", oc.ServiceName, err)
	}
	if res != nil {
		he, err := NewHTTPEnvelopeFromResponse(res)
		if err != nil {
			return he, fmt.Errorf("could not create HTTP envelope from PUT response: %w", err)
		}
		return he, he.CheckResponse()
	}
	return he, fmt.Errorf("%s PUT response was empty", oc.ServiceName)
}

// PatchData is a wrapper around MakeOchamiRequest that sends a PATCH request to
// endpoint, using an optional token, optional headers, a body, and returns an
// HTTPEnvelope containg the response metadata and the data received in the
// response along with a nil error. If the HTTP response code is unsuccessful
// (i.e. not 2XX), then the returned error will contain an
// UnsuccessfulHTTPError. Otherwise, the error that occurred is returned. query
// is the raw query string (without the '?') to be added to the URI. It should
// already be URL-encoded, e.g. generated using url.Values' Encode() function.
func (oc *OchamiClient) PatchData(endpoint, query string, headers *HTTPHeaders, body HTTPBody) (HTTPEnvelope, error) {
	var he HTTPEnvelope

	res, err := oc.MakeOchamiRequest(http.MethodPatch, endpoint, query, headers, body)
	if err != nil {
		return he, fmt.Errorf("error making PATCH request to %s, %w", oc.ServiceName, err)
	}
	if res != nil {
		he, err := NewHTTPEnvelopeFromResponse(res)
		if err != nil {
			return he, fmt.Errorf("could not create HTTP envelope from PATCH response: %w", err)
		}
		return he, he.CheckResponse()
	}
	return he, fmt.Errorf("%s PATCH response was empty", oc.ServiceName)
}

// DeleteData is a wrapper around MakeOchamiRequest that sends a DELETE request
// to endpoint, using an optional token, optional headers, a body, and returns
// an HTTPEnvelope containg the response metadata and the data received in the
// response along with a nil error. If the HTTP response code is unsuccessful
// (i.e. not 2XX), then the returned error will contain an
// UnsuccessfulHTTPError. Otherwise, the error that occurred is returned. query
// is the raw query string (without the '?') to be added to the URI. It should
// already be URL-encoded, e.g. generated using url.Values' Encode() function.
func (oc *OchamiClient) DeleteData(endpoint, query string, headers *HTTPHeaders, body HTTPBody) (HTTPEnvelope, error) {
	var he HTTPEnvelope

	res, err := oc.MakeOchamiRequest(http.MethodDelete, endpoint, query, headers, body)
	if err != nil {
		return he, fmt.Errorf("error making PATCH request to %s, %w", oc.ServiceName, err)
	}
	if res != nil {
		he, err := NewHTTPEnvelopeFromResponse(res)
		if err != nil {
			return he, fmt.Errorf("could not create HTTP envelope from PATCH response: %w", err)
		}
		return he, he.CheckResponse()
	}
	return he, fmt.Errorf("%s PATCH response was empty", oc.ServiceName)
}

// MakeOchamiRequest is a wrapper around MakeRequest that calls GetURI to form
// the final URI to make the request with and pass to MakeRequest.
func (oc *OchamiClient) MakeOchamiRequest(method, endpoint, query string, headers *HTTPHeaders, body HTTPBody) (*http.Response, error) {
	uri, err := oc.GetURI(endpoint, query)
	if err != nil {
		if query == "" {
			return nil, fmt.Errorf("failed to generate URI for endpoint %s: %w", endpoint, err)
		} else {
			return nil, fmt.Errorf("failed to generate URI for endpoint %s and query %s: %w", endpoint, query, err)
		}
	}

	return oc.MakeRequest(method, uri, headers, body)
}

// MakeRequest is a convenience function that, using an OchamiClient as the HTTP
// client, sends an HTTP request to the passed uri including optional headers
// and body, and uses the passed HTTP method.
func (oc *OchamiClient) MakeRequest(method, uri string, headers *HTTPHeaders, body HTTPBody) (*http.Response, error) {
	// Create request using function args
	log.Logger.Debug().Msgf("%s: %s", method, uri)
	req, err := http.NewRequest(method, uri, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create new HTTP request: %w", err)
	}

	// Create empty headers if headers pointer is nil so range works
	if headers == nil {
		headers = NewHTTPHeaders()
	}

	// Add headers, including user agent
	req.Header.Add("User-Agent", userAgent)
	for key, vals := range *headers {
		for _, val := range vals {
			req.Header.Add(key, val)
		}
	}

	// Debug info for request
	if len(req.Header) > 0 {
		log.Logger.Debug().Msg("Request headers:")
		for k, v := range req.Header {
			log.Logger.Debug().Msgf("  %s: %s", k, v)
		}
	} else {
		log.Logger.Debug().Msg("No headers in request")
	}
	if len(body) > 0 {
		log.Logger.Debug().Msg("Request body:")
		log.Logger.Debug().Msgf("%s", string(body))
	} else {
		log.Logger.Debug().Msg("No body in request")
	}

	// Execute HTTP request
	res, err := oc.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	// Debug info for response
	if res != nil {
		log.Logger.Debug().Msg("Response status: " + res.Status)
		if len(res.Header) > 0 {
			log.Logger.Debug().Msg("Response headers:")
			for k, v := range res.Header {
				log.Logger.Debug().Msgf("  %s: %s", k, v)
			}
		} else {
			log.Logger.Debug().Msg("No headers in response")
		}
		resBodyLen := res.ContentLength
		if resBodyLen > 0 {
			var resBodyCopy bytes.Buffer
			resBodyReader := io.TeeReader(res.Body, &resBodyCopy)
			resBodyBytes, err := ioutil.ReadAll(resBodyReader)
			if err != nil {
				log.Logger.Error().Err(err).Msg("failed to read body for debug message")
			}
			log.Logger.Debug().Msg("Response body:")
			log.Logger.Debug().Msgf("%s", string(resBodyBytes))
			res.Body = io.NopCloser(bytes.NewReader(resBodyBytes))
		} else {
			log.Logger.Debug().Msg("No body in response")
		}
	} else {
		log.Logger.Debug().Msg("Response was nil")
	}

	return res, err
}

// UseCACert takes a path to a CA certificate bundle in PEM format and sets it
// as the OchamiClient's certificate authority certificate to verify the
// certificates of connections to TLS-enabled HTTP URIs (HTTPS).
func (oc *OchamiClient) UseCACert(caCertPath string) error {
	cacert, err := os.ReadFile(caCertPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", caCertPath, err)
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

// FileToHTTPBody takes a file path and string representing the format of the
// file, reads the file, and tries to marshal it into an HTTPBody (byte array)
// in JSON form, returning it. If an unmarshalling error occurs or either of the
// arguments are empty, nil and an error are returned. Current file formats
// supported are JSON and YAML.
func FileToHTTPBody(path, format string) (HTTPBody, error) {
	if path == "" {
		return nil, fmt.Errorf("file path is empty")
	}
	if format == "" {
		return nil, fmt.Errorf("format is empty")
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", path, err)
	}

	var b HTTPBody
	switch strings.ToLower(format) {
	case "json":
		var j interface{}
		err = json.Unmarshal(contents, &j)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON contents from %q: %w", path, err)
		}
		b, err = json.Marshal(j)
		if err != nil {
			err = fmt.Errorf("failed to marshal JSON from file %q: %w", path, err)
		}
	case "yaml":
		var y interface{}
		err = yaml.Unmarshal(contents, &y)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal YAML contents from %q: %w", path, err)
		}
		y = CanonicalizeInterface(y)
		b, err = json.Marshal(y)
		if err != nil {
			err = fmt.Errorf("failed to marshal JSON (converted from YAML) from file %q: %w", path, err)
		}
	}

	return b, err
}

// ReadPayload reads in the file pointed to by path and unmarshals the data into
// value v. The data can be in formats other than JSON (whichever formats
// FileToHTTPBody supports), such as YAML. If a marshalling/unmarshalling error
// occurs or either path or format are empty, an error is returned.
func ReadPayload(path, format string, v any) error {
	log.Logger.Debug().Msgf("Payload file: %s", path)
	log.Logger.Debug().Msgf("Payload file format: %s", format)

	body, err := FileToHTTPBody(path, format)
	if err != nil {
		return fmt.Errorf("unable to create HTTP body from file: %w", err)
	}
	log.Logger.Debug().Msgf("Body bytes: %s", string(body))

	err = json.Unmarshal(body, v)
	if err != nil {
		err = fmt.Errorf("unable to unmarshal bytes into value: %w", err)
	}

	return err
}

// CanonicalizeInterface takes an arbitrary map of data (e.g. returned from
// unmarshalling) and ensures that the keys of the nested map structures are
// comparable (e.g. preparing it for a future marshaling), doing this
// recursively. This can be necessary as the yaml package's unmarshalling
// function unmarshals key-value pairs into map[interface{}]interface{} if
// unmarshalling into an interface{}.
//
// Adapted from: https://stackoverflow.com/a/75503306
func CanonicalizeInterface(i interface{}) interface{} {
	switch iType := i.(type) {
	case map[interface{}]interface{}:
		cmap := map[string]interface{}{}
		for k, v := range iType {
			cmap[k.(string)] = CanonicalizeInterface(v)
		}
		return cmap
	// We also need to recursively visit string-mapped values in case there
	// are any nested interfaces with non-comparable keys there.
	case map[string]interface{}:
		cmap := map[string]interface{}{}
		for k, v := range iType {
			cmap[k] = CanonicalizeInterface(v)
		}
		return cmap
	case []interface{}:
		for k, v := range iType {
			iType[k] = CanonicalizeInterface(v)
		}
	}

	return i
}
