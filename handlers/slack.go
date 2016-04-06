package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
)

type slackRequest struct {
	token         string
	teamID        string
	teamDomain    string
	channelID     string
	channelName   string
	userID        string
	username      string
	commandName   string
	commandParams []string
	text          string
	responseURL   string
}

// newSlackRequest creates a new slackRequest object fromthe given HTTP
// request. If the request does not contain every expected parameter, then an
// error will be returned.
func newSlackRequest(r *http.Request) (slackRequest, error) {
	r.ParseForm()
	// List of expected request parameter names
	paramNames := []string{"token", "team_id", "team_domain", "channel_id",
		"channel_name", "user_id", "user_name", "command", "text", "response_url"}
	params := make(map[string]string)

	// Check if all the expected request parameters are available
	for _, val := range paramNames {
		_, ok := r.Form[val]
		if !ok {
			return slackRequest{}, errors.New("missing parameter '" + val + "' in request body")
		}
		params[val] = r.Form[val][0]
	}

	// Preparse the text as space-seperated values
	var commandName string
	var commandParams []string
	command := strings.Split(strings.Trim(params["text"], " "), " ")
	if len(command) != 0 {
		commandName = strings.ToUpper(command[0])
		if len(command) > 1 {
			commandParams = command[1:]
		}
	}

	return slackRequest{
		token:         params["token"],
		teamID:        params["team_id"],
		teamDomain:    params["team_domain"],
		channelID:     params["channel_id"],
		channelName:   params["channel_name"],
		userID:        params["user_id"],
		username:      params["user_name"],
		commandName:   commandName,
		commandParams: commandParams,
		text:          params["text"],
		responseURL:   params["response_url"]}, nil
}

// slackDataJSON is JSON data for a standard slack request or response.
type slackDataJSON struct {
	Text         string `json:"text"`
	Markdown     bool   `json:"mrkdwn"`
	ResponseType string `json:"response_type"`
}

// regularSlackResponse responds to a request with a standard markdown-formatted
// text response based on the given message. The message will only be seen by the
// user.
func regularSlackResponse(w http.ResponseWriter, r *http.Request, message string) {
	// Create the JSON response
	respJSON := slackDataJSON{
		Text:         message,
		Markdown:     true,
		ResponseType: "ephemeral"}

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

// regularSlackRequest sends a Slack request to the given URL with standard
// markdown-formatted data based on the given message. The message will only be
// seen by the user.
func regularSlackRequest(client *http.Client, url string, message string) error {
	reqJSON := slackDataJSON{
		Text:         message,
		Markdown:     true,
		ResponseType: "ephemeral"}

	// Turn the JSON information into formatted data
	postData, err := json.Marshal(reqJSON)
	if err != nil {
		return err
	}

	_, err = client.Post(url, "application/json", bytes.NewBuffer(postData))
	return err
}

// publicSlackRequest sends a Slack request to the given URL with standard
// markdown-formatted data based on the given message. The message will be seen
// by the whole channel.
func publicSlackRequest(client *http.Client, url string, message string) error {
	reqJSON := slackDataJSON{
		Text:         message,
		Markdown:     true,
		ResponseType: "in_channel"}

	// Turn the JSON information into formatted data
	postData, err := json.Marshal(reqJSON)
	if err != nil {
		return err
	}

	_, err = client.Post(url, "application/json", bytes.NewBuffer(postData))
	return err
}
