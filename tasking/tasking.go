package tasking

import (
	"github.com/pkg/errors"

	"golang.org/x/net/context"
)

// Queue describes an object that acts as a work queue. Work is executed by
// sending HTTP requests to the appropriate hook that knows how to get the work
// done.
type Queue interface {
	// Add adds a task to the work queue. When it is time for the task to
	// execute, an HTTP POST request will be sent to the given URL containing
	// the given data. The content type of the request will be
	// application/octet-stream.
	Add(ctx context.Context, url string, data []byte)
}

var implementations map[string]Queue

func init() {
	implementations = make(map[string]Queue)
}

// Register registers an implementation of the queue interface.
func Register(name string, queue Queue) {
	implementations[name] = queue
}

// Get returns an implementation of the queue interface with the given name or
// an error if no implementation with that name exists.
func Get(name string) (Queue, error) {
	if q, ok := implementations[name]; ok {
		return q, nil
	}

	return nil, errors.New("no Queue implementation with the name '" + name + "' found")
}
