package httplog

import (
	"context"
	"reflect"
	"testing"
)

func TestNewAudit(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts *AuditOpts
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, requestID, "test123")
	// ctx = context.WithValue(ctx, requestHost, "testhost")

	opt := new(AuditOpts)
	// opt.Host = false

	arg := args{ctx, opt}

	aud := new(Audit)
	aud.RequestID = "test123"
	aud.URL = AuditURL{}

	tests := []struct {
		name    string
		args    args
		want    *Audit
		wantErr bool
	}{
		{"Test 1", arg, aud, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAudit(tt.args.ctx, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAudit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAudit() = %v, want %v", got, tt.want)
			}
		})
	}
}
