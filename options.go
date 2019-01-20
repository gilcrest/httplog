package httplog

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Opts represent HTTP Logging Options
type Opts struct {
	Log2StdOut Log2StdOut `json:"log_json"`
	Log2DB     Log2DB     `json:"log_2DB"`
	HTTPUtil   HTTPUtil   `json:"httputil"`
}

// HTTPUtil struct hold the options for using
// the net/http/httputil package
type HTTPUtil struct {
	DumpRequest DumpRequest
}

// DumpRequest holds the options for the
// net/http/httputil.DumpRequest function
type DumpRequest struct {
	Enable bool `json:"enable"`
	Body   bool `json:"body"`
}

// Log2StdOut (Log to Standard Output) struct holds the options
// for logging requests and responses to stdout (using zerolog)
type Log2StdOut struct {
	Request  L2SOpt
	Response L2SOpt
}

// L2SOpt struct holds the log2StdOut options.
// Enable should be true if you want httplog to write to stdout, set
// the ROpt Header and Body booleans accordingly if you want to write
// those
type L2SOpt struct {
	Enable  bool `json:"enable"`
	Options ROpt
}

// Log2DB struct holds the options for logging to a database
// Set Enable to true you want any database logging
// Set the Request and Response options according to whether
// you want to log request and/or response to the database
// Requests/Responses will only be logged if Enable is true
type Log2DB struct {
	Enable   bool `json:"enable"`
	Request  ROpt
	Response ROpt
}

// ROpt is the http request/response logging options
// choose whether you want to log the http headers or body
// by setting either value to true
type ROpt struct {
	Header bool `json:"header"`
	Body   bool `json:"body"`
}

// FileOpts constructs an Opts struct using the httpLogOpt.json file
// included with the library
// TODO - relying on GOPATH is a bad idea given modules - need to figure
// out modules and figure out a better way. The idea here is to have a config
// file that you can swap out on different servers - many enterprises
// will not let you touch "source code", but allow for manipulation of
// a config file like this... go figure
func FileOpts() (*Opts, error) {

	g := os.Getenv("GOPATH")
	filepath := g + "/src/github.com/gilcrest/httplog/httpLogOpt.json"

	raw, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var l Opts
	if err := json.Unmarshal(raw, &l); err != nil {
		return nil, err
	}

	return &l, nil
}

type option func(*Opts)

// Option sets the options specified.
func (o *Opts) Option(opts ...option) {
	for _, opt := range opts {
		opt(o)
	}
}

// LogRequest2Stdout sets the options for logging http requests to
// Standard Output (stdout).
// enable turns on the functionality
// header logs http request headers
// body logs the http request body
func LogRequest2Stdout(enable bool, header bool, body bool) option {
	return func(o *Opts) {
		o.Log2StdOut.Request.Enable = enable
		o.Log2StdOut.Request.Options.Header = header
		o.Log2StdOut.Request.Options.Body = body
	}
}

// LogResponse2Stdout sets the options for logging http responses to
// Standard Output (stdout).
// enable turns on the functionality
// header logs http response headers
// body logs the http response body
func LogResponse2Stdout(enable bool, header bool, body bool) option {
	return func(o *Opts) {
		o.Log2StdOut.Response.Enable = enable
		o.Log2StdOut.Response.Options.Header = header
		o.Log2StdOut.Response.Options.Body = body
	}
}

// Log2Database sets the options for logging to the database.
// enable turns on the functionality - if this is set to false, the
// parameters afterward are irrelevant as nothing will log.
// reqHdr logs http request headers
// reqBody logs the http request body
// respHdr logs http response headers
// respBody logs the http response body
func Log2Database(enable bool, reqHdr bool, reqBody bool, respHdr bool, respBody bool) option {
	return func(o *Opts) {
		o.Log2DB.Enable = enable
		o.Log2DB.Request.Header = reqHdr
		o.Log2DB.Request.Body = reqHdr
		o.Log2DB.Response.Header = respHdr
		o.Log2DB.Response.Body = respBody
	}
}

// LogRequestViaHTTPUtil sets the options for logging requests
// using the standard HTTPUtil package
// enable turns on the functionality
// body logs the http request body
// This is just wrapper functionality and in hindsight is really
// kinda silly to include, but alas...
func LogRequestViaHTTPUtil(enable bool, body bool) option {
	return func(o *Opts) {
		o.HTTPUtil.DumpRequest.Enable = enable
		o.HTTPUtil.DumpRequest.Body = body
	}
}
