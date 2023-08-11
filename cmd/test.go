/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mrmonaghan/stitch/internal/stitch/actions"
	"github.com/slack-go/slack"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var data = []byte(`
name: test-template
actions:
  slack:
    - name: slack-1
      type: slack
      config:
        channels:
          - mike-notification-testing
        blocks: false
        message: test message {{ .name }}
  http:
    - name: http-1
      type: http
      config:
        method: get
        path: https://fake-url.com/webhook
        headers:
          - name: header1
            value: value1
          - name: header2
            value: value2
        body: http body {{ .name }}
`)

var jsonData = []byte(`
{
	"name": "mike-rules"
}`)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		var t actions.Template

		if err := yaml.Unmarshal(data, &t); err != nil {
			panic(err)
		}

		m := make(map[string]interface{})

		if err := json.Unmarshal(jsonData, &m); err != nil {
			panic(err)
		}

		token := os.Getenv("SLACK_TOKEN")
		if token == "" {
			panic(fmt.Errorf("SLACK_TOKEN is nil"))
		}

		slackClient := slack.New(token)
		fmt.Printf("%#v\n", slackClient)

		exec, err := actions.NewSlackExecutor(slackClient)
		if err != nil {
			panic(err)
		}

		for _, slackAction := range t.Actions.Slack {
			if err := exec.Execute(slackAction, m); err != nil {
				panic(err)
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
