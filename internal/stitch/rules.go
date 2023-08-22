package stitch

import (
	"io/fs"

	"github.com/mrmonaghan/stitch/internal/files"
	"gopkg.in/yaml.v3"
)

// represents a rule definition
type Rule struct {
	Name          string     `yaml:"name" json:"name"`
	Enabled       bool       `yaml:"enabled" json:"enabled"`
	TemplateNames []string   `yaml:"templates" json:"templates"`
	Templates     []Template `yaml:"-" json:"-"`
}

// bind defined templates to a Rule
func (r *Rule) associateTemplates(templates []Template) {
	for _, templateName := range r.TemplateNames {
		for _, template := range templates {
			if template.Name == templateName {
				r.Templates = append(r.Templates, template)
			}
		}
	}
}

// Load Rules from yaml files in the defined directory.
func LoadRules(fsys fs.FS, templates []Template) ([]Rule, error) {

	var rules []Rule

	ruleFiles, err := files.ReadDirYaml(fsys)
	if err != nil {
		return rules, err
	}

	for _, file := range ruleFiles {
		var r Rule

		if err := yaml.Unmarshal(file.Contents, &r); err != nil {
			return rules, err
		}
		r.associateTemplates(templates)
		rules = append(rules, r)
	}
	return rules, nil
}
