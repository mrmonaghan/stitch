package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/mrmonaghan/stitch/internal/stitch"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

// SlackClient describes the interface used by Handler to send Slack messages defined by a stitch.SlackAction
type SlackClient interface {
	PostMessage(string, ...slack.MsgOption) (string, string, error)
}

// HTTPClient describes the interface used by Handler to send HTTP requests defined by a stitch.HTTPAction
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// Handler provides methods for handling HTTP requests made to /webhook and /rules endpoints,
// as well as methods for executing stitch.HTTPAction and stitch.SlackAction
type Handler struct {
	Rules       []stitch.Rule
	Logger      *zap.SugaredLogger
	SlackClient SlackClient
	HTTPClient  HTTPClient
}

// Handle HTTP requests made to the /webhooks endpoint
func (h *Handler) HandleWebhooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	reqLogger := h.Logger.With("request_id", uuid.New().String())
	reqLogger.Debugw("initialized request logger")

	requestedRule, err := h.getRequestedRule(r)
	if err != nil {
		reqLogger.Errorw("unable to determine rule for request", "url", r.URL.RequestURI(), "headers", r.Header, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var rule stitch.Rule
	for _, r := range h.Rules {
		if r.Name == requestedRule {
			rule = r
			break
		}
	}

	if rule.Name == "" {
		reqLogger.Debugw("request rule not found", "rule", requestedRule)
		http.Error(w, fmt.Sprintf("rule '%s' not found", requestedRule), http.StatusNotFound)
		return
	}

	reqLogger = reqLogger.With("rule_name", rule.Name)
	reqLogger.Debugw("located requested rule")

	b, err := io.ReadAll(r.Body)
	if err != nil {
		reqLogger.Errorw("error reading request body", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	data := make(map[string]interface{})
	if err := json.Unmarshal(b, &data); err != nil {
		reqLogger.Errorw("error unmarshalling request body", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	executionErrors := make(map[string]error)
	for _, template := range rule.Templates {

		reqLogger.Debugw("begin processing template", "template_name", template.Name)

		for _, action := range template.Actions.Slack {
			if err := h.executeSlackAction(action, data); err != nil {
				reqLogger.Infow("error executing slack action", "name", action.Name, "template_name", template.Name, err)
				executionErrors[template.Name] = err
			} else {
				reqLogger.Debugw("successfully executed slack action", "name", action.Name, "template_name", template.Name)
			}
		}

		for _, action := range template.Actions.HTTP {
			if err := h.executeHTTPAction(action, data); err != nil {
				reqLogger.Infow("error executing http action", "name", action.Name, "template_name", template.Name, err)
				executionErrors[template.Name] = err
			} else {
				reqLogger.Debugw("successfully executed http action", "name", action.Name, "template_name", template.Name)
			}
		}

		reqLogger.Debugw("successfully processed template", "template_name", template.Name)
	}

	if len(executionErrors) > 0 {
		reqLogger.Infow("encountered one or more errors while executing rule", "errorCount", len(executionErrors))
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	http.Error(w, "", http.StatusOK)
	return
}

// Handle requests made to the /rules endpoint
func (h *Handler) HandleRules(w http.ResponseWriter, r *http.Request) {
	var rules []stitch.Rule

	for _, rule := range h.Rules {
		rules = append(rules, rule)
	}

	b, err := json.Marshal(rules)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	http.Error(w, string(b), http.StatusOK)
	return
}

// helper function parses the requested rule from the URL query parameters or request headers
func (h *Handler) getRequestedRule(r *http.Request) (string, error) {
	if requestedRule := r.URL.Query().Get("rule"); requestedRule != "" {
		return requestedRule, nil
	} else if requestedRule := r.Header.Get("rule"); requestedRule != "" {
		return requestedRule, nil
	} else {
		return "", errors.New("no 'stitch_rule' parameter included in URL or request headers")
	}
}

// render and execute a stitch.SlackAction using SlackClient
func (h *Handler) executeSlackAction(action stitch.SlackAction, data any) error {
	err := action.Render(data)
	if err != nil {
		return err
	}

	r := action.Config.Message

	var slackOpts slack.MsgOption
	if action.Config.Blocks == true {
		var blocks Blocks
		if err := blocks.UnmarshalJSON([]byte(r)); err != nil {
			return fmt.Errorf("unable to process blocks for action '%s': %w", action.Name, err)
		}
		slackOpts = slack.MsgOptionBlocks(blocks.Blocks...)
	} else {
		slackOpts = slack.MsgOptionText(r, false)
	}

	for _, channel := range action.Config.Channels {
		_, _, err := h.SlackClient.PostMessage(channel, slackOpts)
		if err != nil {
			return fmt.Errorf("unable to send slack message for action '%s': %w", action.Name, err)
		}
	}
	return nil
}

// render and execute a stitch.HTTPAction using HTTPClient
func (h *Handler) executeHTTPAction(action stitch.HTTPAction, data any) error {
	req, err := action.Request(data)
	if err != nil {
		return err
	}

	w, err := h.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	if err := action.CheckStatusCode(w.StatusCode); err != nil {
		return err
	}

	return nil
}
