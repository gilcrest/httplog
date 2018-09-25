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

// APIAudit struct holds the http request attributes needed
// for auditing an http request
type APIAudit struct {
	RequestID    string        `json:"request_id"`
	ClientID     string        `json:"client_id"`
	TimeStarted  time.Time     `json:"time_started"`
	TimeFinished time.Time     `json:"time_finished"`
	Duration     time.Duration `json:"time_in_millis"`
	ResponseCode int           `json:"response_code"`
	request
	response request
}

type request struct {
	Proto            string `json:"protocol"`
	ProtoMajor       int    `json:"protocol_major"`
	ProtoMinor       int    `json:"protocol_minor"`
	Method           string `json:"request_method"`
	Scheme           string `json:"scheme"`
	Host             string `json:"host"`
	Port             string `json:"port"`
	Path             string `json:"path"`
	RawQuery         string `json:"query"`
	Fragment         string `json:"fragment"`
	Header           string `json:"header"`
	Body             string `json:"body"`
	ContentLength    int64  `json:"content_length"`
	TransferEncoding string `json:"transfer_encoding"`
	Close            bool   `json:"close"`
	Trailer          string `json:"trailer"`
	RemoteAddr       string `json:"remote_address"`
	RequestURI       string `json:"request_uri"`
}

// sets the start time in the APIAudit object
func (aud *APIAudit) startTimer() {
	// set APIAudit TimeStarted to current time in UTC
	loc, _ := time.LoadLocation("UTC")
	aud.TimeStarted = time.Now().In(loc)
}

// stopTimer sets the stop time in the APIAudit object and
// subtracts the stop time from the start time to determine the
// service execution duration as this is after the response
// has been written and sent
func (aud *APIAudit) stopTimer() {
	loc, _ := time.LoadLocation("UTC")
	aud.TimeFinished = time.Now().In(loc)
	duration := aud.TimeFinished.Sub(aud.TimeStarted)
	aud.Duration = duration
}

// setResponse sets the response elements of the APIAudit payload
func (aud *APIAudit) setResponse(log zerolog.Logger, rec *httptest.ResponseRecorder) error {
	// set ResponseCode from ResponseRecorder
	aud.ResponseCode = rec.Code

	// set Header JSON from Header map in ResponseRecorder
	headerJSON, err := convertHeader(log, rec.HeaderMap)
	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}
	aud.response.Header = headerJSON

	// Dump body to text using dumpBody function - need an http request
	// struct, so use httptest.NewRequest to get one
	req := httptest.NewRequest("POST", "http://example.com/foo", rec.Body)

	body, err := dumpBody(req)
	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}
	aud.response.Body = body

	return nil
}

// splitRequest populates the APIAudit struct being passed
// as well as adds multiple request fields to the context
func newAPIAudit(ctx context.Context, log zerolog.Logger, req *http.Request) (context.Context, *APIAudit, error) {

	var (
		scheme string
	)

	aud := new(APIAudit)

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
	ctx = SetRequestID(ctx)
	aud.RequestID = RequestID(ctx)
	aud.request.Proto = req.Proto
	aud.request.ProtoMajor = req.ProtoMajor
	aud.request.ProtoMinor = req.ProtoMinor
	aud.request.Method = req.Method
	aud.request.Scheme = scheme
	aud.request.Host = host
	aud.request.Port = port
	aud.request.Path = req.URL.Path
	aud.request.RawQuery = req.URL.RawQuery
	aud.request.Fragment = req.URL.Fragment
	aud.request.Body = body
	aud.request.Header = headerJSON
	aud.request.ContentLength = req.ContentLength
	aud.request.TransferEncoding = strings.Join(req.TransferEncoding, ",")
	aud.request.Close = req.Close
	aud.request.Trailer = trailerJSON
	aud.request.RemoteAddr = req.RemoteAddr
	aud.request.RequestURI = req.RequestURI

	return ctx, aud, nil
}
