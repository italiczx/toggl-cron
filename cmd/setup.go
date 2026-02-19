package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/italiczx/toggl-cron/internal/config"
	"github.com/italiczx/toggl-cron/internal/toggl"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup wizard — configure your Toggl time entries",
	RunE:  runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) error {
	cfgPath := config.DefaultPath()

	var apiToken string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Toggl API Token").
				Description("Find it at https://track.toggl.com/profile").
				Value(&apiToken).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("API token is required")
					}
					return nil
				}),
		),
	).Run()
	if err != nil {
		return err
	}
	apiToken = strings.TrimSpace(apiToken)

	client := toggl.NewClient(apiToken)

	fmt.Println("\nFetching your Toggl account info...")
	me, err := client.GetMe()
	if err != nil {
		return fmt.Errorf("failed to authenticate — check your API token: %w", err)
	}
	fmt.Printf("Authenticated as %s (%s)\n\n", me.Fullname, me.Email)

	workspaces, err := client.GetWorkspaces()
	if err != nil {
		return fmt.Errorf("fetch workspaces: %w", err)
	}
	if len(workspaces) == 0 {
		return fmt.Errorf("no workspaces found")
	}

	var selectedWorkspace int
	if len(workspaces) == 1 {
		selectedWorkspace = 0
		fmt.Printf("Using workspace: %s\n\n", workspaces[0].Name)
	} else {
		opts := make([]huh.Option[int], len(workspaces))
		for i, w := range workspaces {
			opts[i] = huh.NewOption(w.Name, i)
		}
		err = huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[int]().
					Title("Select workspace").
					Options(opts...).
					Value(&selectedWorkspace),
			),
		).Run()
		if err != nil {
			return err
		}
	}

	ws := workspaces[selectedWorkspace]

	cfg := &config.Config{
		APIToken:      apiToken,
		WorkspaceID:   ws.ID,
		WorkspaceName: ws.Name,
	}

	var addMore bool
	for first := true; first || addMore; first = false {
		schedule, err := promptSchedule(client, ws.ID)
		if err != nil {
			return err
		}
		cfg.Schedules = append(cfg.Schedules, *schedule)

		addMore = false
		err = huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Add another schedule?").
					Value(&addMore),
			),
		).Run()
		if err != nil {
			return err
		}
	}

	if err := cfg.Save(cfgPath); err != nil {
		return err
	}

	fmt.Printf("\nConfig saved to %s\n", cfgPath)
	fmt.Println("Run `toggl-cron run` to start the scheduler, or `toggl-cron run --once` to test.")
	return nil
}

func promptSchedule(client *toggl.Client, workspaceID int) (*config.Schedule, error) {
	fmt.Println("\nFetching projects...")
	projects, err := client.GetProjects(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("fetch projects: %w", err)
	}
	if len(projects) == 0 {
		return nil, fmt.Errorf("no active projects found in this workspace")
	}

	projectOpts := make([]huh.Option[int], len(projects))
	for i, p := range projects {
		projectOpts[i] = huh.NewOption(p.Name, i)
	}

	var selectedProject int
	err = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Select project").
				Options(projectOpts...).
				Value(&selectedProject),
		),
	).Run()
	if err != nil {
		return nil, err
	}

	proj := projects[selectedProject]

	schedule := &config.Schedule{
		ProjectName: proj.Name,
		ProjectID:   proj.ID,
	}

	tasks, err := client.GetTasks(workspaceID, proj.ID)
	if err != nil {
		return nil, fmt.Errorf("fetch tasks: %w", err)
	}

	if len(tasks) > 0 {
		noTaskOpt := huh.NewOption("(no task)", -1)
		taskOpts := []huh.Option[int]{noTaskOpt}
		for i, t := range tasks {
			taskOpts = append(taskOpts, huh.NewOption(t.Name, i))
		}

		var selectedTask int
		err = huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[int]().
					Title("Select task (optional)").
					Options(taskOpts...).
					Value(&selectedTask),
			),
		).Run()
		if err != nil {
			return nil, err
		}

		if selectedTask >= 0 {
			schedule.TaskName = tasks[selectedTask].Name
			schedule.TaskID = tasks[selectedTask].ID
		}
	}

	var (
		description = "Auto-logged by toggl-cron"
		duration    = "8h"
		billable    bool
		cronExpr    string
		startHour   = "8"
	)

	cronOptions := []huh.Option[string]{
		huh.NewOption("Weekdays at 5pm (0 17 * * 1-5)", "0 17 * * 1-5"),
		huh.NewOption("Every day at 5pm (0 17 * * *)", "0 17 * * *"),
		huh.NewOption("Weekdays at 9am (0 9 * * 1-5)", "0 9 * * 1-5"),
		huh.NewOption("Custom", "custom"),
	}

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Description").
				Value(&description),
			huh.NewInput().
				Title("Duration").
				Description("e.g. 8h, 7h30m, 4h").
				Value(&duration).
				Validate(func(s string) error {
					_, err := config.ParseDuration(s)
					return err
				}),
			huh.NewConfirm().
				Title("Billable?").
				Value(&billable),
			huh.NewInput().
				Title("Start hour").
				Description("Hour of day for the time entry start (0-23)").
				Value(&startHour).
				Validate(func(s string) error {
					h, err := strconv.Atoi(s)
					if err != nil || h < 0 || h > 23 {
						return fmt.Errorf("must be 0-23")
					}
					return nil
				}),
			huh.NewSelect[string]().
				Title("Schedule").
				Options(cronOptions...).
				Value(&cronExpr),
		),
	).Run()
	if err != nil {
		return nil, err
	}

	if cronExpr == "custom" {
		err = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Custom cron expression").
					Description("e.g. 0 17 * * 1-5 (min hour dom month dow)").
					Value(&cronExpr).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return fmt.Errorf("cron expression is required")
						}
						parts := strings.Fields(s)
						if len(parts) != 5 {
							return fmt.Errorf("must have 5 fields: min hour dom month dow")
						}
						return nil
					}),
			),
		).Run()
		if err != nil {
			return nil, err
		}
	}

	hour, _ := strconv.Atoi(startHour)

	schedule.Description = description
	schedule.Duration = duration
	schedule.Billable = billable
	schedule.Cron = cronExpr
	schedule.StartHour = hour

	fmt.Printf("\nSchedule: %q → %s (%s, billable=%v, cron=%s)\n",
		schedule.Description, schedule.ProjectName, schedule.Duration, schedule.Billable, schedule.Cron)

	return schedule, nil
}
