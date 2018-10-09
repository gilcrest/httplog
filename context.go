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

func setRequest2Context(ctx context.Context, aud *tracker) context.Context {
	ctx = setRequestHost(ctx, aud)
	ctx = setRequestPort(ctx, aud)
	ctx = setRequestPath(ctx, aud)
	ctx = setRequestRawQuery(ctx, aud)
	ctx = setRequestFragment(ctx, aud)

	return ctx
}

// SetRequestID adds a unique ID as RequestID to the context
func setRequestID(ctx context.Context) context.Context {
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
func setRequestHost(ctx context.Context, audit *tracker) context.Context {
	host := audit.request.host
	ctx = context.WithValue(ctx, requestHost, host)
	return ctx
}

// RequestHost gets the request host from the context
func RequestHost(ctx context.Context) string {
	url := ctx.Value(requestHost).(string)
	return url
}

// SetRequestPort adds the request port to the context
func setRequestPort(ctx context.Context, audit *tracker) context.Context {
	port := audit.request.port
	ctx = context.WithValue(ctx, requestPort, port)
	return ctx
}

// RequestPort gets the request port from the context
func RequestPort(ctx context.Context) string {
	url := ctx.Value(requestPort).(string)
	return url
}

// SetRequestPath adds the request URL to the context
func setRequestPath(ctx context.Context, audit *tracker) context.Context {
	path := audit.request.path
	ctx = context.WithValue(ctx, requestPath, path)
	return ctx
}

// RequestPath gets the request URL from the context
func RequestPath(ctx context.Context) string {
	url := ctx.Value(requestPath).(string)
	return url
}

// SetRequestRawQuery adds the request Query string details to the context
func setRequestRawQuery(ctx context.Context, audit *tracker) context.Context {
	query := audit.request.rawQuery
	ctx = context.WithValue(ctx, requestRawQuery, query)
	return ctx
}

// RequestRawQuery gets the request Query string details from the context
func RequestRawQuery(ctx context.Context) string {
	url := ctx.Value(requestRawQuery).(string)
	return url
}

// SetRequestFragment adds the request Fragment details to the context
func setRequestFragment(ctx context.Context, audit *tracker) context.Context {
	frag := audit.request.fragment
	ctx = context.WithValue(ctx, requestFragment, frag)
	return ctx
}

// RequestFragment gets the request Fragment details from the context
func RequestFragment(ctx context.Context) string {
	url := ctx.Value(requestFragment).(string)
	return url
}
