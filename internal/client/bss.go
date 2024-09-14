package client

import (
	"encoding/json"
	"fmt"

	"github.com/OpenCHAMI/bss/pkg/bssTypes"
)

// BSSClient is an OchamiClient that has its BasePath set configured to the one
// that BSS uses.
type BSSClient struct {
	*OchamiClient
}

const (
	serviceNameBSS = "BSS"
	basePathBSS    = "/boot/v1"

	BSSRelpathBootParams = "/bootparameters"
)

// NewBSSClient takes a baseURI and basePath and returns a pointer to a new
// BSSClient. If an error occurred creating the embedded OchamiClient, it is
// returned. If insecure is true, TLS certificates will not be verified.
func NewBSSClient(baseURI string, insecure bool) (*BSSClient, error) {
	oc, err := NewOchamiClient(serviceNameBSS, baseURI, basePathBSS, insecure)
	if err != nil {
		return nil, fmt.Errorf("failed to create OchamiClient for %s: %v", serviceNameBSS, err)
	}
	bc := &BSSClient{
		OchamiClient: oc,
	}

	return bc, err
}

func (bc *BSSClient) PostBootParams(bp bssTypes.BootParams, token string) (HTTPEnvelope, error) {
	var (
		henv    HTTPEnvelope
		headers *HTTPHeaders
		body    HTTPBody
		err     error
	)
	if body, err = json.Marshal(bp); err != nil {
		return henv, fmt.Errorf("PostBootParams(): failed to marshal BootParams: %v", err)
	}
	headers = NewHTTPHeaders()
	if token != "" {
		if err = headers.SetAuthorization(token); err != nil {
			return henv, fmt.Errorf("PostBootParams(): error setting token in HTTP headers")
		}
	}
	henv, err = bc.PostData(BSSRelpathBootParams, "", headers, body)
	if err != nil {
		err = fmt.Errorf("PostBootParams(): failed to POST boot parameters to BSS: %v", err)
	}

	return henv, err
}

func (bc *BSSClient) GetBootParams(query, token string) (HTTPEnvelope, error) {
	var (
		henv    HTTPEnvelope
		headers *HTTPHeaders
		err     error
	)
	headers = NewHTTPHeaders()
	if token != "" {
		if err = headers.SetAuthorization(token); err != nil {
			return henv, fmt.Errorf("GetBootParams(): error setting token in HTTP headers")
		}
	}
	henv, err = bc.GetData(BSSRelpathBootParams, query, headers)
	if err != nil {
		err = fmt.Errorf("GetBootParams(): error getting boot parameters: %v", err)
	}

	return henv, err
}
