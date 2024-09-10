package client

import (
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/synackd/ochami/internal/log"
)

// BSSClient is an OchamiClient that has its BasePath set configured to the one
// that BSS uses.
type BSSClient struct {
	*OchamiClient
}

const basePathBSS = "/boot/v1"

// NewBSSClient takes a baseURI and basePath and returns a pointer to a new
// BSSClient. If an error occurred creating the embedded OchamiClient, it is
// returned.
func NewBSSClient(baseURI string) (*BSSClient, error) {
	oc, err := NewOchamiClient(baseURI, basePathBSS)
	if err != nil {
		return nil, fmt.Errorf("failed to create OchamiClient: %v", err)
	}
	bc := &BSSClient{
		OchamiClient: oc,
	}

	return bc, err
}

// MakeBSSRequest is a wrapper around MakeRequest that calls GetURI to form the
// final URI to make the request with and pass to MakeRequest.
func (bc *BSSClient) MakeBSSRequest(method, endpoint string, headers *HTTPHeaders, body HTTPBody) (*http.Response, HTTPBody, error) {
	uri, err := bc.GetURI(endpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate URI for endpoint %s: %v", endpoint, err)
	}

	return bc.MakeRequest(method, uri, headers, body)
}

// GetURI takes an endpoint and joins it with the BSSClient's BaseURI and
// BasePath to form the final URI to be used for a request.
func (bc *BSSClient) GetURI(endpoint string) (string, error) {
	uri, err := url.Parse(bc.BaseURI.String())
	if err != nil {
		return "", fmt.Errorf("failed to parse base URI %s: %v", bc.BaseURI, err)
	}
	uri.Path = path.Join(uri.Path, bc.BasePath, endpoint)
	return uri.String(), err
}

// GetData is a wrapper around MakeBSSRequest that sends a GET request to
// endpoint, using an optional token and optional headers, and returns the data
// received in the response along with a nil error. If the HTTP response code is
// unsuccessful (i.e. not 2XX), then the returned error will contain an
// UnsuccessfulHTTPError. Otherwise, the error that occurred is returned.
func (bc *BSSClient) GetData(endpoint, token string, headers *HTTPHeaders) (string, error) {
	if token != "" {
		if headers == nil {
			headers = NewHTTPHeaders()
		}
		if err := headers.SetAuthorization(token); err != nil {
			return "", fmt.Errorf("error setting token in HTTP headers: %v", err)
		}
	}

	res, resBody, err := bc.MakeBSSRequest(http.MethodGet, endpoint, headers, nil)
	if err != nil {
		return "", fmt.Errorf("error making request to BSS: %v", err)
	}
	if res != nil {
		statusOK := res.StatusCode >= 200 && res.StatusCode < 300
		if statusOK {
			log.Logger.Info().Msgf("Response status: %s %s", res.Proto, res.Status)
			return string(resBody), nil
		} else {
			if len(resBody) > 0 {
				return "", fmt.Errorf("%w: %s %s: %s", UnsuccessfulHTTPError, res.Proto, res.Status, string(resBody))
			} else {
				return "", fmt.Errorf("%w: %s %s", UnsuccessfulHTTPError, res.Proto, res.Status)
			}
		}
	}
	return "", fmt.Errorf("BSS response was empty")
}
