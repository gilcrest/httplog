package httplog

import (
	"reflect"
	"testing"
)

func Test_newOpts(t *testing.T) {

	opts := testOpts()

	tests := []struct {
		name    string
		want    *Opts
		wantErr bool
	}{
		{"Initialize all to false", opts, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newOpts()
			if (err != nil) != tt.wantErr {
				t.Errorf("newOpts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newOpts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewOpts(t *testing.T) {

	opts := testOpts()

	tests := []struct {
		name string
		want *Opts
	}{
		{"Initialize all to false", opts},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewOpts(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewOpts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func testOpts() *Opts {
	opts := new(Opts)

	opts.Log2StdOut.Request.Enable = false
	opts.Log2StdOut.Request.Options.Header = false
	opts.Log2StdOut.Request.Options.Body = false
	opts.Log2StdOut.Response.Enable = false
	opts.Log2StdOut.Response.Options.Header = false
	opts.Log2StdOut.Response.Options.Body = false

	opts.Log2DB.Enable = false
	opts.Log2DB.Request.Header = false
	opts.Log2DB.Request.Body = false
	opts.Log2DB.Response.Header = false
	opts.Log2DB.Response.Body = false

	opts.HTTPUtil.DumpRequest.Enable = false
	opts.HTTPUtil.DumpRequest.Body = false

	return opts
}
