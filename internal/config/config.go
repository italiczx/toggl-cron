package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Schedule struct {
	ProjectName string `yaml:"project"`
	ProjectID   int    `yaml:"project_id"`
	TaskName    string `yaml:"task,omitempty"`
	TaskID      int    `yaml:"task_id,omitempty"`
	Description string `yaml:"description"`
	Duration    string `yaml:"duration"`
	Billable    bool   `yaml:"billable"`
	Cron        string `yaml:"cron"`
	StartHour   int    `yaml:"start_hour"`
}

type Config struct {
	APIToken      string     `yaml:"api_token"`
	WorkspaceID   int        `yaml:"workspace_id"`
	WorkspaceName string     `yaml:"workspace"`
	Schedules     []Schedule `yaml:"schedules"`
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".toggl-cron.yaml")
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func ParseDuration(s string) (int, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", s, err)
	}
	return int(d.Seconds()), nil
}

func FormatDuration(seconds int) string {
	d := time.Duration(seconds) * time.Second
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh%dm", h, m)
}
