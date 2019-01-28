package httplog

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// tracker struct holds the http request attributes needed
// for auditing an http request
type tracker struct {
	requestID    string
	clientID     string
	timeStarted  time.Time
	timeFinished time.Time
	duration     time.Duration
	responseCode int
	request
	response request
}

type request struct {
	proto            string
	protoMajor       int
	protoMinor       int
	method           string
	scheme           string
	host             string
	port             string
	path             string
	rawQuery         string
	fragment         string
	header           string
	body             string
	contentLength    int64
	transferEncoding string
	close            bool
	trailer          string
	remoteAddr       string
	requestURI       string
}

// sets the start time in the APIAudit object
func (t *tracker) startTimer() {
	// set APIAudit TimeStarted to current time in UTC
	loc, _ := time.LoadLocation("UTC")
	t.timeStarted = time.Now().In(loc)
}

// stopTimer sets the stop time in the APIAudit object and
// subtracts the stop time from the start time to determine the
// service execution duration as this is after the response
// has been written and sent
func (t *tracker) stopTimer() {
	loc, _ := time.LoadLocation("UTC")
	t.timeFinished = time.Now().In(loc)
	duration := t.timeFinished.Sub(t.timeStarted)
	t.duration = duration
}

// setResponse sets the response elements of the APIAudit payload
func (t *tracker) setResponse(log zerolog.Logger, rec *httptest.ResponseRecorder) error {
	// set ResponseCode from ResponseRecorder
	t.responseCode = rec.Code

	// set Header JSON from Header map in ResponseRecorder
	headerJSON, err := convertHeader(log, rec.HeaderMap)
	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}
	t.response.header = headerJSON

	// Dump body to text using dumpBody function - need an http request
	// struct, so use httptest.NewRequest to get one
	req := httptest.NewRequest("POST", "http://example.com/foo", rec.Body)

	body, err := dumpBody(req)
	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}
	t.response.body = body

	return nil
}

// splitRequest populates the APIAudit struct being passed
// as well as adds multiple request fields to the context
func newAPIAudit(ctx context.Context, log zerolog.Logger, req *http.Request) (context.Context, *tracker, error) {

	var (
		scheme string
	)

	t := new(tracker)

	// split host and port out for cleaner logging
	host, port, err := net.SplitHostPort(req.Host)
	if err != nil {
		log.Error().Err(err).Msg("")
		return ctx, nil, err
	}

	// determine if the request is an HTTPS request
	isHTTPS := req.TLS != nil

	if isHTTPS {
		scheme = "https"
	} else {
		scheme = "http"
	}

	// convert the Header map from the request to a JSON string
	headerJSON, err := convertHeader(log, req.Header)
	if err != nil {
		log.Error().Err(err).Msg("")
		return ctx, nil, err
	}

	// convert the Trailer map from the request to a JSON string
	trailerJSON, err := convertHeader(log, req.Trailer)
	if err != nil {
		log.Error().Err(err).Msg("")
		return ctx, nil, err
	}

	body, err := dumpBody(req)
	if err != nil {
		log.Error().Err(err).Msg("")
		return ctx, nil, err
	}

	// Sets a Unique ID into the context
	ctx = setRequestID(ctx)
	id, err := RequestID(ctx)
	if err != nil {
		log.Error().Err(err).Msg("")
		return ctx, nil, err
	}
	t.requestID = id
	t.request.proto = req.Proto
	t.request.protoMajor = req.ProtoMajor
	t.request.protoMinor = req.ProtoMinor
	t.request.method = req.Method
	t.request.scheme = scheme
	t.request.host = host
	t.request.port = port
	t.request.path = req.URL.Path
	t.request.rawQuery = req.URL.RawQuery
	t.request.fragment = req.URL.Fragment
	t.request.body = body
	t.request.header = headerJSON
	t.request.contentLength = req.ContentLength
	t.request.transferEncoding = strings.Join(req.TransferEncoding, ",")
	t.request.close = req.Close
	t.request.trailer = trailerJSON
	t.request.remoteAddr = req.RemoteAddr
	t.request.requestURI = req.RequestURI

	return ctx, t, nil
}
