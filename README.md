# httplog - HTTP Request and Response Logging

## Installation

```go
go get -u github.com/gilcrest/httplog
```

## GoDoc

[https://godoc.org/github.com/gilcrest/httplog](https://godoc.org/github.com/gilcrest/httplog)

## External Dependencies

**httplog** has two external libraries that it depends on, listed below.

- github.com/rs/zerolog
- github.com/rs/xid

> Note: These dependencies will be included as "vendored" dependencies once I figure out modules and how to vendor dependencies within modules. This is [Issue #2](https://github.com/gilcrest/httplog/issues/2) for the library.

If you plan to use the Database Logging feature of httplog, you will need to extract the httplogDDL.sql file included in the workspace and run this on your own PostgreSQL database. This script is pretty raw right now and will be made better if there is interest.

## Overview

**httplog** logs http requests and responses. It’s highly configurable, e.g. in production, log all response and requests, but don’t log the body or headers, in your dev environment log everything and so on. httplog also has different ways to log depending on your preference — structured logging via JSON, relational database logging or just plain standard library logging.

httplog has logic to turn on/off logging based on options you can either pass in to the middleware handler or from a JSON input file included with the library.

httplog offers three middleware choices, each of which adhere to fairly common middleware patterns: a simple HandlerFunc (`LogHandlerFunc`), a function (`LogHandler`) which takes a handler and returns a handler (often used with [alice](https://github.com/justinas/alice)) and finally, a function (`LogAdapter`), which returns an Adapter type (based on [Mat Ryer’s post](https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81)). An `httplog.Adapt` function and `httplog.Adapter` type are provided for the latter.

Beyond logging request and response elements, **httplog** creates a unique id for each incoming request (using [xid](https://github.com/rs/xid)) and sets it (and a few other key request elements) into the request context. You can access these context items using provided helper functions, including a function that returns an `httplog.Audit` struct which bundles all these items for response payloads to provide clients with helpful information for support.

## Features

- [Middleware](#middleware)
- [Configurable http request/response logging](#configurable-logging) (ability to turn on and off logging style based on file configuration)
  - [Log Style 1](#log-style-1-structured-via-json): Structured (JSON), leveled (debug, error, info, etc.) logging to stdout
  - [Log Style 2](#log-style-2-relational-db-logging-via-postgreSQL): Relational database (PostgreSQL) logging (certain data points broken out into standard column datatypes, request/response headers and body stored in TEXT datatype columns).
  - [Log Style 3](#log-style-3-httputil-dumpRequest-or-dumpResponse): httputil DumpRequest or DumpResponse - there's not much to this, really - httplog just allows you to turn these standard library functions on or off through the configuration options
- [Add Unique ID and Key Request Elements to Context](#add-unique-id-and-key-request-elements-to-context)
- [Retrieve Unique ID and Key Request Elements from Context](#retrieve-unique-id-and-key-request-elements-from-context)
- [Audit Struct for Response Payload](#audit-struct-for-response-payload)

### Middleware

Each middleware takes a minimum of three parameters:

- `log` - an instance of zerolog.logger

- `db` - a pointer to a sql database (PostgreSQL)
  - You can set this parameter to nil if you are not planning to log to PostgreSQL

- `o` - an `httplog.Opts` struct which has the all of the logging configurations
  - You can set this parameter to nil and httplog will use options from the `httpLogOpt.json` file
  - If you prefer not to use the `httpLogOpt.json` file, simply initialize the `httplog.Opts` struct and all values are set to false (the whole struct is boolean flags and in Go, a boolean's zero value is false). You can then pick and choose which flags to turn on via code.

#### Middleware Examples

The below examples are taken from [go-API-template](https://github.com/gilcrest/go-API-template). It uses all three httplog middlewares for example sake. You obviously would choose 1 pattern and stick to that (for the most part).

```go
package app

import (
    "github.com/gilcrest/go-API-template/datastore"
    "github.com/gilcrest/httplog"
    "github.com/justinas/alice"
)

// routes registers handlers to the router
func (s *server) routes() error {

    // Get a logger instance from the server struct
    log := s.logger

    // Get pointer to logging database to pass into httplog
    // Only need this if you plan to use the PostgreSQL
    // logging style of httplog
    logdb, err := s.ds.DB(datastore.LogDB)
    if err != nil {
        return err
    }

    // httplog.NewOpts gets a new httplog.Opts struct
    // (with all flags set to false)
    opts := httplog.NewOpts()

    // For the examples below, I chose to turn on db logging only
    // Log the request headers only (body has password on this api!)
    // Log both the response headers and body
    opts.Log2DB.Enable = true
    opts.Log2DB.Request.Header = true
    opts.Log2DB.Response.Header = true
    opts.Log2DB.Response.Body = true

    // HandlerFunc middleware example
    // function takes an http.HandlerFunc and returns an http.HandlerFunc
    // Also, match only POST requests with Content-Type header = application/json
    s.router.HandleFunc("/v1/handlefunc/user",
        httplog.LogHandlerFunc(s.handleUserCreate(), log, logdb, opts)).
        Methods("POST").
        Headers("Content-Type", "application/json")

    // function (`LogHandler`) that takes a handler and returns a handler (aka Constructor)
    // (`func (http.Handler) http.Handler`)    - used with alice
    // Also, match only POST requests with Content-Type header = application/json
    s.router.Handle("/v1/alice/user",
        alice.New(httplog.LogHandler(log, logdb, opts)).
            ThenFunc(s.handleUserCreate())).
        Methods("POST").
        Headers("Content-Type", "application/json")

    // Adapter Type middleware example
    // Also, match only POST requests with Content-Type header = application/json
    s.router.Handle("/v1/adapter/user",
        httplog.Adapt(s.handleUserCreate(),
            httplog.LogAdapter(log, logdb, opts))).
        Methods("POST").
        Headers("Content-Type", "application/json")

    return nil
}
```

----

### Configurable Logging

The boolean fields found within the Opts struct type drive the rules for what logging features are turned on.  You can have one to three log styles turned on using this file (or none, if you so choose). Below are all the boolean options in the struct.

> Note: Right now, this struct for options serves its purpose and is pretty simple. I have read Dave Cheney's great post on [Functional Options for Friendly APIs](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis) and it's great - I may switch to this style later.

```go
// Log2StdOut
opts.Log2StdOut.Request.Enable
opts.Log2StdOut.Request.Options.Header
opts.Log2StdOut.Request.Options.Body
opts.Log2StdOut.Response.Enable
opts.Log2StdOut.Response.Options.Header
opts.Log2StdOut.Response.Options.Body

// Log2DB
opts.Log2DB.Enable
opts.Log2DB.Request.Header
opts.Log2DB.Request.Body
opts.Log2DB.Response.Header
opts.Log2DB.Response.Body

// DumpRequest
opts.HTTPUtil.DumpRequest.Body
opts.HTTPUtil.DumpRequest.Body
```

1. Pass an `Opts` struct when using one of the given middleware functions. `httplog.NewOpts` will return an Opts struct with all logging turned off. You can then set whichever logging style and option you like.
1. If you do not pass an Opts struct to one of the provided middlewares, there is code in each that will import/marshal the `httpLogOpt.json` file found in the root of the httplog library into the `Opts` struct type. You can change log configuration by altering the boolean values present in this file.

#### Log Style 1: Structured via JSON

##### JSON Request Logging

Set `log_json.Request.enable` in the [HTTP Log Config File](#log-config-file) or `opts.Log2StdOut.Request.Enable` to true in the `httplog.Opts` struct to enable http request logging as JSON (so long as you have properly "chained" the middleware).  The output for a request looks something like:

```json
{"time":1517970302,"level":"info","request_id":"b9t66vma6806ln8iak8g","header_json":"{\"Accept\":[\"*/*\"],\"Accept-Encoding\":[\"gzip, deflate\"],\"Cache-Control\":[\"no-cache\"],\"Connection\":[\"keep-alive\"],\"Content-Length\":[\"129\"],\"Content-Type\":[\"application/json\"],\"Postman-Token\":[\"9949f5e5-b406-4e22-aff3-ab6ba6e7d841\"],\"User-Agent\":[\"PostmanRuntime/7.1.1\"]}","body":"{\"username\": \"repoMan\",\"mobile_ID\": \"1-800-repoman\",\"email\":\"repoman@alwaysintense.com\",\"First_Name\":\"Otto\",\"Last_Name\":\"Maddox\"}","method":"POST","scheme":"http","host":"127.0.0.1","port":"8080","path":"/api/v1/appuser","protocol":"HTTP/1.1","proto_major":1,"proto_minor":1,"Content Length":129,"Transfer-Encoding":"","Close":false,"RemoteAddr":"127.0.0.1:58689","RequestURI":"/api/v1/appuser","message":"Request received"}
```

>NOTE - the HTTP header key:value pairs and json from the body are represented as escaped JSON within the actual message. If you don't want this data, set these fields to false in the JSON config file or `httplog.Opts` struct.

##### JSON Response Logging

Set `log_json.Response.enable` in the [HTTP Log Config File](#Log-Config-File) or `opts.Log2StdOut.Response.Enable` to true to enable http response logging as JSON. The response output will look something like:

```json
{"time":1517970302,"level":"info","request_id":"b9t66vma6806ln8iak8g","response_code":200,"response_header":"{\"Content-Type\":[\"text/plain; charset=utf-8\"],\"Request-Id\":[\"b9t66vma6806ln8iak8g\"]}","response_body":"{\"username\":\"repoMan\",\"mobile_id\":\"1-800-repoman\",\"email\":\"repoman@alwaysintense.com\",\"first_name\":\"Otto\",\"last_name\":\"Maddox\",\"create_user_id\":\"gilcrest\",\"create_date\":\"2018-02-06T21:25:02.538322Z\",\"update_user_id\":\"\",\"update_date\":\"0001-01-01T00:00:00Z\"}\n{\"username\":\"repoMan\",\"mobile_id\":\"1-800-repoman\",\"email\":\"repoman@alwaysintense.com\",\"first_name\":\"Otto\",\"last_name\":\"Maddox\",\"create_user_id\":\"gilcrest\",\"create_date\":\"2018-02-06T21:25:02.538322Z\",\"update_user_id\":\"\",\"update_date\":\"0001-01-01T00:00:00Z\"}\n","message":"Response Sent"}
```

>NOTE - same as request - the HTTP header key:value pairs and json from the body are represented as escaped JSON within the actual message. If you don't want this data, set these fields to false in the JSON config file (`httpLogOpt.json`) or `httplog.Opts` struct.

#### Log Style 2: Relational DB Logging via PostgreSQL

Set `log_2DB.enable` to true in the [HTTP Log Config File](#Log-Config-File) to enable Database logging to a PostgreSQL database.  The DDL is provided within the ddl directory (`httplogDDL.sql`) and consists of one table and one stored function. Once enabled, Request and Response information will be logged as one transaction to the database.  You can optionally choose to log request and response headers using the Options fields within the [HTTP Log Config File](#Log-Config-File) or `httplog.Opts` struct.

![Database Log](dbLog.png)

##### Logging Database Table

In total 20 fields are logged as part of the database transaction.

| Column Name   | Datatype    | Description          |
| ------------- | ----------- | -------------------- |
| request_id                | VARCHAR(100)  | Unique Request ID
| client_id                 | VARCHAR(100)  | API Client ID
| request_timestamp         | TIMESTAMP     | UTC time request received
| response_code             | INTEGER       | HTTP Response Code
| response_timestamp        | TIMESTAMP     | UTC time response sent
| duration_in_millis        | BIGINT        | Response time duration in milliseconds
| protocol                  | VARCHAR(20)   | HTTP protocol version, e.g. HTTP/1.1
| protocol_major            | INTEGER       | HTTP protocol major version
| protocol_minor            | INTEGER       | HTTP protocol minor version
| request_method            | VARCHAR(10)   | HTTP method (GET, POST, PUT, etc.)
| scheme                    | VARCHAR(100)  | URL scheme (http, https, etc.)
| host                      | VARCHAR(100)  | URL host
| port                      | VARCHAR(100)  | URL port
| path                      | VARCHAR(4000) | URL path
| remote_address            | VARCHAR(100)  | Network address which sent request
| request_content_length    | BIGINT        | Request content length
| request_header            | JSONB         | Key:Value pairs from HTTP request in JSON format
| request_body              | TEXT          | Request body content
| response_header           | JSONB         | Key:Value pairs from HTTP response in JSON format
| response_body             | TEXT          | Response body content

#### Log Style 3: httputil DumpRequest or DumpResponse

##### httputil.DumpRequest

Set `httputil.DumpRequest.enable` in the [HTTP Log Config File](#log-config-file) or `opts.HTTPUtil.DumpRequest.Enable` in `httplog.Opts` to true to enable logging the request via the [httputil.DumpRequest](https://golang.org/pkg/net/http/httputil/#DumpRequest) method. Nothing special here, really - just providing an easy way to turn this on or off.  Output typically looks like:

```bash
httputil.DumpRequest output:
POST /api/v1/appuser HTTP/1.1
Host: 127.0.0.1:8080
Accept: */*
Accept-Encoding: gzip, deflate
Cache-Control: no-cache
Connection: keep-alive
Content-Length: 129
Content-Type: application/json
Postman-Token: 6d1b2461-59e2-4c87-baf5-d8e64a93c55b
User-Agent: PostmanRuntime/7.1.1

{"username": "repoMan","mobile_ID": "1-800-repoman","email":"repoman@alwaysintense.com","First_Name":"Otto","Last_Name":"Maddox"}
```

>NOTE - in order to log the body, set `httputil.DumpRequest.body` in `httplogOpt.json` or `opts.HTTPUtil.DumpRequest.Enable` in `httplog.Opts` to true. If you don't want this data, set the appropriate field to false in the JSON config file (`httpLogOpt.json`) or `httplog.Opts` struct (depending on which method you chose).

### Add Unique ID and Key Request Elements to Context

**httplog** middleware creates a unique ID to track each request. In addition, it adds several request elements to the request context that can be accessed with helper functions later.

#### Unique Request ID

Each request is given a 20 character Unique Request ID generated by [xid](https://github.com/rs/xid). This unique ID is populated throughout each log type for easy tracking. This ID is also meant to be sent back to the client of your API either in the response header or response body (see [below](#retrieve-unique-id-and-key-request-elements-from-context) for further help on including httplog context items in a response body).

#### Other Request Elements added to Context

In addition to the generated Unique ID, httplog also adds the following request elements to the context:

- Host
- Port
- Path
- Raw Query
- Fragment

### Retrieve Unique ID and Key Request Elements from Context

In order to retrieve particular key:value pairs from the request context, the following helper functions are provided:

```go
// RequestID gets the Request ID from the context.
func RequestID(ctx context.Context) string {
```

```go
// RequestHost gets the request host from the context
func RequestHost(ctx context.Context) string {
```

```go
// RequestPort gets the request port from the context
func RequestPort(ctx context.Context) string {
```

```go
// RequestPath gets the request URL from the context
func RequestPath(ctx context.Context) string {
```

```go
// RequestRawQuery gets the request Query string details from the context
func RequestRawQuery(ctx context.Context) string {
```

```go
// RequestFragment gets the request Fragment details from the context
func RequestFragment(ctx context.Context) string {
```

### Audit Struct for Response Payload

Some APIs may find it helpful to echo back certain request elements or helpful contextual information in the response payload. **httplog** provides [httplog.Audit](https://godoc.org/github.com/gilcrest/httplog#Audit) for just this purpose. Use constructor function `httplog.NewAudit` to initialize this struct. The unique Request ID will always be sent back as part of the struct -- the other request elements are optional and can be turned on/off using the `httplog.AuditOpts` config struct. Below is a sample response with the audit struct included to give an idea of how it can be used. The example below is from the [go-API-template](https://github.com/gilcrest/go-API-template) repository which has examples of this audit struct in use.

```json
{
    "username": "15",
    "mobile_id": "1-800-repoman",
    "email": "repoman@alwaysintense.com",
    "first_name": "Otto",
    "last_name": "Maddox",
    "update_user_id": "chillcrest",
    "created": 1539138260,
    "audit": {
        "id": "beum5l708qml02e3hvag",
        "url": {
            "host": "127.0.0.1",
            "port": "8080",
            "path": "/api/v1/adapter/user",
            "query": "qskey1=fake&qskey2=test"
        }
    }
}
```

A snippet from the `handleUserCreate` handler function within [go-API-template](https://github.com/gilcrest/go-API-template) shows how to setup the `AuditOpts` struct and turn on a few options as well as plugging `httplog.Audit` into the response.

```go
    // create new AuditOpts struct and set options to true that you
    // want to see in the response body (Request ID is always present)
    aopt := new(httplog.AuditOpts)
    aopt.Host = true
    aopt.Port = true
    aopt.Path = true
    aopt.Query = true

    // get a new httplog.Audit struct from NewAudit using the
    // above set options and request context
    aud, err := httplog.NewAudit(ctx, aopt)
    if err != nil {
        err = HTTPErr{
            Code: http.StatusInternalServerError,
            Kind: errors.Other,
            Err:  err,
        }
        httpError(w, err)
        return
    }

    // create a new response struct and set Audit and other
    // relevant elements
    resp := new(response)
    resp.Audit = aud
    resp.Username = usr.Username()
    resp.MobileID = usr.MobileID()
    resp.Email = usr.Email()
    resp.FirstName = usr.FirstName()
    resp.LastName = usr.LastName()
    resp.UpdateUserID = usr.UpdateUserID()
    resp.UpdateUnixTime = usr.UpdateTimestamp().Unix()

    // Encode response struct to JSON for the response body
    json.NewEncoder(w).Encode(*resp)
```

----

### Helpful Resources I've used in this library (outside of the standard, yet amazing blog.golang.org and golang.org/doc/, etc.)

websites/youtube

- [JustforFunc](https://www.youtube.com/channel/UC_BzFbxG2za3bp5NRRRXJSw)
- [Go By Example](https://gobyexample.com/)

Books

- [Go in Action](https://www.amazon.com/Go-Action-William-Kennedy/dp/1617291781)
- [The Go Programming Language](https://www.amazon.com/Programming-Language-Addison-Wesley-Professional-Computing/dp/0134190440/ref=pd_lpo_sbs_14_t_0?_encoding=UTF8&psc=1&refRID=P9Z5CJMV36NXRZNXKG1F)

Blog/Medium Posts

- [How I write Go HTTP services after seven years](https://medium.com/statuscode/how-i-write-go-http-services-after-seven-years-37c208122831)
- [The http Handler Wrapper Technique in #golang, updated -- by Mat Ryer](https://medium.com/@matryer/the-http-handler-wrapper-technique-in-golang-updated-bc7fbcffa702)
- [Writing middleware in #golang and how Go makes it so much fun. -- by Mat Ryer](https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81)
