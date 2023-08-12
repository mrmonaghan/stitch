package stitch

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Rule struct {
	Name          string     `yaml:"name"`
	Enabled       bool       `yaml:"enabled"`
	TemplateNames []string   `yaml:"templates"`
	Templates     []Template `yaml:"-"`
}

func (r *Rule) associateTemplates(templates []Template) {
	for _, templateName := range r.TemplateNames {
		for _, template := range templates {
			if template.Name == templateName {
				r.Templates = append(r.Templates, template)
			}
		}
	}
}

func LoadRules(dir string, templates []Template) (map[string]Rule, error) {
	files, err := getYamlFileNamesFromDir(dir)
	if err != nil {
		return map[string]Rule{}, err
	}

	rules := make(map[string]Rule)

	for _, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			return rules, err
		}
		var r Rule

		if err := yaml.Unmarshal(b, &r); err != nil {
			return rules, err
		}
		r.associateTemplates(templates)
		rules[r.Name] = r
	}
	return rules, nil
}
