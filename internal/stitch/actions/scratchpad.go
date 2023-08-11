package actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/slack-go/slack"
)

type Action interface {
	GetType() string
	Render(any) (string, error)
}

type ActionExector interface {
	Execute(Action, any) error
}

type SlackClient interface {
	PostMessage(string, ...slack.MsgOption) (string, string, error)
}

func NewSlackExecutor(client SlackClient) (*SlackExecutor, error) {
	return &SlackExecutor{
		Client: client,
	}, nil
}

type SlackExecutor struct {
	Client SlackClient
}

func (e *SlackExecutor) Execute(action Action, data any) error {
	if t := action.GetType(); t != "slack" {
		return fmt.Errorf("expected action type of 'slack', got '%s'", t)
	}

	slackAction, ok := action.(SlackAction)
	if !ok {
		return fmt.Errorf("error asserting action type")
	}

	rendered, err := slackAction.Render(data)
	if err != nil {
		return err
	}

	var slackOpts slack.MsgOption
	if slackAction.Config.Blocks == true {
		var blocks Blocks
		if err := blocks.UnmarshalJSON([]byte(rendered)); err != nil {
			return fmt.Errorf("unable to process blocks for action '%s': %w", slackAction.Name, err)
		}
		slackOpts = slack.MsgOptionBlocks(blocks.Blocks...)
	} else {
		slackOpts = slack.MsgOptionText(rendered, false)
	}

	for _, channel := range slackAction.Config.Channels {
		_, _, err := e.Client.PostMessage(channel, slackOpts)
		if err != nil {
			return fmt.Errorf("unable to send slack message for action '%s': %w", slackAction.Name, err)
		}
	}

	return nil
}

type Template struct {
	Name    string `yaml:"name"`
	Actions struct {
		Slack []SlackAction `yaml:"slack,flow"`
		Http  []HTTPAction  `yaml:"http,flow"`
	}
}

type SlackAction struct {
	Name      string `yaml:"name"`
	Type      string `yaml:"type"`
	Config    SlackConfig
	templater *template.Template
}

func (a SlackAction) GetName() string {
	return a.Name
}

func (a SlackAction) GetType() string {
	return a.Type
}

func (a SlackAction) Render(data any) (string, error) {
	if a.templater == nil {
		a.initTemplater()
	}

	var buf bytes.Buffer
	if err := a.templater.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error templating values for action '%s': %w", a.Name, err)
	}

	return buf.String(), nil

}

func (a *SlackAction) initTemplater() error {

	t, err := template.New(a.Name).Parse(a.Config.Message)
	if err != nil {
		return err
	}

	a.templater = t
	return nil
}

type SlackConfig struct {
	Channels []string `yaml:"channels"`
	Blocks   bool     `yaml:"blocks"`
	Message  string   `yaml:"message"`
}

type HTTPAction struct {
	Name   string     `yaml:"name"`
	Type   string     `yaml:"type"`
	Config HTTPConfig `yaml:"config"`
}

func (a *HTTPAction) GetName() string {
	return a.Name
}

type HTTPConfig struct {
	Method  string              `yaml:"method"`
	Path    string              `yaml:"path"`
	Headers []map[string]string `yaml:"headers,flow"`
}

// Blocks & associated methods allow easier serializing of `blocks` JSON objects returned by the Slack API or obtained from the Block Kit Builder
type Blocks struct {
	Blocks []slack.Block `json:"blocks"`
}
type blockhint struct {
	Typ string `json:"type"`
}

func (b *Blocks) UnmarshalJSON(data []byte) error {
	var proxy struct {
		Blocks []json.RawMessage `json:"blocks"`
	}
	if err := json.Unmarshal(data, &proxy); err != nil {
		return fmt.Errorf(`failed to unmarshal blocks array: %w`, err)
	}
	for _, rawBlock := range proxy.Blocks {
		var hint blockhint
		if err := json.Unmarshal(rawBlock, &hint); err != nil {
			return fmt.Errorf(`failed to unmarshal next block for hint: %w`, err)
		}
		var block slack.Block
		switch hint.Typ {
		case "actions":
			block = &slack.ActionBlock{}
		case "context":
			block = &slack.ContextBlock{}
		case "divider":
			block = &slack.DividerBlock{}
		case "file":
			block = &slack.FileBlock{}
		case "header":
			block = &slack.HeaderBlock{}
		case "image":
			block = &slack.ImageBlock{}
		case "input":
			block = &slack.InputBlock{}
		case "section":
			block = &slack.SectionBlock{}
		default:
			block = &slack.UnknownBlock{}
		}
		if err := json.Unmarshal(rawBlock, block); err != nil {
			return fmt.Errorf(`failed to unmarshal next block: %w`, err)
		}
		b.Blocks = append(b.Blocks, block)
	}
	return nil
}
