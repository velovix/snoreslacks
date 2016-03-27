package app

import (
	"runtime/debug"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type logger interface {
	infof(ctx context.Context, format string, data ...interface{})
	debugf(ctx context.Context, format string, data ...interface{})
	warningf(ctx context.Context, format string, data ...interface{})
	errorf(ctx context.Context, format string, data ...interface{})
	criticalf(ctx context.Context, format string, data ...interface{})
}

type appengineLogger struct{}

func (l appengineLogger) infof(ctx context.Context, format string, data ...interface{}) {
	log.Infof(ctx, format, data...)
}

func (l appengineLogger) debugf(ctx context.Context, format string, data ...interface{}) {
	log.Debugf(ctx, format, data...)
}

func (l appengineLogger) warningf(ctx context.Context, format string, data ...interface{}) {
	log.Warningf(ctx, format, data...)
}

func (l appengineLogger) errorf(ctx context.Context, format string, data ...interface{}) {
	log.Errorf(ctx, string(debug.Stack())) // Print a stack trace
	log.Errorf(ctx, format, data...)
}

func (l appengineLogger) criticalf(ctx context.Context, format string, data ...interface{}) {
	log.Criticalf(ctx, format, data...)
}
