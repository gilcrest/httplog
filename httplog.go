// Package httplog logs http requests and responses. It’s highly configurable,
// e.g. in production, log all response and requests, but don’t log
// the body or headers, in your dev environment log everything and so
// on. httplog also has different ways to log depending on your
// preference — structured logging via JSON, relational database
// logging or just plain standard library logging. httplog has logic
// to turn on/off logging based on options you can either pass in to
// the middleware handler or from a JSON input file included with the
// library. httplog offers three middleware choices, each of which
// adhere to fairly common middleware patterns: a simple HandlerFunc
// (`LogHandlerFunc`), a function (`LogHandler`) that takes a handler
// and returns a handler (aka Constructor) (`func (http.Handler) http.Handler`)
// often used with alice (https://github.com/justinas/alice) and finally a
// function (`LogAdapter`) that returns an Adapter type. An `httplog.Adapt`
// function and `httplog.Adapter` type are provided. Beyond logging request
// and response elements, httplog creates a unique id for each incoming
// request (using xid (https://github.com/rs/xid)) and sets it (and a few
// other key request elements) into the request context. You can access
// these context items using provided helper functions, including a function
// that returns an audit struct you can add to response payloads that provide
// clients with helpful information for support.
//
// !!!!WARNING!!!! - This package works, but is something I wrote a long time
// ago and really needs to be updated. I logged Issue #8 to some day address
// this.
package httplog

import (
	"database/sql"
	"net/http"
	"net/http/httptest"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/gilcrest/httplog/errs"
)

// LogHandlerFunc middleware records and logs as much as possible about an
// incoming HTTP request and response
func LogHandlerFunc(next http.HandlerFunc, logger zerolog.Logger, db *sql.DB, o *Opts) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		var (
			opts *Opts
			err  error
		)

		// If o is passed in, then use it, otherwise opts will
		// remain its zero value and seeing as the elements within are
		// all booleans, all will be false (false is the boolean zero value)
		if o != nil {
			opts = o
		} else {
			errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, "gilcrest/httplog unsupported: nil cannot be passed currently for *Opts until bug #6 has been resolved"))
			return
			// opts, err = FileOpts()
			// if err != nil {
			// 	errStr := fmt.Sprintf("Unable to load logging options from file, error = %s", err.Error())
			// 	http.Error(w, errStr, http.StatusBadRequest)
			// 	return
			// }
		}

		// Pull the context from the request
		ctx := req.Context()

		// Create an instance of APIaudit and pass it to startTimer
		// to begin the API response timer
		ctx, aud, err := newAPIAudit(ctx, logger, req)
		if err != nil {
			errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, "Unable to log request"))
			return
		}

		aud.startTimer()

		ctx = setRequest2Context(ctx, aud)

		// RequestLogController determines which of the logging methods
		// you wish to use will be employed (based on the options passed in)
		err = requestLogController(ctx, logger, aud, req, opts)
		if err != nil {
			errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, "Unable to log request"))
			return
		}

		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, req.WithContext(ctx))

		// copy everything from response recorder
		// to actual response writer
		for k, v := range rec.Header() {
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
		err = aud.setResponse(logger, rec)
		if err != nil {
			log.Warn().Err(err).Msg("Error from setResponse in httplog")
		}

		// call responseLogController to determine if and where to log
		err = responseLogController(ctx, logger, db, aud, opts)
		if err != nil {
			log.Warn().Err(err).Msg("Error from responseLogController in httplog")
		}
	}
}

// LogHandler records and logs as much as possible about an
// incoming HTTP request and response
func LogHandler(logger zerolog.Logger, db *sql.DB, o *Opts) (mw func(http.Handler) http.Handler) {
	mw = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			var (
				opts *Opts
				err  error
			)

			// If o is passed in, then use it, otherwise opts will
			// remain its zero value and seeing as the elements within are
			// all booleans, all will be false (false is the boolean zero value)
			if o != nil {
				opts = o
			} else {
				errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, "gilcrest/httplog unsupported: nil cannot be passed currently for *Opts until bug #6 has been resolved"))
				return
				// opts, err = FileOpts()
				// if err != nil {
				// 	errStr := fmt.Sprintf("Unable to load logging options from file, error = %s", err.Error())
				// 	http.Error(w, errStr, http.StatusBadRequest)
				// 	return
				// }
			}

			// Pull the context from the request
			ctx := req.Context()

			// Create an instance of APIaudit and pass it to startTimer
			// to begin the API response timer
			ctx, aud, err := newAPIAudit(ctx, logger, req)
			if err != nil {
				errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, "Unable to log request"))
				return
			}

			aud.startTimer()

			ctx = setRequest2Context(ctx, aud)

			// RequestLogController determines which of the logging methods
			// you wish to use will be employed (based on the options passed in)
			err = requestLogController(ctx, logger, aud, req, opts)
			if err != nil {
				errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, "Unable to log request"))
				return
			}

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req.WithContext(ctx))

			// copy everything from response recorder
			// to actual response writer
			for k, v := range rec.Header() {
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
			err = aud.setResponse(logger, rec)
			if err != nil {
				log.Warn().Err(err).Msg("Error from setResponse in httplog")
			}

			// call responseLogController to determine if and where to log
			err = responseLogController(ctx, logger, db, aud, opts)
			if err != nil {
				log.Warn().Err(err).Msg("Error from responseLogController in httplog")
			}
		})
	}
	return
}

// LogAdapter records and logs as much as possible about an
// incoming HTTP request and response using the Adapter pattern
// Found adapter pattern in a Mat Ryer post
func LogAdapter(logger zerolog.Logger, db *sql.DB, o *Opts) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			var (
				opts *Opts
				err  error
			)

			// If o is passed in, then use it, otherwise opts will
			// remain its zero value and seeing as the elements within are
			// all booleans, all will be false (false is the boolean zero value)
			if o != nil {
				opts = o
			} else {
				errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, "Unable to log request"))
				return
				// opts, err = FileOpts()
				// if err != nil {
				// 	errStr := fmt.Sprintf("Unable to load logging options from file, error = %s", err.Error())
				// 	http.Error(w, errStr, http.StatusBadRequest)
				// 	return
				// }
			}

			// Pull the context from the request
			ctx := req.Context()

			// Create an instance of APIaudit and pass it to startTimer
			// to begin the API response timer
			ctx, aud, err := newAPIAudit(ctx, logger, req)
			if err != nil {
				errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, "Unable to log request"))
				return
			}
			aud.startTimer()

			ctx = setRequest2Context(ctx, aud)

			// RequestLogController determines which of the logging methods
			// you wish to use will be employed (based on the options passed in)
			// It will also populate the APIAudit struct based on the incoming request
			err = requestLogController(ctx, logger, aud, req, opts)
			if err != nil {
				errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, "Unable to log request"))
				return
			}

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req.WithContext(ctx))

			// copy everything from response recorder
			// to actual response writer
			for k, v := range rec.Header() {
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
			err = aud.setResponse(logger, rec)
			if err != nil {
				log.Warn().Err(err).Msg("Error from setResponse in httplog")
			}

			// call responseLogController to determine if and where to log
			err = responseLogController(ctx, logger, db, aud, opts)
			if err != nil {
				log.Warn().Err(err).Msg("Error from responseLogController in httplog")
			}
		})
	}
}

// Adapter type (it gets its name from the adapter pattern — also known as the
// decorator pattern) above is a function that both takes in and returns an
// http.Handler. This is the essence of the wrapper; we will pass in an existing
// http.Handler, the Adapter will adapt it, and return a new (probably wrapped)
// http.Handler for us to use in its place. So far this is not much different
// from just wrapping http.HandlerFunc types, however, now, we can instead write
// functions that themselves return an Adapter. - Mat Ryer @matryer
type Adapter func(http.Handler) http.Handler

// Adapt function takes the handler you want to adapt, and a list of our
// Adapter types. The result of any wrapper function should be an
// acceptable Adapter.  Our Adapt function will simply iterate over all
// adapters, calling them one by one (in reverse order) in a chained manner,
// returning the result of the first adapter. - Mat Ryer @matryer
func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
	for _, adapter := range adapters {
		h = adapter(h)
	}
	return h
}
