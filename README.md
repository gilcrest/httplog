# httplog - HTTP Request and Response Logging

## Installation

```go
go get -u github.com/gilcrest/httplog
```

If you plan to use the Database Logging feature of httplog, you will need to extract the httplogDDL.sql file included in the workspace and run this on your own PostgreSQL database.

## Overview

httplog logs http requests and responses. It’s highly configurable, e.g. in production, log all response and requests, but don’t log the body or headers, in your dev environment log everything and so on. httplog also has different ways to log depending on your preference — structured logging via JSON, relational database logging or just plain standard library logging.

httplog has logic to turn on/off logging based on options you can either pass in to the middleware handler or from a JSON input file included with the library.

httplog offers three middleware choices, each of which adhere to fairly common middleware patterns: a simple HandlerFunc (`LogHandlerFunc`), a function (`LogHandler`) that takes a handler and returns a handler (aka Constructor) (`func (http.Handler) http.Handler`) often used with [alice](https://github.com/justinas/alice) and finally a function (`LogAdapter`) that returns an Adapter type (based on [Mat Ryer’s post](https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81)). An `httplog.Adapt` function and `httplog.Adapter` type are provided.

### Middleware

Each middleware takes the same parameters:

- `log` - an instance of zerolog.logger

- `db` - a pointer to a sql database (PostgreSQL)
  - You can set this parameter to nil if you're not planning to log to PostgreSQL

- `o` - an `httplog.Opts` struct which has the all of the logging configurations
  - You can set this parameter to nil and httplog will pull defaults from the provided json file
  - Use the `httplog.NewOpts` constructor to get an Opts struct with all log flags set to false

```go
// Log2StdOut
opts.Log2StdOut.Request.Enable = false
opts.Log2StdOut.Request.Options.Header = false
opts.Log2StdOut.Request.Options.Body = false
opts.Log2StdOut.Response.Enable = false
opts.Log2StdOut.Response.Options.Header = false
opts.Log2StdOut.Response.Options.Body = false

// Log2DB
opts.Log2DB.Enable = false
opts.Log2DB.Request.Header = false
opts.Log2DB.Request.Body = false
opts.Log2DB.Response.Header = false
opts.Log2DB.Response.Body = false

// DumpRequest
opts.HTTPUtil.DumpRequest.Body = false
opts.HTTPUtil.DumpRequest.Body = false
```

#### Middleware Examples

The below examples are taken from [go-API-template](https://github.com/gilcrest/go-API-template). It uses all three middlewares for example sake. You obviously would choose 1 pattern and stick to that (for the most part).

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

    // Get logging Database to pass into httplog
    // Only need this if you plan to use the PostgreSQL
    // logging style of httplog
    logdb, err := s.ds.DB(datastore.LogDB)
    if err != nil {
        return err
    }

    // HandlerFunc middleware example
    // match only POST requests on /v1/adapter/user
    // having a Content-Type header = application/json
    s.router.HandleFunc("/v1/handlefunc/user",
        httplog.LogHandlerFunc(s.handleUserCreate(), log, logdb, nil)).
        Methods("POST").
        Headers("Content-Type", "application/json")

    s.router.Handle("/v1/alice/user",
        alice.New(httplog.LogHandler(log, logdb, nil)).
            ThenFunc(s.handleUserCreate())).
        Methods("POST").
        Headers("Content-Type", "application/json")

    // match only POST requests on /v1/adapter/user
    // having a Content-Type header = application/json
    s.router.Handle("/v1/adapter/user",
        httplog.Adapt(s.handleUserCreate(),
            httplog.LogAdapter(log, logdb, nil))).
        Methods("POST").
        Headers("Content-Type", "application/json")

    return nil
}
```

## Features

- [configurable http request/response logging](#configurable-logging) (ability to turn on and off logging style based on file configuration)
  - [Log Style 1](#log-style-1-structured-via-json): Structured (JSON), leveled (debug, error, info, etc.) logging to stdout
  - [Log Style 2](#log-style-2-relational-db-logging-via-postgreSQL): Relational database (PostgreSQL) logging (certain data points broken out into standard column datatypes, request/response headers and body stored in TEXT datatype columns).
  - [Log Style 3](#log-style-3-httputil-dumpRequest-or-dumpResponse): httputil DumpRequest or DumpResponse - there's not much to this, really - httplog just allows you to turn these standard library functions on or off through the configuration options

----

### Configurable Logging

Configurable http request/response logging is achieved in one of two ways:

1. Pass an `Opts` struct when using one of the given middleware functions.
- `httplog.NewOpts` will return an Opts struct with all logging turned off. You can then choose whichever logging style and option you like.

1. through import/marshaling of a JSON file into the struct type `HTTPLogOpts`. The boolean fields found within this type drive the rules for what logging features are turned on.  You can have one to three log styles turned on using this file (or none, if you so choose).  I will eventually make this dynamic using some type of cacheing mechanism for the various choices.

#### Unique Request ID

Each request is given a 20 character Unique Request ID generated by [xid](https://github.com/rs/xid). This unique ID is added to the Request Response Header as well as populated throughout each log type for easy tracking.

#### Log Style 1: Structured via JSON

##### JSON Request Logging

Set `log_json.Request.enable` to true in the [HTTP Log Config File](#log-config-file) to enable http request logging as JSON (so long as you have properly "chained" the LogRequest handler/adapter middleware).  The output for a request looks something like:

```json
{"time":1517970302,"level":"info","request_id":"b9t66vma6806ln8iak8g","header_json":"{\"Accept\":[\"*/*\"],\"Accept-Encoding\":[\"gzip, deflate\"],\"Cache-Control\":[\"no-cache\"],\"Connection\":[\"keep-alive\"],\"Content-Length\":[\"129\"],\"Content-Type\":[\"application/json\"],\"Postman-Token\":[\"9949f5e5-b406-4e22-aff3-ab6ba6e7d841\"],\"User-Agent\":[\"PostmanRuntime/7.1.1\"]}","body":"{\"username\": \"repoMan\",\"mobile_ID\": \"1-800-repoman\",\"email\":\"repoman@alwaysintense.com\",\"First_Name\":\"Otto\",\"Last_Name\":\"Maddox\"}","method":"POST","scheme":"http","host":"127.0.0.1","port":"8080","path":"/api/v1/appuser","protocol":"HTTP/1.1","proto_major":1,"proto_minor":1,"Content Length":129,"Transfer-Encoding":"","Close":false,"RemoteAddr":"127.0.0.1:58689","RequestURI":"/api/v1/appuser","message":"Request received"}
```

>NOTE - the HTTP header key:value pairs and json from the body are represented as escaped JSON within the actual message. If you don't want this data, set these fields to false in the JSON config file (`httpLogOpt.json`)

##### JSON Response Logging

Set `log_json.Response.enable` to true in the [HTTP Log Config File](#Log-Config-File) to enable http response logging as JSON (so long as you have properly "chained" the LogResponse handler/adapter middleware).  The response output will look something like:

```json
{"time":1517970302,"level":"info","request_id":"b9t66vma6806ln8iak8g","response_code":200,"response_header":"{\"Content-Type\":[\"text/plain; charset=utf-8\"],\"Request-Id\":[\"b9t66vma6806ln8iak8g\"]}","response_body":"{\"username\":\"repoMan\",\"mobile_id\":\"1-800-repoman\",\"email\":\"repoman@alwaysintense.com\",\"first_name\":\"Otto\",\"last_name\":\"Maddox\",\"create_user_id\":\"gilcrest\",\"create_date\":\"2018-02-06T21:25:02.538322Z\",\"update_user_id\":\"\",\"update_date\":\"0001-01-01T00:00:00Z\"}\n{\"username\":\"repoMan\",\"mobile_id\":\"1-800-repoman\",\"email\":\"repoman@alwaysintense.com\",\"first_name\":\"Otto\",\"last_name\":\"Maddox\",\"create_user_id\":\"gilcrest\",\"create_date\":\"2018-02-06T21:25:02.538322Z\",\"update_user_id\":\"\",\"update_date\":\"0001-01-01T00:00:00Z\"}\n","message":"Response Sent"}
```

>NOTE - same as request - the HTTP header key:value pairs and json from the body are represented as escaped JSON within the actual message. If you don't want this data, set these fields to false in the JSON config file (`httpLogOpt.json`)

#### Log Style 2: Relational DB Logging via PostgreSQL

Set `log_2DB.enable` to true in the [HTTP Log Config File](#Log-Config-File) to enable Database logging to a PostgreSQL database.  The DDL is provided within the ddl directory (`audit_log.sql`) and consists of one table and one stored function. Once enabled, Request and Response information will be logged as one transaction to the database.  You can optionally choose to log request and response headers using the Options fields within the [HTTP Log Config File](#Log-Config-File).

![Database Log](dbLog.png)

##### Logging Database Table

In total 19 fields are logged as part of the database transaction.

| Column Name   | Datatype    | Description          |
| ------------- | ----------- | -------------------- |
| request_id                | VARCHAR(100)  | Unique Request ID
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

Set `httputil.DumpRequest.enable` to true in the [HTTP Log Config File](#Log-Config-File) to enable logging the request via the [httputil.DumpRequest](https://golang.org/pkg/net/http/httputil/#DumpRequest) method. Nothing special here, really - just providing an easy way to turn this on or off.  Output typically looks like:

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

#### Log Config File

`/input/httpLogOpt.json`

```json
{
    "log_json": {
        "Request": {
            "enable": true,
            "Options": {
                "header": true,
                "body": true
            }
        },
        "Response": {
            "enable": true,
            "Options": {
                "header": true,
                "body": true
            }
        }
    },
    "log_2DB": {
        "enable": false,
        "Request": {
            "header": true,
            "body": true
        },
        "Response": {
            "header": true,
            "body": true
        }
    },
    "httputil": {
        "DumpRequest": {
            "enable": false,
            "body": true
        }
    }
}
```

----

### HTTP JSON Error Responses

For error responses, the api sends a simple structured JSON message in the response body, similar to [Stripe](https://stripe.com/docs/api#errors), [Uber](https://developer.uber.com/docs/riders/guides/errors) and many others, e.g.:

```json
{
    "error": {
        "type": "validation_failed",
        "message": "Username is a required field"
    }
}
```

This is achieved by wrapping the final true app handler (in the below case, CreateUser) inside an ErrHandler type within the dispatch function - `eh.ErrHandler{Env: env, H: handler.CreateUser}`. Note I’m passing in a global environment type as well. I chose this method based on a great article by Matt Silverlock on his blog [here](https://elithrar.github.io/article/http-handler-error-handling-revisited/).

```go
package dispatch

import (
    "github.com/gilcrest/go-API-template/appuser/handler"
    "github.com/gilcrest/go-API-template/env"
    eh "github.com/gilcrest/go-API-template/server/errorHandler"
    "github.com/gilcrest/go-API-template/server/middleware"
    "github.com/gorilla/mux"
)

// Dispatch is a way of organizing routing to handlers (versioning as well)
func Dispatch(env *env.Env, rtr *mux.Router) *mux.Router {

    // initialize new instance of APIAudit
    audit := new(middleware.APIAudit)

    // match only POST requests on /api/appuser/create
    // This is the original (v1) version for the API and the response for this
    // will never change with versioning in order to maintain a stable contract
    rtr.Handle("/appuser", middleware.Adapt(eh.ErrHandler{Env: env, H: handler.CreateUser}, middleware.LogRequest(env, audit), middleware.LogResponse(env, audit))).
        Methods("POST").
        Headers("Content-Type", "application/json")

    // match only POST requests on /api/v1/appuser/create
    rtr.Handle("/v1/appuser", middleware.Adapt(eh.ErrHandler{Env: env, H: handler.CreateUser}, middleware.LogRequest(env, audit), middleware.LogResponse(env, audit))).
        Methods("POST").
        Headers("Content-Type", "application/json")

    return rtr
}
```

I’m using Matt’s article almost word for word for error handling, but made a few tweaks so that I could return a structured JSON response. Check out the server/errorHandler package for the full details.

The package makes error handling pretty nice — given the wrapper logic, you’ll always return a pretty good looking error and setting up errors is pretty easy.

When creating errors within your app, you don’t have to have every error take the HTTPErr form — you can return normal errors lower down in the code and, depending on how you organize your code, you can catch and form the HTTPErr at a very high level so you’re not having to deal with populating a cumbersome struct all throughout your code. The below code snippet illustrates catching any exception thrown from within the Create method and giving the error a certain Error Code and Error Type. Within the Create method itself, errors are set using the Go error "standard".

```go
tx, err := usr.Create(ctx, env)
if err != nil {
    return errorHandler.HTTPErr{
        Code: http.StatusBadRequest,
        Type: "validation_failed",
        Err:  err,
    }
}
```

The SetErr method of the HTTPErr struct allows you to initialize the struct with some default values and add the actual error on the fly. For instance, below as part of my super high level edit checks on my service inputs, the HTTPErr object is initialized at the beginning of the function and then edit checks are performed to allow for brevity in error creation.

```go
func newUser(ctx context.Context, env *env.Env, cur *createUserRequest) (*appuser.User, error) {

    // declare a new instance of appuser.User
    usr := new(appuser.User)

    // initialize an errorHandler with the default Code and Type for
    // service validations (Err is set to nil as it will be set later)
    e := errorHandler.HTTPErr{
        Code: http.StatusBadRequest,
        Type: "validation_error",
        Err:  nil,
    }

    // for each field you can go through whatever validations you wish
    // and use the SetErr method of the HTTPErr struct to add the proper
    // error text
    switch {
    // Username is required
    case cur.Username == "":
        e.SetErr("Username is a required field")
        return nil, e
    // Username cannot be blah...
    case cur.Username == "blah":
        e.SetErr("Username cannot be blah")
        return nil, e
    default:
        usr.Username = cur.Username
    }
    ...
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

- [The http Handler Wrapper Technique in #golang, updated -- by Mat Ryer](https://medium.com/@matryer/the-http-handler-wrapper-technique-in-golang-updated-bc7fbcffa702)
- [Writing middleware in #golang and how Go makes it so much fun. -- by Mat Ryer](https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81)
- [http.Handler and Error Handling in Go -- by Matt Silverlock](https://elithrar.github.io/article/http-handler-error-handling-revisited/)
- [How to correctly use context.Context in Go 1.7 -- by Jack Lindamood](https://medium.com/@cep21/how-to-correctly-use-context-context-in-go-1-7-8f2c0fafdf39)
