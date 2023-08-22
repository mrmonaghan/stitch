package stitch

import (
	"io/fs"
	"net/http"
	"reflect"
	"testing"
)

var testTemplateYaml = `name: test-template-1
actions:
    slack:
        - name: slack-action-1
          config:
              channels:
                  - test-channel-1
              blocks: false
              message: test message {{ .testVal }}
    http:
        - name: http-action-1
          config:
              method: POST
              url: https://fake-url.com/fakepath
              headers:
                  - name: Content-Type
                    value: application/json
                  - name: TestHeader
                    value: "{{ .testVal }}"
`

var testTemplate = Template{
	Name: "test-template-1",
	Actions: struct {
		Slack []SlackAction `yaml:"slack,flow"`
		HTTP  []HTTPAction  `yaml:"http,flow"`
	}{
		Slack: []SlackAction{
			{
				Name: "slack-action-1",
				Config: SlackConfig{
					Channels: []string{"test-channel-1"},
					Blocks:   false,
					Message:  `test message {{ .testVal }}`,
				},
			},
		},
		HTTP: []HTTPAction{
			{
				Name: "http-action-1",
				Config: HTTPConfig{
					Method: http.MethodPost,
					URL:    "https://fake-url.com/fakepath",
					Headers: []Header{
						{
							Name:  "Content-Type",
							Value: "application/json",
						},
						{
							Name:  "TestHeader",
							Value: "{{ .testVal }}",
						},
					},
				},
			},
		},
	},
}

var testTemplateFiles = []testFile{
	{
		Name:     "test-template-1.yaml",
		Mode:     fs.ModePerm,
		Contents: []byte(testTemplateYaml),
	},
}

func TestLoadTemplates(t *testing.T) {
	type args struct {
		fsys fs.FS
	}
	tests := []struct {
		name    string
		args    args
		want    []Template
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				fsys: newMapFS(testTemplateFiles),
			},
			want: []Template{testTemplate},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadTemplates(tt.args.fsys)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadTemplates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadTemplates() = %v, want %v", got, tt.want)
			}
		})
	}
}
