package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mrmonaghan/stitch/internal/files"
	"github.com/mrmonaghan/stitch/internal/stitch"
	"github.com/spf13/cobra"
)

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:  `template <template-name> '{"key1": "val1"}'`,
	Long: `The template command hydrates the specified template with the provided data, then prints the results.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]
		data := args[1]
		templateDir, _ := rootCmd.PersistentFlags().GetString("template-dir")

		d, err := files.ResolveDirInput(templateDir)
		if err != nil {
			panic(err)
		}

		tpls, err := stitch.LoadTemplates(os.DirFS(d))
		if err != nil {
			panic(fmt.Errorf("unable to load templates: %w", err))
		}

		fmt.Println(tpls)

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

		renderedActions := make(map[string]string)

		for _, action := range tmpl.Actions.Slack {
			if err := action.Render(m); err != nil {
				panic(fmt.Errorf("unable to render template '%s' action '%s': %w", tmpl.Name, action.Name, err))
			}
			renderedActions[action.Name] = action.String()
		}

		for _, action := range tmpl.Actions.HTTP {
			if err := action.Render(m); err != nil {
				panic(fmt.Errorf("unable to render template '%s' action '%s': %w", tmpl.Name, action.Name, err))
			}
			renderedActions[action.Name] = action.String()
		}

		fmt.Printf("|| --- rendered actions for template: %s --- ||\n", tmpl.Name)
		for _, rendered := range renderedActions {
			fmt.Println(rendered)
		}
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
}
