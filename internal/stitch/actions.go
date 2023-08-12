package stitch

type SlackExecutor func(string, SlackConfig) error
type HTTPExecutor func(string, HTTPConfig) error

type SlackAction struct {
	Name   string      `yaml:"name"`
	Type   string      `yaml:"type"`
	Config SlackConfig `yaml:"config"`
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

type HTTPConfig struct {
	Method  string              `yaml:"method"`
	Path    string              `yaml:"path"`
	Headers []map[string]string `yaml:"headers"`
	Body    string              `yaml:"body"`
}
