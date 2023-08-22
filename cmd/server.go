package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/mrmonaghan/stitch/internal/files"
	"github.com/mrmonaghan/stitch/internal/handlers"
	"github.com/mrmonaghan/stitch/internal/stitch"
	"github.com/slack-go/slack"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the stitch HTTP server",
	Run: func(cmd *cobra.Command, args []string) {
		debug, _ := cmd.Flags().GetBool("debug")
		tDir, _ := rootCmd.PersistentFlags().GetString("template-dir")
		rDir, _ := rootCmd.PersistentFlags().GetString("rules-dir")

		// initialize logger
		var log *zap.Logger
		var err error
		if debug {
			log, err = zap.NewDevelopment()

		} else {
			log, err = zap.NewProduction()
		}
		if err != nil {
			panic(fmt.Errorf("unable to initialize logger: %w", err))
		}
		logger := log.Sugar()
		logger.Debugw("initialized logger", "debug", debug)

		templateDir, err := files.ResolveDirInput(tDir)
		if err != nil {
			logger.Panicw("error resolving template-dir", err)
		}

		rulesDir, err := files.ResolveDirInput(rDir)
		if err != nil {
			logger.Panicw("error resolving rules-dir", err)
		}

		templates, err := stitch.LoadTemplates(os.DirFS(templateDir))
		if err != nil {
			logger.Panicw("error loading templates", err)
		}

		rules, err := stitch.LoadRules(os.DirFS(rulesDir), templates)
		if err != nil {
			logger.Panicw("error loading rules", err)
		}
		logger.Debugw("loaded rules and templates", "rulesDir", rulesDir, "templateDir", templateDir)

		token := os.Getenv("SLACK_TOKEN")
		if token == "" {
			panic(fmt.Errorf("SLACK_TOKEN cannot be null"))
		}

		slack := slack.New(token)

		httpClient := http.Client{}

		// create handler
		handler := handlers.Handler{
			Rules:       rules,
			Logger:      logger,
			SlackClient: slack,
			HTTPClient:  &httpClient,
		}

		// initialize server
		mux := http.NewServeMux()
		mux.HandleFunc("/webhook", handler.HandleWebhooks)
		mux.HandleFunc("/rules", handler.HandleRules)

		port, err := cmd.Flags().GetString("port")
		if err != nil {
			logger.Panicw("error retrieving --port flag value", err)
		}

		if err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux); err != nil {
			panic(fmt.Errorf("error starting HTTP server: %w", err))
		}
	},
}

func init() {
	serverCmd.Flags().String("port", "8888", "specify a port for server to bind to")
	serverCmd.Flags().Bool("debug", false, "enable debug-level logging")
	rootCmd.AddCommand(serverCmd)
}
