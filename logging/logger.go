// Package logging contains an interface for logging. This package should be
// used directly instead of the implementations. A package that provides this
// functionality should register themselves to this package and not have to be
// used directly.
package logging

import (
	"github.com/pkg/errors"

	"golang.org/x/net/context"
)

// Logger describes an object that provides a basic logging service.
type Logger interface {
	Infof(ctx context.Context, format string, data ...interface{})
	Debugf(ctx context.Context, format string, data ...interface{})
	Warningf(ctx context.Context, format string, data ...interface{})
	Errorf(ctx context.Context, format string, data ...interface{})
	Criticalf(ctx context.Context, format string, data ...interface{})
}

var registered map[string]Logger

func init() {
	registered = make(map[string]Logger)
}

// Register registers an implementation of the Logger interface.
func Register(name string, logger Logger) {
	registered[name] = logger
}

// Get returns an implementation of a Logger with the given name or an error
// if no implementation with that name exists.
func Get(name string) (Logger, error) {
	if l, ok := registered[name]; ok {
		return l, nil
	}

	return nil, errors.New("no logger with the name '" + name + "' found")
}
