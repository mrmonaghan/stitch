package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mrmonaghan/stitch/internal/handlers/mocks"
	"github.com/mrmonaghan/stitch/internal/stitch"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

var testTemplate = stitch.Template{
	Name: "test-template-1",
	Actions: struct {
		Slack []stitch.SlackAction `yaml:"slack,flow"`
		HTTP  []stitch.HTTPAction  `yaml:"http,flow"`
	}{
		Slack: []stitch.SlackAction{
			{
				Name: "slack-action-1",
				Config: stitch.SlackConfig{
					Channels: []string{"test-channel-1"},
					Blocks:   false,
					Message:  `test slack message`,
				},
			},
		},
		HTTP: []stitch.HTTPAction{
			{
				Name: "http-action-1",
				Config: stitch.HTTPConfig{
					Method: http.MethodPost,
					URL:    "https://fake-url.com/fakepath",
					Headers: []stitch.Header{
						{
							Name:  "Content-Type",
							Value: "application/json",
						},
					},
					Body: `test http body`,
				},
			},
		},
	},
}

var testRules = []stitch.Rule{
	{
		Name:          "test-rule-1",
		Enabled:       true,
		TemplateNames: []string{"test-template-1"},
		Templates:     []stitch.Template{testTemplate},
	},
}

func TestHandler_HandleWebhooks(t *testing.T) {
	type fields struct {
		Rules  []stitch.Rule
		Logger *zap.SugaredLogger
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name             string
		setupSlackClient func(*gomock.Controller) *mocks.MockSlackClient
		setupHTTPClient  func(*gomock.Controller) *mocks.MockHTTPClient
		wantStatusCode   int
		fields           fields
		args             args
	}{
		{
			name: "happy path",
			fields: fields{
				Rules:  testRules,
				Logger: zap.NewNop().Sugar(),
			},
			args: args{
				r: httptest.NewRequest(http.MethodPost, "https://fake-url.com/webhook?rule=test-rule-1", strings.NewReader(`{"testVal": "testing-value"}`)),
			},
			setupSlackClient: func(ctrl *gomock.Controller) *mocks.MockSlackClient {
				c := mocks.NewMockSlackClient(ctrl)
				c.EXPECT().PostMessage("test-channel-1", gomock.Any()).Times(1)

				return c
			},
			setupHTTPClient: func(ctrl *gomock.Controller) *mocks.MockHTTPClient {

				resp := &http.Response{
					StatusCode: 200,
				}

				c := mocks.NewMockHTTPClient(ctrl)
				c.EXPECT().Do(gomock.Any()).Times(1).Return(resp, nil)
				return c
			},
			wantStatusCode: 200,
		},
		{
			name: "no rule found",
			fields: fields{
				Rules:  testRules,
				Logger: zap.NewNop().Sugar(),
			},
			args: args{
				r: httptest.NewRequest(http.MethodPost, "https://fake-url.com/webhook?rule=test-rule-invalid", strings.NewReader(`{"testVal": "testing-value"}`)),
			},
			setupSlackClient: func(ctrl *gomock.Controller) *mocks.MockSlackClient {
				return mocks.NewMockSlackClient(ctrl)
			},
			setupHTTPClient: func(ctrl *gomock.Controller) *mocks.MockHTTPClient {
				return mocks.NewMockHTTPClient(ctrl)
			},
			wantStatusCode: 404,
		},
		{
			name: "no rule included",
			fields: fields{
				Rules:  testRules,
				Logger: zap.NewNop().Sugar(),
			},
			args: args{
				r: httptest.NewRequest(http.MethodPost, "https://fake-url.com/webhook?somethingelse=123", strings.NewReader(`{"testVal": "testing-value"}`)),
			},
			setupSlackClient: func(ctrl *gomock.Controller) *mocks.MockSlackClient {
				return mocks.NewMockSlackClient(ctrl)
			},
			setupHTTPClient: func(ctrl *gomock.Controller) *mocks.MockHTTPClient {
				return mocks.NewMockHTTPClient(ctrl)
			},
			wantStatusCode: 400,
		},
		{
			name: "error executing slack action",
			fields: fields{
				Rules:  testRules,
				Logger: zap.NewNop().Sugar(),
			},
			args: args{
				r: httptest.NewRequest(http.MethodPost, "https://fake-url.com/webhook?rule=test-rule-1", strings.NewReader(`{"testVal": "testing-value"}`)),
			},
			setupSlackClient: func(ctrl *gomock.Controller) *mocks.MockSlackClient {
				c := mocks.NewMockSlackClient(ctrl)

				c.EXPECT().PostMessage("test-channel-1", gomock.Any()).Return("", "", errors.New("TEST ERROR - could not execute slack action"))
				return c
			},
			setupHTTPClient: func(ctrl *gomock.Controller) *mocks.MockHTTPClient {
				resp := &http.Response{
					StatusCode: 200,
				}
				c := mocks.NewMockHTTPClient(ctrl)
				c.EXPECT().Do(gomock.Any()).Times(1).Return(resp, nil)
				return c
			},
			wantStatusCode: 500,
		},
		{
			name: "error executing HTTP action",
			fields: fields{
				Rules:  testRules,
				Logger: zap.NewNop().Sugar(),
			},
			args: args{
				r: httptest.NewRequest(http.MethodPost, "https://fake-url.com/webhook?rule=test-rule-1", strings.NewReader(`{"testVal": "testing-value"}`)),
			},
			setupSlackClient: func(ctrl *gomock.Controller) *mocks.MockSlackClient {
				c := mocks.NewMockSlackClient(ctrl)

				c.EXPECT().PostMessage("test-channel-1", gomock.Any())
				return c
			},
			setupHTTPClient: func(ctrl *gomock.Controller) *mocks.MockHTTPClient {
				resp := &http.Response{
					StatusCode: 404,
				}
				c := mocks.NewMockHTTPClient(ctrl)
				c.EXPECT().Do(gomock.Any()).Times(1).Return(resp, errors.New("TEST ERROR - unable to execute HTTP action"))
				return c
			},
			wantStatusCode: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			rec := httptest.NewRecorder()

			slackClient := tt.setupSlackClient(ctrl)

			httpClient := tt.setupHTTPClient(ctrl)

			h := &Handler{
				Rules:       tt.fields.Rules,
				Logger:      tt.fields.Logger,
				SlackClient: slackClient,
				HTTPClient:  httpClient,
			}
			h.HandleWebhooks(rec, tt.args.r)

			resp := rec.Result()

			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("Handler.HandleWebhook() expected statusCode %d, got %d", tt.wantStatusCode, resp.StatusCode)
			}
		})
	}
}

func TestHandler_HandleRules(t *testing.T) {
	type fields struct {
		Rules  []stitch.Rule
		Logger *zap.SugaredLogger
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name             string
		fields           fields
		setupSlackClient func(*gomock.Controller) *mocks.MockSlackClient
		setupHTTPClient  func(*gomock.Controller) *mocks.MockHTTPClient
		args             args
		wantStatusCode   int
	}{
		{
			name: "get from query param",
			fields: fields{
				Rules:  testRules,
				Logger: zap.NewNop().Sugar(),
			},
			setupSlackClient: func(ctrl *gomock.Controller) *mocks.MockSlackClient {
				return mocks.NewMockSlackClient(ctrl)

			},
			setupHTTPClient: func(ctrl *gomock.Controller) *mocks.MockHTTPClient {
				return mocks.NewMockHTTPClient(ctrl)

			},
			args: args{
				r: httptest.NewRequest(http.MethodGet, "https://fake-url.com/rules", nil),
			},
			wantStatusCode: 200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			rec := httptest.NewRecorder()

			slackClient := tt.setupSlackClient(ctrl)
			httpClient := tt.setupHTTPClient(ctrl)

			h := &Handler{
				Rules:       tt.fields.Rules,
				Logger:      tt.fields.Logger,
				SlackClient: slackClient,
				HTTPClient:  httpClient,
			}
			h.HandleRules(rec, tt.args.r)

			resp := rec.Result()

			if tt.wantStatusCode != resp.StatusCode {
				t.Errorf("Handler.HandleRules() expected statusCode %d, got %d", tt.wantStatusCode, resp.StatusCode)
			}
		})
	}
}
