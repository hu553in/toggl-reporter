# Toggl reporter

## Description

This project is an app for creating custom daily reports based on the data from Toggl.\
It is inspired by [senior-sigan/toggl-reporter](https://github.com/senior-sigan/toggl-reporter), but rewritten in Go 
for more customization and learning purposes.

## How to run

1. Install Go (≥ 1.15)
2. Run `rm -rf $(go env GOPATH)/**/github.com/hu553in/toggl-reporter && rm -rf /usr/local/bin/toggl-reporter` in case 
if you want to reinstall the app
3. Run the following sequence of commands:
    ```
    go get github.com/hu553in/toggl-reporter
    go install github.com/hu553in/toggl-reporter
    mv $(go env GOPATH)/bin/toggl-reporter /usr/local/bin/
    ```
4. Run `toggl-reporter` with `--help` or some another CLI args

## CLI args

* `-date` — a report date (can be: `today`, `yesterday` or `YYYY-MM-DD`, default — `today`)
* `-doNotMergeEqual` — do not merge tasks with equal descriptions
* `-printWorkspaces` — print workspaces instead of the report
* `-showDurationForEach` — show duration for each task
* `-token` — a Toggl API token (you can get it from your Toggl profile page)
* `-workspaceId` — a workspace ID (you can get the list of IDs by running the app with `-printWorkspaces` flag or just 
with `-token`)
