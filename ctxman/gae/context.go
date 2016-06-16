// Package gaectxman provides implementations of the context management APIs
// for the Google App Engine platform. This package should not be used
// directly. Instead, it should be imported once in your project for its
// side-effects.
//
//	import _ "github.com/velovix/snoreslacks/ctxman/gae"
//  ...
//  ctxCreator, err := database.Get("gae")
package gaectxman

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/velovix/snoreslacks/ctxman"
	"google.golang.org/appengine"
)

// GAECreator is the GAE implementation of the ctxman.Creator interface.
type GAECreator struct{}

// Create creates a new context from the given request.
func (c GAECreator) Create(r *http.Request) (context.Context, error) {
	return appengine.NewContext(r), nil
}

func init() {
	ctxman.Register("gae", GAECreator{})
}
