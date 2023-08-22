package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/slack-go/slack"
)

// Blocks & associated methods allow easier serializing of `blocks` JSON objects returned by the Slack API or obtained from the Block Kit Builder
type Blocks struct {
	Blocks []slack.Block `json:"blocks"`
}
type blockhint struct {
	Typ string `json:"type"`
}

func (b *Blocks) UnmarshalJSON(data []byte) error {
	var proxy struct {
		Blocks []json.RawMessage `json:"blocks"`
	}
	if err := json.Unmarshal(data, &proxy); err != nil {
		return fmt.Errorf(`failed to unmarshal blocks array: %w`, err)
	}
	for _, rawBlock := range proxy.Blocks {
		var hint blockhint
		if err := json.Unmarshal(rawBlock, &hint); err != nil {
			return fmt.Errorf(`failed to unmarshal next block for hint: %w`, err)
		}
		var block slack.Block
		switch hint.Typ {
		case "actions":
			block = &slack.ActionBlock{}
		case "context":
			block = &slack.ContextBlock{}
		case "divider":
			block = &slack.DividerBlock{}
		case "file":
			block = &slack.FileBlock{}
		case "header":
			block = &slack.HeaderBlock{}
		case "image":
			block = &slack.ImageBlock{}
		case "input":
			block = &slack.InputBlock{}
		case "section":
			block = &slack.SectionBlock{}
		default:
			block = &slack.UnknownBlock{}
		}
		if err := json.Unmarshal(rawBlock, block); err != nil {
			return fmt.Errorf(`failed to unmarshal next block: %w`, err)
		}
		b.Blocks = append(b.Blocks, block)
	}
	return nil
}
