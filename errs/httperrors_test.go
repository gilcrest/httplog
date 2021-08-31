package errs

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/gilcrest/httplog/logger"
)

func Test_httpErrorStatusCode(t *testing.T) {
	type args struct {
		k Kind
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"Exist", args{k: Exist}, http.StatusBadRequest},
		{"NotExist", args{k: NotExist}, http.StatusBadRequest},
		{"Invalid", args{k: Invalid}, http.StatusBadRequest},
		{"Private", args{k: Private}, http.StatusBadRequest},
		{"BrokenLink", args{k: BrokenLink}, http.StatusBadRequest},
		{"Validation", args{k: Validation}, http.StatusBadRequest},
		{"InvalidRequest", args{k: InvalidRequest}, http.StatusBadRequest},
		{"Other", args{k: Other}, http.StatusInternalServerError},
		{"IO", args{k: IO}, http.StatusInternalServerError},
		{"Internal", args{k: Internal}, http.StatusInternalServerError},
		{"Database", args{k: Database}, http.StatusInternalServerError},
		{"Unanticipated", args{k: Unanticipated}, http.StatusInternalServerError},
		{"Default", args{k: 99}, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := httpErrorStatusCode(tt.args.k); got != tt.want {
				t.Errorf("httpErrorStatusCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPErrorResponse_StatusCode(t *testing.T) {

	type args struct {
		w   *httptest.ResponseRecorder
		l   zerolog.Logger
		err error
	}

	l := logger.NewLogger(os.Stdout, zerolog.DebugLevel, false)

	gErr := E("some authentication error from google")
	aErr := E("some authorization error")
	unauthenticatedErr := &UnauthenticatedError{WWWAuthenticateRealm: "DeepInTheRealm", Err: gErr}
	unauthorizedErr := &UnauthorizedError{Err: aErr}

	tests := []struct {
		name string
		args args
		want int
	}{
		{"nil error", args{httptest.NewRecorder(), l, nil}, http.StatusInternalServerError},
		{"empty *Error", args{httptest.NewRecorder(), l, &Error{}}, http.StatusInternalServerError},
		{"unauthenticated", args{httptest.NewRecorder(), l, unauthenticatedErr}, http.StatusUnauthorized},
		{"unauthorized", args{httptest.NewRecorder(), l, unauthorizedErr}, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HTTPErrorResponse(tt.args.w, l, tt.args.err)
			if got := tt.args.w.Result().StatusCode; got != tt.want {
				t.Errorf("httpErrorStatusCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPErrorResponse_Body(t *testing.T) {

	type args struct {
		w   *httptest.ResponseRecorder
		l   zerolog.Logger
		err error
	}

	var b bytes.Buffer
	l := logger.NewLogger(&b, zerolog.DebugLevel, false)

	gErr := E("some authentication error from google")
	aErr := E("some authorization error")
	unauthenticatedErr := &UnauthenticatedError{WWWAuthenticateRealm: "DeepInTheRealm", Err: gErr}
	unauthorizedErr := &UnauthorizedError{Err: aErr}

	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty Error", args{httptest.NewRecorder(), l, &Error{}}, ""},
		{"unauthenticated", args{httptest.NewRecorder(), l, unauthenticatedErr}, ""},
		{"unauthorized", args{httptest.NewRecorder(), l, unauthorizedErr}, ""},
		{"normal", args{httptest.NewRecorder(), l, E(Exist, Parameter("some_param"), Code("some_code"), errors.New("some error"))}, `{"error":{"kind":"item_already_exists","code":"some_code","param":"some_param","message":"some error"}}`},
		{"not via E", args{httptest.NewRecorder(), l, errors.New("some error")}, "{\"error\":{\"kind\":\"unanticipated_error\",\"code\":\"Unanticipated\",\"message\":\"Unexpected error - contact support\"}}"},
		{"nil error", args{httptest.NewRecorder(), l, nil}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HTTPErrorResponse(tt.args.w, l, tt.args.err)
			if got := strings.TrimSpace(tt.args.w.Body.String()); got != tt.want {
				t.Errorf("httpErrorStatusCode() = %v, want %v", got, tt.want)
			}
		})
	}
}
