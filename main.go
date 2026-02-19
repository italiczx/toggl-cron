package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

type config struct {
	APIToken    string
	WorkspaceID int
	ProjectID   int
	TaskID      int
	Description string
	Billable    bool
	Duration    int
}

type timeEntry struct {
	Billable    bool   `json:"billable"`
	CreatedWith string `json:"created_with"`
	Description string `json:"description"`
	Duration    int    `json:"duration"`
	ProjectID   int    `json:"project_id"`
	Start       string `json:"start"`
	TaskID      int    `json:"task_id"`
	WorkspaceID int    `json:"workspace_id"`
}

func loadConfig() config {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	apiToken := os.Getenv("TOGGL_API_TOKEN")
	if apiToken == "" {
		log.Fatal("TOGGL_API_TOKEN is not set")
	}

	workspaceID, err := strconv.Atoi(os.Getenv("WORKSPACE_ID"))
	if err != nil {
		log.Fatal("WORKSPACE_ID must be a valid integer")
	}

	projectID, err := strconv.Atoi(os.Getenv("PROJECT_ID"))
	if err != nil {
		log.Fatal("PROJECT_ID must be a valid integer")
	}

	taskID, err := strconv.Atoi(os.Getenv("TASK_ID"))
	if err != nil {
		log.Fatal("TASK_ID must be a valid integer")
	}

	billable, _ := strconv.ParseBool(os.Getenv("BILLABLE"))

	duration, err := strconv.Atoi(os.Getenv("DURATION"))
	if err != nil {
		duration = 28800
	}

	description := os.Getenv("DESCRIPTION")
	if description == "" {
		description = "Auto-logged by toggl-cron"
	}

	return config{
		APIToken:    apiToken,
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		TaskID:      taskID,
		Description: description,
		Billable:    billable,
		Duration:    duration,
	}
}

func createTimeEntry(cfg config) {
	now := time.Now().UTC()
	local := now.Local()
	start := time.Date(local.Year(), local.Month(), local.Day(), 8, 0, 0, 0, time.Local)

	entry := timeEntry{
		Billable:    cfg.Billable,
		CreatedWith: "toggl-cron",
		Description: cfg.Description,
		Duration:    cfg.Duration,
		ProjectID:   cfg.ProjectID,
		Start:       start.Format(time.RFC3339),
		TaskID:      cfg.TaskID,
		WorkspaceID: cfg.WorkspaceID,
	}

	body, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal time entry: %v", err)
		return
	}

	url := fmt.Sprintf(
		"https://api.track.toggl.com/api/v9/workspaces/%d/time_entries",
		cfg.WorkspaceID,
	)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(cfg.APIToken, "api_token")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		log.Printf("Failed to send request: %v", err)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response: %v", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("API error (status %d): %s", resp.StatusCode, string(respBody))
		return
	}

	log.Printf("Time entry created: %s", string(respBody))
}

func main() {
	cfg := loadConfig()

	log.Println("Starting toggl-cron...")
	log.Printf("Scheduled to create a %ds entry for project %d daily at 17:00 local time", cfg.Duration, cfg.ProjectID)

	c := cron.New()
	c.AddFunc("0 17 * * 1-5", func() {
		log.Println("Running scheduled time entry creation...")
		createTimeEntry(cfg)
	})
	c.Start()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down...")
	c.Stop()
}
