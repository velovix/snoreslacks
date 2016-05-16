package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strings"
)

type msgType int

const (
	infoMsgType msgType = iota
	errorMsgType
	goodMsgType
	importantMsgType
)

// colorFromMsgType converts a message type to an appropriate color value that
// Slack understands.
func colorFromMsgType(t msgType) string {
	switch t {
	case infoMsgType:
		return ""
	case errorMsgType:
		return "danger"
	case goodMsgType:
		return "good"
	case importantMsgType:
		return "warning"
	default:
		panic("unsupported message type")
	}
}

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
	ThumbURL   string   `json:"thumb_url"`
}

// slackDataJSON is JSON data for a standard slack request or response.
type slackDataJSON struct {
	Markdown     bool             `json:"mrkdwn"`
	ResponseType string           `json:"response_type"`
	Attachments  []attachmentJSON `json:"attachments"`
}

// message contains information on a message to be sent to Slack.
type message struct {
	public bool
	t      msgType
	image  string
	text   string
}

// templMessage contains information on a message to be sent to Slack using
// a template.
type templMessage struct {
	public    bool
	t         msgType
	image     string
	templ     *template.Template
	templInfo interface{}
}

// sendMessage sends a Slack request to the given URL based on the given
// message.
func sendMessage(client *http.Client, url string, msg message) error {
	// Find the appropriate publicity value
	var publicity string
	if msg.public {
		publicity = "in_channel"
	} else {
		publicity = "ephemeral"
	}

	// Construct the response
	reqJSON := slackDataJSON{
		Markdown:     true,
		ResponseType: publicity,
		Attachments: []attachmentJSON{
			{
				ThumbURL:   msg.image,
				Fallback:   msg.text,
				Text:       msg.text,
				Color:      colorFromMsgType(msg.t),
				MarkdownIn: []string{"text"}}}}

	// Turn the JSON information into formatted data
	postData, err := json.Marshal(reqJSON)
	if err != nil {
		return err
	}

	_, err = client.Post(url, "application/json", bytes.NewBuffer(postData))
	return err
}

// sendTemplMessage sends a Slack message based on the given template message.
func sendTemplMessage(client *http.Client, url string, msg templMessage) error {
	templData := &bytes.Buffer{}
	err := msg.templ.Execute(templData, msg.templInfo)
	if err != nil {
		return err
	}

	return sendMessage(client, url, message{
		public: msg.public,
		image:  msg.image,
		t:      msg.t,
		text:   string(templData.Bytes())})
}
