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

// Log2StdOut struct holds the options for logging
// requests and responses to stdout (using zerolog)
type Log2StdOut struct {
	Request  L2SOpt
	Response L2SOpt
}

// L2SOpt is the log2StdOut Options
// Enable should be true if you want to write the log, set
// the rOpt Header and Body accordingly if you want to write those
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

// newOpts constructs an Opts struct using the httpLogOpt.json file
// included with the library
// TODO - this is probably a bad idea given modules - need to figure
// out modules and get this to work. The idea here is to have a config
// file that you can swap out on different servers - many enterprises
// will not let you touch "source code", but allow for manipulation of
// a config file like this... go figure
func newOpts() (*Opts, error) {

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

// NewOpts constructs an Opts struct
// By Default all logging is turned off
func NewOpts() *Opts {

	opts := new(Opts)

	// Log2StdOut
	opts.Log2StdOut.Request.Enable = false
	opts.Log2StdOut.Request.Options.Header = false
	opts.Log2StdOut.Request.Options.Body = false
	opts.Log2StdOut.Response.Enable = false
	opts.Log2StdOut.Response.Options.Header = false
	opts.Log2StdOut.Response.Options.Body = false

	// Log2DB
	opts.Log2DB.Enable = false
	opts.Log2DB.Request.Header = false
	opts.Log2DB.Request.Body = false
	opts.Log2DB.Response.Header = false
	opts.Log2DB.Response.Body = false

	// DumpRequest
	opts.HTTPUtil.DumpRequest.Body = false
	opts.HTTPUtil.DumpRequest.Body = false

	return opts
}
