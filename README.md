# toggl-cron

A Go tool that automatically creates Toggl time entries on a schedule.

## Setup

1. Clone the repo
2. Copy the example env file and fill in your values:

```bash
cp .env.example .env
```

3. Add your `TOGGL_API_TOKEN` — find it at the bottom of your [Toggl profile page](https://track.toggl.com/profile).

4. Add your `WORKSPACE_ID` — grab it from your workspace settings URL:
   `https://track.toggl.com/{org_id}/workspaces/{workspace_id}/settings/activity`

5. Install dependencies:

```bash
go mod tidy
```

## Finding your Project ID

Before running the cron, you need your `PROJECT_ID`. List all projects in your workspace:

```bash
go run ./cmd/projects
```

This prints a table of projects with their ID, name, and status. Copy the ID of the project you want and add it to your `.env` as `PROJECT_ID`.

## Finding your Task ID

Once `PROJECT_ID` is set in your `.env`, list the tasks for that project:

```bash
go run ./cmd/tasks
```

This prints a table of tasks with their ID, name, and active status. Copy the ID of the task you want and add it to your `.env` as `TASK_ID`.

## Configure your time entry

Set the remaining values in your `.env`:

| Variable      | Description                   | Default                     |
| ------------- | ----------------------------- | --------------------------- |
| `DESCRIPTION` | Title of the time entry       | `Auto-logged by toggl-cron` |
| `DURATION`    | Duration in seconds           | `28800` (8 hours)           |
| `BILLABLE`    | Whether the entry is billable | `false`                     |

## Running the cron

Once your `.env` is fully configured:

```bash
go run .
```

This starts a long-running process that creates a time entry every day at **5pm local time**. It will keep running until you stop it with `Ctrl+C`.

To run it in the background:

```bash
nohup go run . > toggl-cron.log 2>&1 &
```
