package client

import (
	"fmt"
)

type HTTPHeaders map[string]string
type HTTPBody []byte

var (
	UnsuccessfulHTTPError = fmt.Errorf("unsuccessful HTTP status")
	NilMapPointerError    = fmt.Errorf("nil map pointer")
)

// NewHTTPHeaders returns a pointer to a new HTTPHeaders.
func NewHTTPHeaders() *HTTPHeaders {
	return &HTTPHeaders{}
}

// SetAuthorization takes a token and adds it as an authentication header to the
// HTTPHeaders map. If the HTTPHeaders map is nil, an error is returned.
func (h *HTTPHeaders) SetAuthorization(token string) error {
	if h == nil {
		return NilMapPointerError
	}
	(*h)["Authorization"] = fmt.Sprintf("Bearer %s", token)
	return nil
}

// SetContentType takes a content type string (e.g. "text/plain") and sets the
// "Content-Type" header to it in the HTTPHeaders map.
func (h *HTTPHeaders) SetContentType(ct string) error {
	if h == nil {
		return NilMapPointerError
	}
	(*h)["Content-Type"] = ct
	return nil
}
