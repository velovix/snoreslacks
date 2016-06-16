package gaetasking

import (
	"net/http"
	"time"

	"github.com/velovix/snoreslacks/tasking"

	"golang.org/x/net/context"
	"google.golang.org/appengine/taskqueue"
)

// GAEQueue provides a tasking.Queue implementation for Google App Engine,
// using the taskqueue library.
type GAEQueue struct {
}

// Add adds a task to the work queue. When it is time to execute the task, an
// HTTP POST request will be sent to the given URL along with the given data.
func (q GAEQueue) Add(ctx context.Context, url string, data []byte) {
	task := &taskqueue.Task{
		Path:    url,
		Payload: data,
		RetryOptions: &taskqueue.RetryOptions{
			RetryLimit: 0,
			AgeLimit:   time.Nanosecond},
		Header: http.Header(map[string][]string{
			"Content-Type": {"application/octet-stream"}}),
		Method: "POST"}
	taskqueue.Add(ctx, task, "")
}

func init() {
	tasking.Register("gae", GAEQueue{})
}
