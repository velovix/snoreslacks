// Package gaemessaging provides an implementation of the interfaces in the
// messaging package for Google App engine. This package should not be used
// directly. Instead, it should be imported once in your project for its
// side-effects.
//
//	import _ "github.com/velovix/snoreslacks/messaging/gae"
//	...
//	clientCreator, err := messaging.GetClientCreator("gae")
package gaemessaging

import (
	"github.com/velovix/snoreslacks/messaging"
	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
)

type GAEClientCreator struct {
}

func (cc GAEClientCreator) Create(ctx context.Context) (messaging.Client, error) {
	return urlfetch.Client(ctx), nil
}

func init() {
	messaging.RegisterClientCreator("gae", GAEClientCreator{})
}
