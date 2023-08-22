package handlers

import (
	"testing"

	"github.com/slack-go/slack"
)

func TestBlocks_UnmarshalJSON(t *testing.T) {
	type fields struct {
		Blocks []slack.Block
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Blocks{
				Blocks: tt.fields.Blocks,
			}
			if err := b.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Blocks.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
