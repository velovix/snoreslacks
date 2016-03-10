package snoreslacks

import (
	"encoding/json"
	"log"
	"net/http"
)

// regularSlackResponseJSON is JSON data for a standard slack response.
type regularSlackResponseJSON struct {
	Text     string `json:"text"`
	Markdown bool   `json:"mrkdwn"`
}

// regularSlackResponse responds to a request with a standard markdown-formatted
// text response based on the given message.
func regularSlackResponse(w http.ResponseWriter, r *http.Request, message string) {
	// Create the JSON response
	respJSON := regularSlackResponseJSON{
		Text:     message,
		Markdown: true}

	// Turn the JSON information into formatted data
	respData, err := json.Marshal(respJSON)
	if err != nil {
		// The response could not be marshalled. This should not happen
		http.Error(w, "failed to construct response", 500)
		log.Println("while creating a regular slack response:", err)
		return
	}

	w.Header()["Content-Type"] = []string{"application/json"}
	w.Write(respData)
}
