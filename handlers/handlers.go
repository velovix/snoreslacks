package handlers

import (
	"encoding/gob"
	"fmt"
	"net/http"

	"github.com/velovix/snoreslacks/ctxman"
	"github.com/velovix/snoreslacks/database"
	"github.com/velovix/snoreslacks/logging"
	"github.com/velovix/snoreslacks/messaging"
	"github.com/velovix/snoreslacks/pokeapi"
	"github.com/velovix/snoreslacks/tasking"
	"golang.org/x/net/context"
)

// Tasker describes a task that can be run using a Runner.
type Tasker interface {
	runTask(ctx context.Context, s Services) error
}

// Preprocessor describes a task that needs to do some preprocessing before
// its main work inside the transaction. This is usually for adding request
// scoped variables to the request context that are specific to this task.
type Preprocessor interface {
	Tasker
	// preprocess does preprocessing work specific to the task. The second
	// return value is false if the task itself should not run for whatever
	// reason.
	preprocess(ctx context.Context, s Services) (context.Context, bool, error)
}

// Runner is a handler object that can run tasks.
type Runner struct {
	Servs Services
	Task  Tasker
}

// ServeHTTP prepares some request-scoped information and runs the task,
// handling any errors.
func (r Runner) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Create the request context
	ctx, err := r.Servs.CtxCreator.Create(req)
	if err != nil {
		http.Error(w, "could not create the request context", 500)
		r.Servs.Log.Errorf(ctx, "while creating the request context: %s", err)
		return
	}

	// Create the HTTP client
	client, err := r.Servs.ClientCreator.Create(ctx)
	if err != nil {
		http.Error(w, "could not create an HTTP client", 500)
		r.Servs.Log.Errorf(ctx, "while creating an HTTP client: %s", err)
		return
	}
	ctx = context.WithValue(ctx, "client", client)

	// Decode the Slack request
	slackReq, err := decodeSlackReq(req)
	if err != nil {
		http.Error(w, "could not decode Slack request", 500)
		r.Servs.Log.Errorf(ctx, "while decoding the Slack request: %s", err)
		return
	}
	ctx = context.WithValue(ctx, "slack request", slackReq)

	// Load the requesting trainer's data, if one exists
	t, err := loadBasicTrainerData(ctx, r.Servs.DB, slackReq.UserID)
	if err == nil {
		// Add the trainer to the context since we have it
		ctx = context.WithValue(ctx, "requesting trainer", t)
	} else if err != nil && !database.IsNoResults(err) {
		// There is an error and it isn't because there isn't a trainer
		// available
		http.Error(w, "could not load basic trainer data", 500)
		r.Servs.Log.Errorf(ctx, "while loading basic trainer data: %s", err)
		return
	}

	// Load the battle data, if it exists and if the trainer exists
	if t, ok := ctx.Value("requesting trainer").(*basicTrainerData); ok {
		bd, err := loadBattleData(ctx, r.Servs.DB, t)
		if err == nil || database.IsNoResults(err) {
			// Add the battle data to the context even if it's incomplete
			ctx = context.WithValue(ctx, "battle data", bd)
		} else if err != nil {
			// There is an error and it is not because of a lack of data
			http.Error(w, "could not load battle data", 500)
			r.Servs.Log.Errorf(ctx, "while loading battle data: %s", err)
			return
		}
	}

	// Do any required preprocessing if the task is a preprocessor
	shouldRunTask := true
	if proc, ok := r.Task.(Preprocessor); ok {
		ctx, shouldRunTask, err = proc.preprocess(ctx, r.Servs)
	}

	// Run the task, so long as the preprocessing was successful and it said
	// the task should be run
	if err == nil && shouldRunTask {
		err = r.Servs.DB.Transaction(ctx, func(ctx context.Context) error {
			return r.Task.runTask(ctx, r.Servs)
		})
	}
	if err != nil {
		// An error has occurred in either processing or preprocessing
		switch err := err.(type) {
		case handlerError:
			// We have special processing for handlerErrors

			// Notify the user of the error
			messaging.Send(client, t.lastContactURL, messaging.Message{
				Text: err.user,
				Type: messaging.Error})
			// Log the fact that the error happened
			r.Servs.Log.Errorf(ctx, "%+v", err.err)
			// Fail the request
			http.Error(w, fmt.Sprintf("%+v", err.err), 500)
		default:
			// A default error has slipped through, so we'll handle it in a
			// generic way

			// Notify the user of the error generically
			messaging.Send(client, t.lastContactURL, messaging.Message{
				Text: "an error has occurred whild processing your request",
				Type: messaging.Error})
			// Log the fact that the error happened
			r.Servs.Log.Errorf(ctx, "%+v", err)
			// Fail the request
			http.Error(w, fmt.Sprintf("%+v", err), 500)
		}
	}
}

// Services contains references to the services needed by every handler. This
// object should be an embedded struct inside of each handler. Helper methods
// that are useful to multiple handlers and operate on these services should be
// a method of this type.
type Services struct {
	CtxCreator    ctxman.Creator
	Log           logging.Logger
	DB            database.Database
	ClientCreator messaging.ClientCreator
	Fetcher       pokeapi.Fetcher
	WorkQueue     tasking.Queue
}

// decodeSlackReq decodes a Slack request from the given HTTP request.
func decodeSlackReq(r *http.Request) (messaging.SlackRequest, error) {
	decoder := gob.NewDecoder(r.Body)
	defer r.Body.Close()

	var slackReq messaging.SlackRequest
	err := decoder.Decode(&slackReq)
	if err != nil {
		return messaging.SlackRequest{}, err
	}

	return slackReq, nil
}
