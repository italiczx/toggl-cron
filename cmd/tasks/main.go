package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/joho/godotenv"
)

type task struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	apiToken := os.Getenv("TOGGL_API_TOKEN")
	if apiToken == "" {
		log.Fatal("TOGGL_API_TOKEN is not set")
	}

	workspaceID := os.Getenv("WORKSPACE_ID")
	if workspaceID == "" {
		log.Fatal("WORKSPACE_ID is not set")
	}

	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		log.Fatal("PROJECT_ID is not set — run `go run ./cmd/projects` first to find it")
	}

	url := fmt.Sprintf(
		"https://api.track.toggl.com/api/v9/workspaces/%s/projects/%s/tasks",
		workspaceID, projectID,
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(apiToken, "api_token")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("API returned status %d", resp.StatusCode)
	}

	var tasks []task
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		log.Fatal(err)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found for this project.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tACTIVE")
	fmt.Fprintln(w, "--\t----\t------")
	for _, t := range tasks {
		fmt.Fprintf(w, "%d\t%s\t%v\n", t.ID, t.Name, t.Active)
	}
	w.Flush()
}
