package stitch

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"gopkg.in/yaml.v3"
)

type Template struct {
	Name    string `yaml:"name"`
	Actions struct {
		Slack []SlackAction `yaml:"slack,flow"`
		HTTP  []HTTPAction  `yaml:"http,flow"`
	}
	Executors struct {
		Slack SlackExecutor
		HTTP  HTTPExecutor
	}
}

func (t *Template) Execute(data any) error {
	for _, slackAction := range t.Actions.Slack {
		rendered, err := t.renderAction(slackAction.Config.Message, data)
		if err != nil {
			return err
		}

		if err := t.Executors.Slack(rendered, slackAction.Config); err != nil {
			return err
		}
	}

	for _, httpAction := range t.Actions.HTTP {
		rendered, err := t.renderAction(httpAction.Config.Body, data)
		if err != nil {
			return err
		}

		if err := t.Executors.HTTP(rendered, httpAction.Config); err != nil {
			return err
		}
	}

	return nil
}

func (t *Template) RenderAction(actionName string, data any) (string, error) {
	for _, slackAction := range t.Actions.Slack {
		if actionName == slackAction.Name {
			return t.renderAction(slackAction.Config.Message, data)
		}
	}

	for _, httpAction := range t.Actions.HTTP {
		if actionName == httpAction.Name {
			return t.renderAction(httpAction.Config.Body, data)
		}
	}

	return "", fmt.Errorf("no action '%s' found", actionName)
}

func (t *Template) RenderAll(data any) (map[string]string, error) {
	m := make(map[string]string)

	for _, slackAction := range t.Actions.Slack {
		rendered, err := t.renderAction(slackAction.Config.Message, data)
		if err != nil {
			return m, err
		}

		m[slackAction.Name] = rendered
	}

	for _, httpAction := range t.Actions.HTTP {
		rendered, err := t.renderAction(httpAction.Config.Body, data)
		if err != nil {
			return m, err
		}

		m[httpAction.Name] = rendered
	}

	return m, nil
}

func (t *Template) renderAction(actionTemplate string, data any) (string, error) {
	templater, err := template.New(t.Name).Parse(actionTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := templater.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func LoadTemplates(dir string) ([]Template, error) {
	tmplFiles, err := getYamlFileNamesFromDir(dir)
	if err != nil {
		return []Template{}, err
	}

	var templates []Template

	for _, file := range tmplFiles {
		b, err := os.ReadFile(file)
		if err != nil {
			return templates, fmt.Errorf("error reading template file '%s': %w", file, err)
		}

		var t Template

		if err := yaml.Unmarshal(b, &t); err != nil {
			return templates, err
		}

		templates = append(templates, t)

	}

	return templates, nil
}
