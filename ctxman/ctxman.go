// Package ctxman provides interfaces for context management. Applications
// should use this library directly and not its implementations.
package ctxman

import (
	"net/http"

	"github.com/pkg/errors"

	"golang.org/x/net/context"
)

// Creator describes an object that can create a context from an HTTP request.
type Creator interface {
	// Create creates a context from the given HTTP request.
	Create(r *http.Request) (context.Context, error)
}

var implementations map[string]Creator

func init() {
	implementations = make(map[string]Creator)
}

// Register registers an implementation of the Creator interface under the
// given name.
func Register(name string, creator Creator) {
	implementations[name] = creator
}

// Get returns the implementation of Creator with the given name, or an error
// if no such implementation exists.
func Get(name string) (Creator, error) {
	if creator, ok := implementations[name]; ok {
		return creator, nil
	}

	return nil, errors.New("no ctxman.Creator implementation called '" + name + "' found")
}
