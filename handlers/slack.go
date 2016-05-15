package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strings"
)

// slackRequest is the information gathered from a Slack slash request.
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

// attachmentJSON is JSON data for a Slack message attachment.
type attachmentJSON struct {
	Fallback   string   `json:"fallback"`
	Text       string   `json:"text"`
	Color      string   `json:"color"`
	MarkdownIn []string `json:"mrkdwn_in"`
}

// slackDataJSON is JSON data for a standard slack request or response.
type slackDataJSON struct {
	Markdown     bool             `json:"mrkdwn"`
	ResponseType string           `json:"response_type"`
	Attachments  []attachmentJSON `json:"attachments"`
}

// regularSlackRequest sends a Slack request to the given URL with standard
// markdown-formatted data based on the given message. The message will only be
// seen by the user.
func regularSlackRequest(client *http.Client, url string, message string) error {
	reqJSON := slackDataJSON{
		Markdown:     true,
		ResponseType: "ephemeral",
		Attachments: []attachmentJSON{
			{
				Fallback:   message,
				Text:       message,
				Color:      "error",
				MarkdownIn: []string{"text"}}}}

	// Turn the JSON information into formatted data
	postData, err := json.Marshal(reqJSON)
	if err != nil {
		return err
	}

	_, err = client.Post(url, "application/json", bytes.NewBuffer(postData))
	return err
}

// regularSlackTemplRequest sends a Slack request to the given URL with
// standard markdown-formatted data using the given template and template info.
// The message will only be seen by the user.
func regularSlackTemplRequest(client *http.Client, url string, templ *template.Template, templInfo interface{}) error {
	templData := &bytes.Buffer{}
	err := templ.Execute(templData, templInfo)
	if err != nil {
		return err
	}

	reqJSON := slackDataJSON{
		Markdown:     true,
		ResponseType: "ephemeral",
		Attachments: []attachmentJSON{
			{
				Fallback:   string(templData.Bytes()),
				Text:       string(templData.Bytes()),
				Color:      "",
				MarkdownIn: []string{"text"}}}}

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
		Markdown:     true,
		ResponseType: "in_channel",
		Attachments: []attachmentJSON{
			{
				Fallback:   message,
				Text:       message,
				Color:      "danger",
				MarkdownIn: []string{"text"}}}}

	// Turn the JSON information into formatted data
	postData, err := json.Marshal(reqJSON)
	if err != nil {
		return err
	}

	_, err = client.Post(url, "application/json", bytes.NewBuffer(postData))
	return err
}

// publicSlackTemplRequest sends a Slack request to the given URL with
// standard markdown-formatted data using the given template and template info.
// The message will be seen by the whole channel
func publicSlackTemplRequest(client *http.Client, url string, templ *template.Template, templInfo interface{}) error {
	templData := &bytes.Buffer{}
	err := templ.Execute(templData, templInfo)
	if err != nil {
		return err
	}

	reqJSON := slackDataJSON{
		Markdown:     true,
		ResponseType: "in_channel",
		Attachments: []attachmentJSON{
			{
				Fallback:   string(templData.Bytes()),
				Text:       string(templData.Bytes()),
				Color:      "",
				MarkdownIn: []string{"text"}}}}

	// Turn the JSON information into formatted data
	postData, err := json.Marshal(reqJSON)
	if err != nil {
		return err
	}

	_, err = client.Post(url, "application/json", bytes.NewBuffer(postData))
	return err
}
