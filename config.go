package snoreslacks

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

// getTokenFromConfig reads the expected token value from the database
// and returns it, or an error if it couldn't be fetched.
func getTokenFromConfig(ctx context.Context) (string, error) {
	key := datastore.NewKey(ctx, "config", "token", 0, nil)
	var token string

	err := datastore.Get(ctx, key, &token)
	if err != nil {
		return "", err
	}

	return token, nil
}
