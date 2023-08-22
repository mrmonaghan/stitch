package stitch

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// represents a Slack action
type SlackAction struct {
	Name   string      `yaml:"name"`
	Config SlackConfig `yaml:"config"`
}

// render the SlackAction's Message
func (a *SlackAction) Render(data any) error {

	r, err := render(a.Name, a.Config.Message, data)
	if err != nil {
		return err
	}

	a.Config.Message = r
	return nil
}

// Returns a yaml representation of SlackAction
func (a *SlackAction) String() string {
	b, _ := yaml.Marshal(a)
	return string(b)
}

// describes configuration options for SlackAction. Processing entities can do with these as they will.
type SlackConfig struct {
	Channels []string `yaml:"channels"`
	Blocks   bool     `yaml:"blocks"`
	Message  string   `yaml:"message"`
}

// represents an HTTPAction
type HTTPAction struct {
	Name   string     `yaml:"name"`
	Type   string     `yaml:"type"`
	Config HTTPConfig `yaml:"config"`
}

// render the HTTPAction's Config.URL, Config.Headers, and Config.Body
func (a *HTTPAction) Render(data any) error {

	if err := a.renderHeaders(data); err != nil {
		return err
	}

	if err := a.renderUrl(data); err != nil {
		return err
	}

	r, err := render(a.Name, a.Config.Body, data)
	if err != nil {
		return err
	}

	a.Config.Body = r
	return nil
}

// render user-supplied headers
func (a *HTTPAction) renderHeaders(data any) error {
	yamlConfig, err := yaml.Marshal(a.Config.Headers)
	if err != nil {
		return err
	}
	r, err := render(a.Name, string(yamlConfig), data)
	if err != nil {
		return err
	}

	var conf HTTPConfig

	if err := yaml.Unmarshal([]byte(r), &conf.Headers); err != nil {
		return err
	}

	a.Config.Headers = conf.Headers

	return nil
}

// render the user-supplied URL
func (a *HTTPAction) renderUrl(data any) error {
	yamlConfig, err := yaml.Marshal(a.Config.URL)
	if err != nil {
		return err
	}
	r, err := render(a.Name, string(yamlConfig), data)
	if err != nil {
		return err
	}

	var conf HTTPConfig

	if err := yaml.Unmarshal([]byte(r), &conf.URL); err != nil {
		return err
	}

	a.Config.URL = conf.URL

	return nil
}

// returns an *http.Request based on HTTPAction.Config
func (a *HTTPAction) Request(data any) (*http.Request, error) {

	err := a.Render(data)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest(a.Config.Method, a.Config.URL, strings.NewReader(a.Config.Body))
	if err != nil {
		return nil, err
	}

	for _, header := range a.Config.Headers {
		r.Header.Add(header.Name, header.Value)
	}

	return r, nil
}

// validates custom statusCode logic defined in HTTPConfig
func (a *HTTPAction) CheckStatusCode(code int) error {

	for _, errorCode := range a.Config.StatusCodes.Failure {
		if code == errorCode {
			return fmt.Errorf("status code '%d' indicates failure based on action configuration", code)
		}
	}

	for _, successCode := range a.Config.StatusCodes.Success {
		if code == successCode {
			return nil
		}
	}

	return nil
}

// returns a YAML representation of HTTPAction
func (a *HTTPAction) String() string {
	b, _ := yaml.Marshal(a)
	return string(b)
}

// describes configuration options for HTTPAction
type HTTPConfig struct {
	Method      string   `yaml:"method"`
	URL         string   `yaml:"url"`
	Headers     []Header `yaml:"headers,omitempty"`
	Body        string   `yaml:"body,omitempty"`
	StatusCodes struct {
		Success []int `yaml:"success,omitempty"`
		Failure []int `yaml:"failure,omitempty"`
	} `yaml:"statusCodes,omitempty"`
}

type Header struct {
	Name  string
	Value string
}

// helper function for rendering template data
func render(name string, tmpl string, data any) (string, error) {
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
