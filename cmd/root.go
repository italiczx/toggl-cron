package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "toggl-cron",
	Short: "Automatically create Toggl time entries on a schedule",
	Long: `toggl-cron is a CLI tool that automatically creates Toggl time entries
on a cron schedule. Run 'toggl-cron setup' to get started.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
