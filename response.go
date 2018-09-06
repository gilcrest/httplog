package httplog

import (
	"context"
	"database/sql"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// logRespController determines which, if any, of the logging methods
// you wish to use will be employed
func responseLogController(ctx context.Context, log zerolog.Logger, db *sql.DB, aud *APIAudit, opts *Opts) error {

	if opts.Log2StdOut.Response.Enable {
		logResp2Stdout(log, aud)
	}

	if opts.Log2DB.Enable {
		err := logReqResp2Db(ctx, db, aud, opts)
		if err != nil {
			log.Error().Err(err).Msg("")
			return err
		}
	}
	return nil
}

func logResp2Stdout(log zerolog.Logger, aud *APIAudit) {

	log.Debug().Msg("logResponse started")
	defer log.Debug().Msg("logResponse ended")

	log.Info().
		Str("request_id", aud.RequestID).
		Int("response_code", aud.ResponseCode).
		Str("response_header", aud.response.Header).
		Str("response_body", aud.response.Body).
		Msg("Response Sent")
}

// logReqResp2Db creates a record in the api.audit_log table
// using a stored function
func logReqResp2Db(ctx context.Context, db *sql.DB, aud *APIAudit, opts *Opts) error {

	var (
		rowsInserted int
		respHdr      interface{}
		respBody     interface{}
		reqHdr       interface{}
		reqBody      interface{}
	)

	// default reqHdr variable to nil
	// if the Request Header logging option is enabled for db logging
	// then check if the header string is it's zero value and if so,
	// switch it to nil, otherwise write it to the variable
	reqHdr = nil
	if opts.Log2DB.Request.Header {
		// This empty string to nil conversion is probably
		// not necessary, but just in case to avoid db exception
		reqHdr = strNil(aud.request.Header)
	}
	// default reqBody variable to nil
	// if the Request Body logging option is enabled for db logging
	// then check if the header string is it's zero value and if so,
	// switch it to nil, otherwise write it to the variable
	reqBody = nil
	if opts.Log2DB.Request.Body {
		reqBody = strNil(aud.request.Body)
	}
	// default respHdr variable to nil
	// if the Response Header logging option is enabled for db logging
	// then check if the header string is it's zero value and if so,
	// switch it to nil, otherwise write it to the variable
	respHdr = nil
	if opts.Log2DB.Response.Header {
		respHdr = strNil(aud.response.Header)
	}
	// default respBody variable to nil
	// if the Response Body logging option is enabled for db logging
	// then check if the header string is it's zero value and if so,
	// switch it to nil, otherwise write it to the variable
	respBody = nil
	if opts.Log2DB.Response.Body {
		respBody = strNil(aud.response.Body)
	}

	// time.Duration is in nanoseconds,
	// need to do below math for milliseconds
	durMS := aud.Duration / time.Millisecond

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Prepare the sql statement using bind variables
	stmt, err := tx.PrepareContext(ctx, `select api.log_request (
		p_request_id => $1,
		p_request_timestamp => $2,
		p_response_code => $3,
		p_response_timestamp => $4,
		p_duration_in_millis => $5,
		p_protocol => $6,
		p_protocol_major => $7,
		p_protocol_minor => $8,
		p_request_method => $9,
		p_scheme => $10,
		p_host => $11,
		p_port => $12,
		p_path => $13,
		p_remote_address => $14,
		p_request_content_length => $15,
		p_request_header => $16,
		p_request_body => $17,
		p_response_header => $18,
		p_response_body => $19)`)

	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx,
		aud.RequestID,             //$1
		aud.TimeStarted,           //$2
		aud.ResponseCode,          //$3
		aud.TimeFinished,          //$4
		durMS,                     //$5
		aud.request.Proto,         //$6
		aud.request.ProtoMajor,    //$7
		aud.request.ProtoMinor,    //$8
		aud.request.Method,        //$9
		aud.request.Scheme,        //$10
		aud.request.Host,          //$11
		aud.request.Port,          //$12
		aud.request.Path,          //$13
		aud.request.RemoteAddr,    //$14
		aud.request.ContentLength, //$15
		reqHdr,   //$16
		reqBody,  //$17
		respHdr,  //$18
		respBody) //$19

	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}
	defer rows.Close()

	// Iterate through the returned record(s)
	for rows.Next() {
		if err := rows.Scan(&rowsInserted); err != nil {
			log.Error().Err(err).Msg("")
			return err
		}
	}

	err = rows.Err()
	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}

	// If we have successfully written rows to the db
	// we commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}

	return nil

}

// strNil checks if the header field is an empty string
// (the empty value for the string type) and switches it to
// a nil.  An empty string is not allowed to be passed to a
// JSONB type in postgres, however, a nil works
func strNil(s string) interface{} {
	var v interface{}

	v = s
	if s == "" {
		v = nil
	}

	return v
}
