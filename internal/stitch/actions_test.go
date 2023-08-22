package stitch

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

var testVal = "test template value"

var testData = map[string]interface{}{
	"testVal": testVal,
}

func TestSlackAction_Render(t *testing.T) {
	type fields struct {
		Name   string
		Config SlackConfig
	}
	type args struct {
		data any
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantMessage string
		wantErr     bool
	}{
		{
			name: "happy path",
			fields: fields{
				Name: "test-slack-acton",
				Config: SlackConfig{
					Channels: []string{
						"test-channel-1",
					},
					Blocks:  false,
					Message: `this is a test message {{ .testVal }}`,
				},
			},
			args: args{
				data: testData,
			},
			wantMessage: `this is a test message test template value`,
			wantErr:     false,
		},
		{
			name: "invalid template",
			fields: fields{
				Name: "test-slack-acton",
				Config: SlackConfig{
					Channels: []string{
						"test-channel-1",
					},
					Blocks:  false,
					Message: `this is a test message { testVal }}`,
				},
			},
			args: args{
				data: testData,
			},
			wantMessage: `this is a test message { testVal }}`,
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &SlackAction{
				Name:   tt.fields.Name,
				Config: tt.fields.Config,
			}
			if err := a.Render(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("SlackAction.Render() error = %v, wantErr %v", err, tt.wantErr)
			}

			if a.Config.Message != tt.wantMessage {
				t.Errorf("SlackAction.Render() expect message '%s', got message '%s'", tt.wantMessage, a.Config.Message)
			}
		})
	}
}

func TestSlackAction_String(t *testing.T) {
	type fields struct {
		Name   string
		Config SlackConfig
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "happy path",
			fields: fields{
				Name: "test-slack-action",
				Config: SlackConfig{
					Channels: []string{
						"test-channel-1",
					},
					Blocks:  false,
					Message: `this is a test message`,
				},
			},
			want: `name: test-slack-action
config:
    channels:
        - test-channel-1
    blocks: false
    message: this is a test message
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &SlackAction{
				Name:   tt.fields.Name,
				Config: tt.fields.Config,
			}
			if got := a.String(); got != tt.want {
				t.Errorf("SlackAction.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPAction_Render(t *testing.T) {
	type fields struct {
		Name   string
		Type   string
		Config HTTPConfig
	}
	type args struct {
		data any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				Name: "test-http-action",
				Type: "http",
				Config: HTTPConfig{
					Method: http.MethodPost,
					URL:    "https://fake-url.com/fakepath",
					Headers: []Header{
						{
							Name:  "Content-Type",
							Value: "application/json",
						},
					},
					Body: `body text {{ .testVal }}`,
				},
			},
			args: args{
				data: testData,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &HTTPAction{
				Name:   tt.fields.Name,
				Type:   tt.fields.Type,
				Config: tt.fields.Config,
			}
			if err := a.Render(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("HTTPAction.Render() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHTTPAction_renderHeaders(t *testing.T) {
	type fields struct {
		Name   string
		Type   string
		Config HTTPConfig
	}
	type args struct {
		data any
	}
	tests := []struct {
		name                string
		fields              fields
		args                args
		wantTestHeaderValue string
		wantErr             bool
	}{
		{
			name: "happy path",
			args: args{
				data: testData,
			},
			fields: fields{
				Name: "test-http-action",
				Type: "http",
				Config: HTTPConfig{
					Method: "post",
					URL:    "https://test-url.com/test-path",
					Headers: []Header{
						{
							Name:  "TestHeader",
							Value: "{{ .testVal }}",
						},
					},
				},
			},
			wantTestHeaderValue: testVal,
			wantErr:             false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &HTTPAction{
				Name:   tt.fields.Name,
				Type:   tt.fields.Type,
				Config: tt.fields.Config,
			}
			if err := a.renderHeaders(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("HTTPAction.renderHeaders() error = %v, wantErr %v", err, tt.wantErr)
			}

			var gotTestHeaderVal string
			for _, header := range a.Config.Headers {
				if header.Name == "TestHeader" {
					gotTestHeaderVal = header.Value
					break
				}
				t.Error("missing header: TestHeader")
			}

			if gotTestHeaderVal != tt.wantTestHeaderValue {
				t.Errorf("HTTPAction.renderHeaders() expected TestHeader value '%s', got TestHeader value '%s'", tt.wantTestHeaderValue, gotTestHeaderVal)
			}
		})
	}
}

func TestHTTPAction_renderUrl(t *testing.T) {
	type fields struct {
		Name   string
		Type   string
		Config HTTPConfig
	}
	type args struct {
		data any
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantURLValue string
		wantErr      bool
	}{
		{
			name: "happy path",
			args: args{
				data: testData,
			},
			fields: fields{
				Name: "test-http-action",
				Type: "http",
				Config: HTTPConfig{
					Method: "post",
					URL:    "https://test-url.com/test-path/{{ .testVal }}",
				},
			},
			wantURLValue: `https://test-url.com/test-path/test template value`,
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &HTTPAction{
				Name:   tt.fields.Name,
				Type:   tt.fields.Type,
				Config: tt.fields.Config,
			}
			if err := a.renderUrl(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("HTTPAction.renderUrl() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantURLValue != a.Config.URL {
				t.Errorf("HTTPAction.renderUrl() expected URL '%s', got URL '%s'", tt.wantURLValue, a.Config.URL)
			}
		})
	}
}

func TestHTTPAction_Request(t *testing.T) {
	type fields struct {
		Name   string
		Type   string
		Config HTTPConfig
	}
	type args struct {
		method   string
		url      string
		bodyText string
		data     any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				Name: "test-http-action",
				Type: "http",
				Config: HTTPConfig{
					Method: http.MethodPost,
					URL:    "https://fake-url.com/fakepath",
					Headers: []Header{
						{
							Name:  "Content-Type",
							Value: "application/json",
						},
					},
					Body: `body text`,
				},
			},
			args: args{
				data:     testData,
				method:   http.MethodPost,
				url:      "https://fake-url.com/fakepath",
				bodyText: `body text`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			wantReq, err := http.NewRequest(tt.args.method, tt.args.url, strings.NewReader(tt.args.bodyText))
			if err != nil {
				t.Errorf("error creating wantReq: %s", err.Error())
			}

			for _, header := range tt.fields.Config.Headers {
				wantReq.Header.Add(header.Name, header.Value)
			}

			a := &HTTPAction{
				Name:   tt.fields.Name,
				Type:   tt.fields.Type,
				Config: tt.fields.Config,
			}
			got, err := a.Request(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPAction.Request() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if wantReq.URL.String() != got.URL.String() {
				t.Errorf("HTTPAction.Request() expected URL '%s', got '%s'", wantReq.URL.String(), got.URL.String())
			}

			if wantReq.Header.Get("Content-Type") != got.Header.Get("Content-Type") {
				t.Errorf("HTTPAction.Request() expected Content-Type header value '%s', got '%s'", wantReq.Header.Get("Content-Type"), got.Header.Get("Content-Type"))
			}

			if wantReq.Method != got.Method {
				t.Errorf("HTTPAction.Request() expected Method  '%s', got '%s'", wantReq.Method, got.Method)
			}
		})
	}
}

func TestHTTPAction_CheckStatusCode(t *testing.T) {
	type fields struct {
		Name   string
		Type   string
		Config HTTPConfig
	}
	type args struct {
		code int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				code: 200,
			},
			fields: fields{
				Name: "test-http-action",
				Type: "http",
				Config: HTTPConfig{
					Method: "post",
					URL:    "https://fake-url/test-path",
					Body:   "body",
					StatusCodes: struct {
						Success []int `yaml:"success,omitempty"`
						Failure []int `yaml:"failure,omitempty"`
					}{
						Success: []int{200},
						Failure: []int{400},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "error on 200",
			args: args{
				code: 200,
			},
			fields: fields{
				Name: "test-http-action",
				Type: "http",
				Config: HTTPConfig{
					Method: "post",
					URL:    "https://fake-url/test-path",
					Body:   "body",
					StatusCodes: struct {
						Success []int `yaml:"success,omitempty"`
						Failure []int `yaml:"failure,omitempty"`
					}{
						Success: []int{400},
						Failure: []int{200},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "no codes defined",
			args: args{
				code: 200,
			},
			fields: fields{
				Name: "test-http-action",
				Type: "http",
				Config: HTTPConfig{
					Method: "post",
					URL:    "https://fake-url/test-path",
					Body:   "body",
					StatusCodes: struct {
						Success []int `yaml:"success,omitempty"`
						Failure []int `yaml:"failure,omitempty"`
					}{
						Success: []int{},
						Failure: []int{},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "succeed on 400",
			args: args{
				code: 400,
			},
			fields: fields{
				Name: "test-http-action",
				Type: "http",
				Config: HTTPConfig{
					Method: "post",
					URL:    "https://fake-url/test-path",
					Body:   "body",
					StatusCodes: struct {
						Success []int `yaml:"success,omitempty"`
						Failure []int `yaml:"failure,omitempty"`
					}{
						Success: []int{400},
						Failure: []int{},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &HTTPAction{
				Name:   tt.fields.Name,
				Type:   tt.fields.Type,
				Config: tt.fields.Config,
			}
			if err := a.CheckStatusCode(tt.args.code); (err != nil) != tt.wantErr {
				t.Errorf("HTTPAction.CheckStatusCode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHTTPAction_String(t *testing.T) {
	type fields struct {
		Name   string
		Type   string
		Config HTTPConfig
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "happy path",
			fields: fields{
				Name: "test-http-action",
				Type: "http",
				Config: HTTPConfig{
					Method: "post",
					URL:    "https://fake-url.com/fakepath",
					Body:   `body template {{ .testVal }}`,
					Headers: []Header{
						{
							Name:  "Content-Type",
							Value: "application/json",
						},
					},
				},
			},
			want: `name: test-http-action
type: http
config:
    method: post
    url: https://fake-url.com/fakepath
    headers:
        - name: Content-Type
          value: application/json
    body: body template {{ .testVal }}
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &HTTPAction{
				Name:   tt.fields.Name,
				Type:   tt.fields.Type,
				Config: tt.fields.Config,
			}
			if got := a.String(); got != tt.want {
				fmt.Printf("%s\n", got)
				fmt.Printf("%s\n", tt.want)

				t.Errorf("HTTPAction.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_render(t *testing.T) {
	type args struct {
		name string
		tmpl string
		data any
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				name: "a",
				tmpl: "this is test data - {{ .testVal }}",
				data: testData,
			},
			want:    "this is test data - test template value",
			wantErr: false,
		},
		{
			name: "invalid template",
			args: args{
				name: "a",
				tmpl: "this is test data - { testVal }}",
				data: testData,
			},
			want:    "this is test data - { testVal }}",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := render(tt.args.name, tt.args.tmpl, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("render() = %v, want %v", got, tt.want)
			}
		})
	}
}
