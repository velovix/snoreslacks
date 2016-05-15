package gaeapp

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"

	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/handlers"
	"github.com/velovix/snoreslacks/logging"
	"github.com/velovix/snoreslacks/pokeapi"

	// Get the GAE implementations
	_ "github.com/velovix/snoreslacks/database/gae"
	_ "github.com/velovix/snoreslacks/logging/gae"
	_ "github.com/velovix/snoreslacks/pokeapi/gae"
)

func init() {
	loadConfig()

	// Inject all Google App Engine dependencies
	db, err := database.Get("gae")
	if err != nil {
		panic(err)
	}
	log, err := logging.Get("gae")
	if err != nil {
		panic(err)
	}
	fetcher, err := pokeapi.Get("gae")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		client := urlfetch.Client(ctx)
		handlers.MainHandler(ctx, w, r, db, log, client, fetcher, config.Token)
	})
}
