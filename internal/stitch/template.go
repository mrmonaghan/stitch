package stitch

import (
	"io/fs"

	"github.com/mrmonaghan/stitch/internal/files"
	"gopkg.in/yaml.v3"
)

// represents a Template
type Template struct {
	Name    string `yaml:"name"`
	Actions struct {
		Slack []SlackAction `yaml:"slack,flow"`
		HTTP  []HTTPAction  `yaml:"http,flow"`
	}
}

// Load Templates from yaml files in the defined directory.
func LoadTemplates(fsys fs.FS) ([]Template, error) {

	var templates []Template

	tmplFiles, err := files.ReadDirYaml(fsys)
	if err != nil {
		return templates, err
	}

	for _, file := range tmplFiles {

		var t Template

		if err := yaml.Unmarshal(file.Contents, &t); err != nil {
			return templates, err
		}

		templates = append(templates, t)

	}

	return templates, nil
}
