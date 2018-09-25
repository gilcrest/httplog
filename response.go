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
		p_client_id => $2,
		p_request_timestamp => $3,
		p_response_code => $4,
		p_response_timestamp => $5,
		p_duration_in_millis => $6,
		p_protocol => $7,
		p_protocol_major => $8,
		p_protocol_minor => $9,
		p_request_method => $10,
		p_scheme => $11,
		p_host => $12,
		p_port => $13,
		p_path => $14,
		p_remote_address => $15,
		p_request_content_length => $16,
		p_request_header => $17,
		p_request_body => $18,
		p_response_header => $19,
		p_response_body => $20)`)

	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx,
		aud.RequestID,             //$1
		aud.ClientID,              //$2
		aud.TimeStarted,           //$3
		aud.ResponseCode,          //$4
		aud.TimeFinished,          //$5
		durMS,                     //$6
		aud.request.Proto,         //$7
		aud.request.ProtoMajor,    //$8
		aud.request.ProtoMinor,    //$9
		aud.request.Method,        //$10
		aud.request.Scheme,        //$11
		aud.request.Host,          //$12
		aud.request.Port,          //$13
		aud.request.Path,          //$14
		aud.request.RemoteAddr,    //$15
		aud.request.ContentLength, //$16
		reqHdr,   //$17
		reqBody,  //$18
		respHdr,  //$19
		respBody) //$20

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
