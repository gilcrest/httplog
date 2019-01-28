package httplog

import (
	"context"
)

// Audit struct can be included as part of an http response body.
// This struct sends back the Unique ID generated upon receiving a
// request as well as echoes back the URL information to help with
// debugging.
type Audit struct {
	RequestID string   `json:"id"`
	URL       AuditURL `json:"url,omitempty"`
}

// AuditURL has URL info which can be included as part of an http response body
type AuditURL struct {
	RequestHost     string `json:"host,omitempty"`
	RequestPort     string `json:"port,omitempty"`
	RequestPath     string `json:"path,omitempty"`
	RequestRawQuery string `json:"query,omitempty"`
	RequestFragment string `json:"fragment,omitempty"`
}

// NewAudit is a constructor for the Audit struct. Elements added to the
// context through the provided middleware functions can be retrieved
// through the various helper functions or if you prefer in this one
// convenient struct.
func NewAudit(ctx context.Context) (*Audit, error) {

	audit := new(Audit)
	id, err := RequestID(ctx)
	if err == nil {
		audit.RequestID = id
	}
	aurl := AuditURL{}

	host, err := RequestHost(ctx)
	if err == nil {
		aurl.RequestHost = host

	}

	post, err := RequestPort(ctx)
	if err == nil {
		aurl.RequestPort = post
	}

	path, err := RequestPath(ctx)
	if err == nil {
		aurl.RequestPath = path
	}

	rrq, err := RequestRawQuery(ctx)
	if err == nil {
		aurl.RequestRawQuery = rrq
	}

	audit.URL = aurl

	return audit, nil
}
