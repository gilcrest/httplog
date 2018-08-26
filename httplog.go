package httplog

import (
	"database/sql"
	"net/http"
	"net/http/httptest"

	"github.com/rs/zerolog"
)

// HTTPLog records and logs as much as possible about an
// incoming HTTP request and response
func HTTPLog(log zerolog.Logger, db *sql.DB, opts *Opts) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			var (
				o   *Opts
				err error
			)

			if opts == nil {
				o, err = newHTTPLogOpts()
				if err != nil {
					log.Error().Err(err).Msg("")
					return
				}
			} else {
				o = opts
			}

			aud := new(APIAudit)

			ctx := req.Context()

			startTimer(aud)

			ctx = SetRequestID(ctx)

			err = logReqDispatch(log, aud, req, o)
			if err != nil {
				http.Error(w, "Unable to log request", http.StatusBadRequest)
				return
			}

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req.WithContext(ctx))

			// copy everything from response recorder
			// to actual response writer
			for k, v := range rec.HeaderMap {
				w.Header()[k] = v
			}
			w.WriteHeader(rec.Code)

			// pull out the response body and write it
			// back to the response writer
			b := rec.Body.Bytes()
			w.Write(b)

			stopTimer(aud)

			// write the data back to the recorder buffer as
			// it's needed for SetResponse
			rec.Body.Write(b)

			// set the response data in the APIAudit object
			err = setResponse(log, aud, rec)
			if err != nil {
				http.Error(w, "Unable to set response", http.StatusBadRequest)
			}

			// call logRespDispatch to determine if and where to log
			err = logRespDispatch(ctx, log, db, aud, o)
			if err != nil {
				http.Error(w, "Error from response dispatch", http.StatusBadRequest)
			}

		})
	}
}
