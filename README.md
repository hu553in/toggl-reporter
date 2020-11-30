# Toggl reporter

## Description

This project is an app for creating custom daily reports
based on data from Toggl.

It is inspired by
[senior-sigan/toggl-reporter](https://github.com/senior-sigan/toggl-reporter),
but rewritten in Go for more customization and learning purposes.

## How to run

1. Install `Go` (≥ 1.15)
2. Run the following sequence of commands:
    ```
    go get github.com/hu553in/toggl-reporter
    go install github.com/hu553in/toggl-reporter
    sudo mv $(go env GOPATH)/bin/toggl-reporter /usr/local/bin/
    ```
3. Run `toggl-reporter` with `--help` or some another CLI args

## CLI args

* `-token` - a Toggl API token (you can get it from your profile page)
* `-workspaceId` - a workspace ID (you can get it by running the app
with `-printWorkspaces` flag)
* `-date` - a report date (can be: `today`, `yesterday`, `YYYY-MM-DD`)
(default `today`)
* `-doNotMergeEqual` - do not merge tasks with equal descriptions
* `-printWorkspaces` - print workspaces instead of the report
