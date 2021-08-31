// Package logger has helpers to setup a zerolog.Logger
//   https://github.com/rs/zerolog
package logger

import (
	"io"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// NewLogger is a convenience function to initialize a zerolog.Logger
func NewLogger(w io.Writer, lvl zerolog.Level, withTimestamp bool) zerolog.Logger {
	// logger is initialized with the writer and level passed in.
	// All logs will be written at the given level (unless raised
	// using zerolog.SetGlobalLevel)
	lgr := zerolog.New(w).Level(lvl)
	if withTimestamp {
		lgr = lgr.With().Timestamp().Logger()
	}

	// Add Severity Hook. Zerolog by default outputs structured logs
	// with "level":"error" as its leveling. Google Cloud as an
	// example expects "severity","ERROR" for its leveling. This
	// hook will add severity to each message
	lgr = lgr.Hook(GCPSeverityHook{})

	return lgr
}

// The GCPSeverityHook struct satisfies the zerolog.Hook interface
// as it has the Run method defined with the appropriate parameters
type GCPSeverityHook struct{}

// Run method satisfies zerolog.Hook interface and adds a severity
// level to all logs, given zerolog.Level passed in. Zerolog levels
// are mapped to log levels recognized by GCP
func (h GCPSeverityHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {

	const lvlKey string = "severity"

	switch level {
	case zerolog.PanicLevel:
		e.Str(lvlKey, "EMERGENCY")
	case zerolog.FatalLevel:
		e.Str(lvlKey, "EMERGENCY")
	case zerolog.ErrorLevel:
		e.Str(lvlKey, "ERROR")
	case zerolog.WarnLevel:
		e.Str(lvlKey, "WARNING")
	case zerolog.InfoLevel:
		e.Str(lvlKey, "INFO")
	case zerolog.DebugLevel:
		e.Str(lvlKey, "DEBUG")
	case zerolog.TraceLevel:
		e.Str(lvlKey, "DEBUG")
	default:
		e.Str(lvlKey, "DEFAULT")
	}
}

// WriteErrorStackGlobal is a convenience wrapper to set the zerolog
// Global variable ErrorStackMarshaler to write Error stacks for logs
func WriteErrorStackGlobal(writeStack bool) {
	if !writeStack {
		zerolog.ErrorStackMarshaler = nil
		return
	}
	// set ErrorStackMarshaler to pkgerrors.MarshalStack
	// to enable error stack traces
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
}
