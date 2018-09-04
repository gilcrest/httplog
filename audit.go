package httplog

import (
	"context"
)

// Audit struct should be used for all responses
type Audit struct {
	RequestID string    `json:"id"`
	URL       *AuditURL `json:"url"`
}

// AuditURL has url info
type AuditURL struct {
	RequestHost     string `json:"host,omitempty"`
	RequestPort     string `json:"port,omitempty"`
	RequestPath     string `json:"path,omitempty"`
	RequestRawQuery string `json:"query,omitempty"`
	RequestFragment string `json:"fragment,omitempty"`
}

// AuditOpts allow you to to turn on or off different elements
// of the URL
type AuditOpts struct {
	Host     bool
	Port     bool
	Path     bool
	Query    bool
	Fragment bool
}

// NewAudit is a constructor for the Audit struct
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
