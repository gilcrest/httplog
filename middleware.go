package httplog

import (
	"database/sql"
	"net/http"
	"net/http/httptest"

	"github.com/rs/zerolog"
)

// LogHandlerFunc middleware records and logs as much as possible about an
// incoming HTTP request and response
func LogHandlerFunc(next http.HandlerFunc, log zerolog.Logger, db *sql.DB, o *Opts) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		var (
			opts *Opts
			err  error
		)

		if o == nil {
			opts, err = newOpts()
			if err != nil {
				log.Error().Err(err).Msg("")
				return
			}
		} else {
			opts = o
		}

		// Pull the context from the request
		ctx := req.Context()

		// Create an instance of APIaudit and pass it to startTimer
		// to begin the API response timer
		ctx, aud, err := newAPIAudit(ctx, log, req)
		if err != nil {
			http.Error(w, "Unable to log request", http.StatusBadRequest)
			return
		}

		aud.startTimer()

		ctx = setRequest2Context(ctx, aud)

		// RequestLogController determines which of the logging methods
		// you wish to use will be employed (based on the options passed in)
		err = requestLogController(ctx, log, aud, req, opts)
		if err != nil {
			http.Error(w, "Unable to log request", http.StatusBadRequest)
			return
		}

		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, req.WithContext(ctx))

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

		aud.stopTimer()

		// write the data back to the recorder buffer as
		// it's needed for SetResponse
		rec.Body.Write(b)

		// set the response data in the APIAudit object
		err = aud.setResponse(log, rec)
		if err != nil {
			http.Error(w, "Unable to set response", http.StatusBadRequest)
		}

		// call responseLogController to determine if and where to log
		err = responseLogController(ctx, log, db, aud, opts)
		if err != nil {
			http.Error(w, "Error from responseLogController", http.StatusBadRequest)
		}
	}
}

// LogHandler records and logs as much as possible about an
// incoming HTTP request and response
func LogHandler(log zerolog.Logger, db *sql.DB, o *Opts) (mw func(http.Handler) http.Handler) {
	mw = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			var (
				opts *Opts
				err  error
			)

			if o == nil {
				opts, err = newOpts()
				if err != nil {
					log.Error().Err(err).Msg("")
					return
				}
			} else {
				opts = o
			}

			// Pull the context from the request
			ctx := req.Context()

			// Create an instance of APIaudit and pass it to startTimer
			// to begin the API response timer
			ctx, aud, err := newAPIAudit(ctx, log, req)
			if err != nil {
				http.Error(w, "Unable to log request", http.StatusBadRequest)
				return
			}

			aud.startTimer()

			ctx = setRequest2Context(ctx, aud)

			// RequestLogController determines which of the logging methods
			// you wish to use will be employed (based on the options passed in)
			err = requestLogController(ctx, log, aud, req, opts)
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

			aud.stopTimer()

			// write the data back to the recorder buffer as
			// it's needed for SetResponse
			rec.Body.Write(b)

			// set the response data in the APIAudit object
			err = aud.setResponse(log, rec)
			if err != nil {
				http.Error(w, "Unable to set response", http.StatusBadRequest)
			}

			// call responseLogController to determine if and where to log
			err = responseLogController(ctx, log, db, aud, opts)
			if err != nil {
				http.Error(w, "Error from responseLogController", http.StatusBadRequest)
			}
		})
	}
	return
}

// LogAdapter records and logs as much as possible about an
// incoming HTTP request and response using the Adapter pattern
func LogAdapter(log zerolog.Logger, db *sql.DB, o *Opts) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			var (
				opts *Opts
				err  error
			)

			if o == nil {
				opts, err = newOpts()
				if err != nil {
					log.Error().Err(err).Msg("")
					return
				}
			} else {
				opts = o
			}

			// Pull the context from the request
			ctx := req.Context()

			// Create an instance of APIaudit and pass it to startTimer
			// to begin the API response timer
			ctx, aud, err := newAPIAudit(ctx, log, req)
			if err != nil {
				http.Error(w, "Unable to log request", http.StatusBadRequest)
				return
			}
			aud.startTimer()

			ctx = setRequest2Context(ctx, aud)

			// RequestLogController determines which of the logging methods
			// you wish to use will be employed (based on the options passed in)
			// It will also populate the APIAudit struct based on the incoming request
			err = requestLogController(ctx, log, aud, req, opts)
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

			aud.stopTimer()

			// write the data back to the recorder buffer as
			// it's needed for SetResponse
			rec.Body.Write(b)

			// set the response data in the APIAudit object
			err = aud.setResponse(log, rec)
			if err != nil {
				http.Error(w, "Unable to set response", http.StatusBadRequest)
			}

			// call responseLogController to determine if and where to log
			err = responseLogController(ctx, log, db, aud, opts)
			if err != nil {
				http.Error(w, "Error from responseLogController", http.StatusBadRequest)
			}
		})
	}
}
