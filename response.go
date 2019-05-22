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
func responseLogController(ctx context.Context, log zerolog.Logger, db *sql.DB, t *tracker, opts *Opts) error {

	if opts.Log2StdOut.Response.Enable {
		logResp2Stdout(log, t)
	}

	if opts.Log2DB.Enable {
		err := logReqResp2Db(ctx, db, t, opts)
		if err != nil {
			log.Error().Err(err).Msg("")
			return err
		}
	}
	return nil
}

func logResp2Stdout(log zerolog.Logger, t *tracker) {

	log.Debug().Msg("logResponse started")
	defer log.Debug().Msg("logResponse ended")

	log.Info().
		Str("request_id", t.requestID).
		Int("response_code", t.responseCode).
		Str("response_header", t.response.header).
		Str("response_body", t.response.body).
		Msg("Response Sent")
}

// logReqResp2Db creates a record in the api.audit_log table
// using a stored function
func logReqResp2Db(ctx context.Context, db *sql.DB, t *tracker, opts *Opts) error {

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
		reqHdr = strNil(t.request.header)
	}
	// default reqBody variable to nil
	// if the Request Body logging option is enabled for db logging
	// then check if the header string is it's zero value and if so,
	// switch it to nil, otherwise write it to the variable
	reqBody = nil
	if opts.Log2DB.Request.Body {
		reqBody = strNil(t.request.body)
	}
	// default respHdr variable to nil
	// if the Response Header logging option is enabled for db logging
	// then check if the header string is it's zero value and if so,
	// switch it to nil, otherwise write it to the variable
	respHdr = nil
	if opts.Log2DB.Response.Header {
		respHdr = strNil(t.response.header)
	}
	// default respBody variable to nil
	// if the Response Body logging option is enabled for db logging
	// then check if the header string is it's zero value and if so,
	// switch it to nil, otherwise write it to the variable
	respBody = nil
	if opts.Log2DB.Response.Body {
		respBody = strNil(t.response.body)
	}

	// time.Duration is in nanoseconds,
	// need to do below math for milliseconds
	durMS := t.duration / time.Millisecond

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Prepare the sql statement using bind variables
	stmt, err := tx.PrepareContext(ctx, `select app.log_request (
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
		t.requestID,             //$1
		t.clientID,              //$2
		t.timeStarted,           //$3
		t.responseCode,          //$4
		t.timeFinished,          //$5
		durMS,                   //$6
		t.request.proto,         //$7
		t.request.protoMajor,    //$8
		t.request.protoMinor,    //$9
		t.request.method,        //$10
		t.request.scheme,        //$11
		t.request.host,          //$12
		t.request.port,          //$13
		t.request.path,          //$14
		t.request.remoteAddr,    //$15
		t.request.contentLength, //$16
		reqHdr,                  //$17
		reqBody,                 //$18
		respHdr,                 //$19
		respBody)                //$20

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
