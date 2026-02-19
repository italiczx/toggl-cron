package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/italiczx/toggl-cron/internal/config"
	"github.com/italiczx/toggl-cron/internal/toggl"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

var runOnce bool

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the scheduler (or fire once with --once)",
	RunE:  runRun,
}

func init() {
	runCmd.Flags().BoolVar(&runOnce, "once", false, "Create time entries immediately and exit")
	rootCmd.AddCommand(runCmd)
}

func runRun(cmd *cobra.Command, args []string) error {
	cfgPath := config.DefaultPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config from %s: %w\nRun `toggl-cron setup` first", cfgPath, err)
	}

	if len(cfg.Schedules) == 0 {
		return fmt.Errorf("no schedules configured — run `toggl-cron setup` to add one")
	}

	client := toggl.NewClient(cfg.APIToken)

	if runOnce {
		return fireAll(client, cfg)
	}

	return startCron(client, cfg)
}

func fireAll(client *toggl.Client, cfg *config.Config) error {
	for i, s := range cfg.Schedules {
		log.Printf("[%d/%d] Creating entry: %q → %s (%s)",
			i+1, len(cfg.Schedules), s.Description, s.ProjectName, s.Duration)

		if err := createEntry(client, cfg.WorkspaceID, s); err != nil {
			log.Printf("  Error: %v", err)
		} else {
			log.Printf("  Done")
		}
	}
	return nil
}

func startCron(client *toggl.Client, cfg *config.Config) error {
	c := cron.New()

	for _, s := range cfg.Schedules {
		s := s
		_, err := c.AddFunc(s.Cron, func() {
			log.Printf("Running scheduled entry: %q → %s", s.Description, s.ProjectName)
			if err := createEntry(client, cfg.WorkspaceID, s); err != nil {
				log.Printf("  Error: %v", err)
			} else {
				log.Printf("  Done")
			}
		})
		if err != nil {
			return fmt.Errorf("invalid cron expression %q for schedule %q: %w", s.Cron, s.Description, err)
		}
	}

	c.Start()

	log.Println("toggl-cron is running. Scheduled entries:")
	for _, s := range cfg.Schedules {
		dur := s.Duration
		log.Printf("  • %q → %s (%s, cron: %s)", s.Description, s.ProjectName, dur, s.Cron)
	}
	log.Println("Press Ctrl+C to stop.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down...")
	c.Stop()
	return nil
}

func createEntry(client *toggl.Client, workspaceID int, s config.Schedule) error {
	dur, err := config.ParseDuration(s.Duration)
	if err != nil {
		return err
	}

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), s.StartHour, 0, 0, 0, time.Local)

	entry := toggl.TimeEntry{
		Billable:    s.Billable,
		CreatedWith: "toggl-cron",
		Description: s.Description,
		Duration:    dur,
		ProjectID:   s.ProjectID,
		Start:       start.Format(time.RFC3339),
		TaskID:      s.TaskID,
		WorkspaceID: workspaceID,
	}

	return client.CreateTimeEntry(workspaceID, entry)
}

