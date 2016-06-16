package messaging

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"golang.org/x/net/context"
)

// Client describes an object that can send HTTP requests. Basically, it's an
// interface version of http.Client for testing and better flexability of
// design.
type Client interface {
	Get(url string) (*http.Response, error)
	Post(url string, bodyType string, body io.Reader) (*http.Response, error)
}

// ClientCreator describes an object that can construct a Client using the
// given context.
type ClientCreator interface {
	Create(ctx context.Context) (Client, error)
}

var clientCreatorImpls map[string]ClientCreator

func init() {
	clientCreatorImpls = make(map[string]ClientCreator)
}

// RegisterClientCreator registers a new implementation of a ClientCreator under
// the given name.
func RegisterClientCreator(name string, creator ClientCreator) {
	clientCreatorImpls[name] = creator
}

// GetClientCreator returns an implementation of a ClientCreator with the
// given name, or an error if no such implementation exists.
func GetClientCreator(name string) (ClientCreator, error) {
	creator, ok := clientCreatorImpls[name]
	if !ok {
		return nil, errors.New("no client creator implementation with the name '" + name + "' found")
	}
	return creator, nil
}

// MsgType describes the type of the message, which is used for smarter
// formatting.
type MsgType int

const (
	Info MsgType = iota
	Error
	Good
	Important
)

// colorFromMsgType converts a message type to an appropriate color value that
// Slack understands.
func colorFromMsgType(t MsgType) string {
	switch t {
	case Info:
		return ""
	case Error:
		return "danger"
	case Good:
		return "good"
	case Important:
		return "warning"
	default:
		panic("unsupported message type")
	}
}

// SlackRequest is the information gathered from a Slack slash request.
type SlackRequest struct {
	Token         string
	TeamID        string
	TeamDomain    string
	ChannelID     string
	ChannelName   string
	UserID        string
	Username      string
	SlashCommand  string
	CommandName   string
	CommandParams []string
	Text          string
	ResponseURL   string
}

// NewSlackRequest creates a new SlackRequest object fromthe given HTTP
// request. If the request does not contain every expected parameter, then an
// error will be returned.
func NewSlackRequest(r *http.Request) (SlackRequest, error) {
	r.ParseForm()
	// List of expected request parameter names
	paramNames := []string{"token", "team_id", "team_domain", "channel_id",
		"channel_name", "user_id", "user_name", "command", "text", "response_url"}
	params := make(map[string]string)

	// Check if all the expected request parameters are available
	for _, val := range paramNames {
		_, ok := r.Form[val]
		if !ok {
			return SlackRequest{}, errors.New("missing parameter '" + val + "' in request body")
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

	return SlackRequest{
		Token:         params["token"],
		TeamID:        params["team_id"],
		TeamDomain:    params["team_domain"],
		ChannelID:     params["channel_id"],
		ChannelName:   params["channel_name"],
		UserID:        params["user_id"],
		Username:      params["user_name"],
		SlashCommand:  params["command"],
		CommandName:   commandName,
		CommandParams: commandParams,
		Text:          params["text"],
		ResponseURL:   params["response_url"]}, nil
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

// Message contains information on a message to be sent to Slack.
type Message struct {
	Public bool
	Type   MsgType
	Image  string
	Text   string
}

// TemplMessage contains information on a message to be sent to Slack using
// a template.
type TemplMessage struct {
	Public    bool
	Type      MsgType
	Image     string
	Templ     *template.Template
	TemplInfo interface{}
}

// Send sends a Slack request to the given URL based on the given
// message.
func Send(client Client, url string, msg Message) error {
	// Find the appropriate publicity value
	var publicity string
	if msg.Public {
		publicity = "in_channel"
	} else {
		publicity = "ephemeral"
	}

	// Construct the request
	reqJSON := slackDataJSON{
		Markdown:     true,
		ResponseType: publicity,
		Attachments: []attachmentJSON{
			{
				ThumbURL:   msg.Image,
				Fallback:   msg.Text,
				Text:       msg.Text,
				Color:      colorFromMsgType(msg.Type),
				MarkdownIn: []string{"text"}}}}

	// Turn the JSON information into formatted data
	postData, err := json.Marshal(reqJSON)
	if err != nil {
		return errors.Wrap(err, "constructing Slack request")
	}

	_, err = client.Post(url, "application/json", bytes.NewBuffer(postData))
	if err != nil {
		return errors.Wrap(err, "sending Slack request")
	}

	return nil
}

// SendTempl sends a Slack message based on the given template message.
func SendTempl(client Client, url string, msg TemplMessage) error {
	templData := &bytes.Buffer{}
	err := msg.Templ.Execute(templData, msg.TemplInfo)
	if err != nil {
		return errors.Wrap(err, "executing Slack template")
	}

	return Send(client, url, Message{
		Public: msg.Public,
		Image:  msg.Image,
		Type:   msg.Type,
		Text:   string(templData.Bytes())})
}
