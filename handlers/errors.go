package handlers

// handlerError is an error the contains a string with a brief description of
// the error suitable to be seen by a user. It also contains the original
// error.
type handlerError struct {
	user string
	err  error
}

// Error returns the message of the original error.
func (e handlerError) Error() string {
	return e.err.Error()
}
