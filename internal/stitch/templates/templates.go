package templates

import "github.com/mrmonaghan/stitch/internal/stitch/actions"

type Template struct {
	Name    string          `yaml:"name"`
	Actions actions.Actions `yaml:"actions"`
}
