package httplog

import (
	"context"
	"errors"
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

// AuditOpts allow you to to turn on or off different elements
// of the AuditURL.
type AuditOpts struct {
	Host     bool
	Port     bool
	Path     bool
	Query    bool
	Fragment bool
}

// NewAudit is a constructor for the Audit struct. Elements added to the
// context through the provided middleware functions can be retrieved
// through the various helper functions or if you prefer in this one
// convenient struct. If you just pass an initialized AuditOpt struct to the
// opts parameter, all options are set to false. All options are boolean flags,
// and the zero value for booleans is false. For the elements you want to appear
// in the response, pass in the AuditOpts struct with that element set to true
func NewAudit(ctx context.Context, opts *AuditOpts) (*Audit, error) {

	if opts == nil {
		return nil, errors.New("opts (*AuditOpts) parameter cannot be nil")
	}

	audit := new(Audit)
	aurl := AuditURL{}

	audit.RequestID = RequestID(ctx)

	if opts.Host {
		aurl.RequestHost = RequestHost(ctx)
	}
	if opts.Port {
		aurl.RequestPort = RequestPort(ctx)
	}
	if opts.Path {
		aurl.RequestPath = RequestPath(ctx)
	}
	if opts.Query {
		aurl.RequestRawQuery = RequestRawQuery(ctx)
	}
	if opts.Fragment {
		aurl.RequestFragment = RequestFragment(ctx)
	}

	audit.URL = aurl

	return audit, nil
}
