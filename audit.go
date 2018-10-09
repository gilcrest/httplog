package httplog

import (
	"context"
)

// Audit struct can be included as part of an http response body.
// This struct sends back the Unique ID generated upon receiving a
// request as well as echoes back the URL information to help with
// debugging.
type Audit struct {
	RequestID string    `json:"id"`
	URL       *AuditURL `json:"url"`
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
// convenient struct. If nil is passed for the opts parameter, all options
// are set to true. If you prefer not to have a particular element of the URL
// in the response, pass in the AuditOpts struct with that element set to false
func NewAudit(ctx context.Context, opts *AuditOpts) (*Audit, error) {
	audit := new(Audit)
	aurl := new(AuditURL)
	o := new(AuditOpts)

	if opts == nil {
		o.Host = true
		o.Port = true
		o.Path = true
		o.Query = true
		o.Fragment = true
	} else {
		o = opts
	}

	audit.RequestID = RequestID(ctx)

	if o.Host {
		aurl.RequestHost = RequestHost(ctx)
	}
	if o.Port {
		aurl.RequestPort = RequestPort(ctx)
	}
	if o.Path {
		aurl.RequestPath = RequestPath(ctx)
	}
	if o.Query {
		aurl.RequestRawQuery = RequestRawQuery(ctx)
	}
	if o.Fragment {
		aurl.RequestFragment = RequestFragment(ctx)
	}

	audit.URL = aurl

	return audit, nil
}
