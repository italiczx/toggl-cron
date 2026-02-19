package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/italiczx/toggl-cron/internal/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current configuration and scheduled entries",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfgPath := config.DefaultPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("no config found at %s — run `toggl-cron setup` first", cfgPath)
	}

	fmt.Printf("Config: %s\n", cfgPath)
	fmt.Printf("Workspace: %s (ID: %d)\n", cfg.WorkspaceName, cfg.WorkspaceID)
	fmt.Printf("Schedules: %d\n\n", len(cfg.Schedules))

	if len(cfg.Schedules) == 0 {
		fmt.Println("No schedules configured. Run `toggl-cron setup` to add one.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "#\tPROJECT\tTASK\tDESCRIPTION\tDURATION\tBILLABLE\tSTART\tCRON")
	fmt.Fprintln(w, "-\t-------\t----\t-----------\t--------\t--------\t-----\t----")

	for i, s := range cfg.Schedules {
		task := s.TaskName
		if task == "" {
			task = "-"
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%v\t%02d:00\t%s\n",
			i+1, s.ProjectName, task, s.Description, s.Duration, s.Billable, s.StartHour, s.Cron)
	}

	return w.Flush()
}
