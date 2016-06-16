// Package GAELogging includes an implementation of the Logger interface for
// Google App Engine. This package should not be used directly. Instead, it
// should be imported once in your app for its side-effects.
//
// 	import _ "github.com/velovix/snoreslacks/logging/gae"
// 	...
//	log, err := logger.Get("gae")
package GAELogging

import (
	"github.com/velovix/snoreslacks/logging"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

// GAELogger implements the Logger interface for Google App Engine.
type GAELogger struct{}

// Infof logs to Google App Engine at the INFO log level.
func (l GAELogger) Infof(ctx context.Context, format string, data ...interface{}) {
	log.Infof(ctx, format, data...)
}

// Debugf logs to Google App Engine at the DEBUG log level.
func (l GAELogger) Debugf(ctx context.Context, format string, data ...interface{}) {
	log.Debugf(ctx, format, data...)
}

// Warningf logs to Google App Engine at the WARN log level.
func (l GAELogger) Warningf(ctx context.Context, format string, data ...interface{}) {
	log.Warningf(ctx, format, data...)
}

// Errorf logs to Google App Engine at the ERROR log level. It also includes a
// stack trace.
func (l GAELogger) Errorf(ctx context.Context, format string, data ...interface{}) {
	log.Errorf(ctx, format, data...)
}

// Criticalf logs to Google App Engine at the ERROR log level. It also includes
// a stack trace.
func (l GAELogger) Criticalf(ctx context.Context, format string, data ...interface{}) {
	log.Criticalf(ctx, format, data...)
}

func init() {
	// Add GAELogger to the list of available loggers
	logging.Register("gae", GAELogger{})
}
