package stitch

import (
	"io/fs"
	"reflect"
	"strings"
	"testing"
)

var testRuleYaml = `name: test-rule-1
enabled: true
templates: [test-template-1]
`

var testRule = Rule{
	Name:          "test-rule-1",
	Enabled:       true,
	TemplateNames: []string{"test-template-1", "test-template-2"},
	Templates:     []Template{testTemplate},
}

var testRuleFiles = []testFile{
	{
		Name:     "test-rule-1.yaml",
		Mode:     fs.ModePerm,
		Contents: []byte(testRuleYaml),
	},
}

func TestRule_associateTemplates(t *testing.T) {
	type fields struct {
		Name          string
		Enabled       bool
		TemplateNames []string
		Templates     []Template
	}
	type args struct {
		templates []Template
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "happy path",
			fields: fields{
				Name:          "test-rule-1",
				Enabled:       true,
				TemplateNames: []string{"test-template-1"},
				Templates:     []Template{},
			},
			args: args{
				templates: []Template{testTemplate},
			},
		},
		{
			name: "only associate test-template-1",
			fields: fields{
				Name:          "test-rule",
				Enabled:       true,
				TemplateNames: []string{"test-template-1"},
				Templates:     []Template{},
			},
			args: args{
				templates: []Template{
					testTemplate,
					{
						Name: "test-template-2",
						Actions: struct {
							Slack []SlackAction `yaml:"slack,flow"`
							HTTP  []HTTPAction  `yaml:"http,flow"`
						}{
							Slack: []SlackAction{
								{
									Name: "slack-action-2",
									Config: SlackConfig{
										Channels: []string{"test-channel-2"},
										Blocks:   false,
										Message:  `test slack message 2`,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Rule{
				Name:          tt.fields.Name,
				Enabled:       tt.fields.Enabled,
				TemplateNames: tt.fields.TemplateNames,
				Templates:     tt.fields.Templates,
			}
			r.associateTemplates(tt.args.templates)

			var hasTemplateNames []string
			for _, template := range r.Templates {
				hasTemplateNames = append(hasTemplateNames, template.Name)
			}

			if !reflect.DeepEqual(hasTemplateNames, r.TemplateNames) {
				t.Errorf("Rule_associateTemplates() expected associated template names '%s', got '%s'", strings.Join(r.TemplateNames, ","), strings.Join(hasTemplateNames, ","))
			}
		})
	}
}

func TestLoadRules(t *testing.T) {
	type args struct {
		fsys      fs.FS
		templates []Template
	}
	tests := []struct {
		name    string
		args    args
		want    []Rule
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				fsys:      newMapFS(testRuleFiles),
				templates: []Template{testTemplate},
			},
			want: []Rule{
				{
					Name:          "test-rule-1",
					Enabled:       true,
					TemplateNames: []string{"test-template-1"},
					Templates:     []Template{testTemplate},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadRules(tt.args.fsys, tt.args.templates)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadRules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadRules() = %v, want %v", got, tt.want)
			}
		})
	}
}
