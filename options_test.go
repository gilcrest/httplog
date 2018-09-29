package httplog

import (
	"reflect"
	"testing"
)

func Test_FileOpts(t *testing.T) {

	opts := new(Opts)

	tests := []struct {
		name    string
		want    *Opts
		wantErr bool
	}{
		{"Initialize all to false", opts, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FileOpts()
			if (err != nil) != tt.wantErr {
				t.Errorf("newOpts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FileOpts() = %v, want %v", got, tt.want)
			}
		})
	}
}
