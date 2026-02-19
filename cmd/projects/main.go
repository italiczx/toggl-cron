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

type project struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
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

	url := fmt.Sprintf("https://api.track.toggl.com/api/v9/workspaces/%s/projects", workspaceID)

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

	var projects []project
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		log.Fatal(err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tSTATUS")
	fmt.Fprintln(w, "--\t----\t------")
	for _, p := range projects {
		fmt.Fprintf(w, "%d\t%s\t%s\n", p.ID, p.Name, p.Status)
	}
	w.Flush()
}
