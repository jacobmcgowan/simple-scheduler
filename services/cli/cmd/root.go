package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var ApiUrl string

var rootCmd = &cobra.Command{
	Use:   "simple-scheduler-cli",
	Short: "CLI interface to Simple Scheduler",
	Long: `Simple Scheduler is a tool to schedule and manage recurring and 
one-time jobs. This CLI application allows you to view and manage these jobs.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func GenMarkdownTree(path string) error {
	return doc.GenMarkdownTree(rootCmd, path)
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&ApiUrl, "url", "u", "http://localhost:8080/api", "The URL of the Simple Scheduler API.")
}
