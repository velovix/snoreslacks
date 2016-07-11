package handlers

import "github.com/velovix/snoreslacks/messaging"

// sendInvalidCommand complains to the user that their command was not
// formatted correctly.
func sendInvalidCommand(client messaging.Client, url string) error {
	err := messaging.SendTempl(client, url, messaging.TemplMessage{
		Templ:     invalidCommandTemplate,
		TemplInfo: nil})
	if err != nil {
		return handlerError{user: "could not populate invalid command template", err: err}
	}
	return nil
}
