package toggl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://api.track.toggl.com/api/v9"

type Client struct {
	apiToken   string
	httpClient *http.Client
}

func NewClient(apiToken string) *Client {
	return &Client{
		apiToken:   apiToken,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

type Me struct {
	ID                 int    `json:"id"`
	Email              string `json:"email"`
	Fullname           string `json:"fullname"`
	DefaultWorkspaceID int    `json:"default_workspace_id"`
}

type Workspace struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Project struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Active bool   `json:"active"`
}

type Task struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type TimeEntry struct {
	Billable    bool   `json:"billable"`
	CreatedWith string `json:"created_with"`
	Description string `json:"description"`
	Duration    int    `json:"duration"`
	ProjectID   int    `json:"project_id"`
	Start       string `json:"start"`
	TaskID      int    `json:"task_id,omitempty"`
	WorkspaceID int    `json:"workspace_id"`
}

func (c *Client) do(method, url string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(c.apiToken, "api_token")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *Client) GetMe() (*Me, error) {
	data, err := c.do(http.MethodGet, baseURL+"/me", nil)
	if err != nil {
		return nil, err
	}
	var me Me
	if err := json.Unmarshal(data, &me); err != nil {
		return nil, fmt.Errorf("decode /me: %w", err)
	}
	return &me, nil
}

func (c *Client) GetWorkspaces() ([]Workspace, error) {
	data, err := c.do(http.MethodGet, baseURL+"/workspaces", nil)
	if err != nil {
		return nil, err
	}
	var ws []Workspace
	if err := json.Unmarshal(data, &ws); err != nil {
		return nil, fmt.Errorf("decode workspaces: %w", err)
	}
	return ws, nil
}

func (c *Client) GetProjects(workspaceID int) ([]Project, error) {
	url := fmt.Sprintf("%s/workspaces/%d/projects?active=true", baseURL, workspaceID)
	data, err := c.do(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("decode projects: %w", err)
	}
	return projects, nil
}

func (c *Client) GetTasks(workspaceID, projectID int) ([]Task, error) {
	url := fmt.Sprintf("%s/workspaces/%d/projects/%d/tasks", baseURL, workspaceID, projectID)
	data, err := c.do(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var tasks []Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("decode tasks: %w", err)
	}
	return tasks, nil
}

func (c *Client) CreateTimeEntry(workspaceID int, entry TimeEntry) error {
	url := fmt.Sprintf("%s/workspaces/%d/time_entries", baseURL, workspaceID)
	_, err := c.do(http.MethodPost, url, entry)
	return err
}
