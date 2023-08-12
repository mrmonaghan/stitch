/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/mrmonaghan/stitch/internal/stitch"
	"github.com/spf13/cobra"
)

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]
		data := args[1]
		templateDir, _ := rootCmd.PersistentFlags().GetString("template-dir")

		tpls, err := stitch.LoadTemplates(templateDir)
		if err != nil {
			panic(fmt.Errorf("unable to load templates: %w", err))
		}

		var tmpl stitch.Template
		found := false
		for _, template := range tpls {
			if template.Name == templateName {
				tmpl = template
				found = true
				break
			}
		}

		if !found {
			fmt.Printf("no template '%s' found in template directory '%s'\n", templateName, templateDir)
		}

		m := make(map[string]interface{})

		if err := json.Unmarshal([]byte(data), &m); err != nil {
			panic(fmt.Errorf("error processing template data: %w", err))
		}

		renderedActions, err := tmpl.RenderAll(m)
		if err != nil {
			panic(err)
		}

		for action, rendered := range renderedActions {
			fmt.Printf("|| --- template: %s | action: %s --- ||\n%s\n", tmpl.Name, action, rendered)
		}
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// templateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// templateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
