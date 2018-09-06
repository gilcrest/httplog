package httplog

import (
	"context"

	"github.com/rs/xid"
)

type contextKey string

func (c contextKey) String() string {
	return "context key " + string(c)
}

// RequestID is a unique identifier for each inbound request
var (
	requestID       = contextKey("RequestID")
	requestHost     = contextKey("RequestHost")
	requestPort     = contextKey("RequestPort")
	requestPath     = contextKey("RequestPath")
	requestRawQuery = contextKey("RequestRawQuery")
	requestFragment = contextKey("RequestFragment")
)

func setRequest2Context(ctx context.Context, aud *APIAudit) context.Context {
	ctx = SetRequestHost(ctx, aud)
	ctx = SetRequestPort(ctx, aud)
	ctx = SetRequestPath(ctx, aud)
	ctx = SetRequestRawQuery(ctx, aud)
	ctx = SetRequestFragment(ctx, aud)

	return ctx
}

// SetRequestID adds a unique ID as RequestID to the context
func SetRequestID(ctx context.Context) context.Context {
	// get byte Array representation of guid from xid package (12 bytes)
	guid := xid.New()

	// use the String method of the guid object to convert byte array to string (20 bytes)
	rID := guid.String()

	ctx = context.WithValue(ctx, requestID, rID)

	return ctx

}

// RequestID gets the Request ID from the context.
func RequestID(ctx context.Context) string {
	requestIDstr := ctx.Value(requestID).(string)
	return requestIDstr
}

// SetRequestHost adds the request host to the context
func SetRequestHost(ctx context.Context, audit *APIAudit) context.Context {
	host := audit.request.Host
	ctx = context.WithValue(ctx, requestHost, host)
	return ctx
}

// RequestHost gets the request host from the context
func RequestHost(ctx context.Context) string {
	url := ctx.Value(requestHost).(string)
	return url
}

// SetRequestPort adds the request port to the context
func SetRequestPort(ctx context.Context, audit *APIAudit) context.Context {
	port := audit.request.Port
	ctx = context.WithValue(ctx, requestPort, port)
	return ctx
}

// RequestPort gets the request port from the context
func RequestPort(ctx context.Context) string {
	url := ctx.Value(requestPort).(string)
	return url
}

// SetRequestPath adds the request URL to the context
func SetRequestPath(ctx context.Context, audit *APIAudit) context.Context {
	path := audit.request.Path
	ctx = context.WithValue(ctx, requestPath, path)
	return ctx
}

// RequestPath gets the request URL from the context
func RequestPath(ctx context.Context) string {
	url := ctx.Value(requestPath).(string)
	return url
}

// SetRequestRawQuery adds the request Query string details to the context
func SetRequestRawQuery(ctx context.Context, audit *APIAudit) context.Context {
	query := audit.request.RawQuery
	ctx = context.WithValue(ctx, requestRawQuery, query)
	return ctx
}

// RequestRawQuery gets the request Query string details from the context
func RequestRawQuery(ctx context.Context) string {
	url := ctx.Value(requestRawQuery).(string)
	return url
}

// SetRequestFragment adds the request Fragment details to the context
func SetRequestFragment(ctx context.Context, audit *APIAudit) context.Context {
	frag := audit.request.Fragment
	ctx = context.WithValue(ctx, requestFragment, frag)
	return ctx
}

// RequestFragment gets the request Fragment details from the context
func RequestFragment(ctx context.Context) string {
	url := ctx.Value(requestFragment).(string)
	return url
}
