# toggl-cron

A CLI tool that automatically creates Toggl time entries on a cron schedule.

## Install

```bash
go install github.com/italiczx/toggl-cron@latest
```

Or clone and build locally:

```bash
git clone https://github.com/italiczx/toggl-cron.git
cd toggl-cron
go build -o toggl-cron .
```

## Quick Start

### 1. Setup

Run the interactive setup wizard. The only thing you need is your [Toggl API token](https://track.toggl.com/profile) (scroll to the bottom of the page).

```bash
toggl-cron setup
```

The wizard will:

- Authenticate with your API token
- Auto-detect your workspace (or let you pick if you have multiple)
- Let you select a project and task by name
- Configure duration, billing, start hour, and cron schedule
- Optionally add multiple schedules
- Save everything to `~/.toggl-cron.yaml`

### 2. Run

Start the scheduler:

```bash
toggl-cron run
```

Or fire all entries immediately (useful for testing):

```bash
toggl-cron run --once
```

### 3. Check status

View your current configuration:

```bash
toggl-cron status
```

## Commands

| Command                 | Description                               |
| ----------------------- | ----------------------------------------- |
| `toggl-cron setup`      | Interactive setup wizard                  |
| `toggl-cron run`        | Start the cron scheduler                  |
| `toggl-cron run --once` | Create all entries immediately and exit   |
| `toggl-cron status`     | Show current config and scheduled entries |

## Config

Configuration is stored in `~/.toggl-cron.yaml`. You can edit it directly if you prefer:

```yaml
api_token: "your_token_here"
workspace_id: 12345
workspace: "My Workspace"
schedules:
  - project: "Client A - Website"
    project_id: 111
    task: "Development"
    task_id: 222
    description: "Daily dev work"
    duration: "8h"
    billable: true
    cron: "0 17 * * 1-5"
    start_hour: 8

  - project: "Internal - Admin"
    project_id: 333
    description: "Admin tasks"
    duration: "1h"
    billable: false
    cron: "0 17 * * 1-5"
    start_hour: 9
```

## Running as a Service

To run in the background:

```bash
nohup toggl-cron run > toggl-cron.log 2>&1 &
```

Or create a systemd service, launchd plist, or use a process manager like `supervisord`.
