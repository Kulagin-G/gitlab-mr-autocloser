# gitlab-mr-autocloser
Simple Go application for closing stale Gitlab Merge Requests via Gitlab API and internal cron scheduler.

# Logic
There is a common block scheme describing application logic:
![Screenshot](docs/mr-closer.jpg)

# Local run without compiling
```bash
cd ./gitlab-mr-autocloser/src
go mod tidy
GITLAB_API_TOKEN=<your_token>;go run main.go -config ../config/config.yaml
```
For compiling options see Dockerfile.

# Tests
There are several unit tests were implemented:
```bash
cd ./gitlab-mr-autocloser
go test ./src/...
```
# Config description
```yaml
---
# Cron expressions.
cronSchedule: "*/1 * * * *"
# The head of MR label: "staleMR::closeAfterDays::<closeMRAfterDays>"
labelHead: "close_if_no_updates_days::"
# Gitlab Base API URL
gitlabBaseApiUrl: "https://gitlab.com/api/v4"

# http/net healthcheck options
healthcheckOptions:
  # Application binding host.
  host: 127.0.0.1
  # Application port for exposing liveness and readiness probes.
  port: 8090
  liveness:
    path: "/healthz/live"
    # goroutineHealthcheckHandler check fails if too many goroutines are running.
    gorMaxNum: 100
  readiness:
    path: "/healthz/ready"
    # dnsHealthcheckHandler returns a Check that makes sure the provided host can resolve to at least one IP address within the specified timeout.
    resolveTimeoutSec: 5
    urlCheck: "gitlab.com"

defaultOptions:
  # All opened MRs will be labeled as stale after staleMRAfterDays since CREATION DATE TIME.
  staleMRAfterDays: 14
  # Close stale opened MRs after closeMRAfterDays since LAST UPDATED TIME.
  closeMRAfterDays: 7

projects:
  - name: "my_group/my_sybgroup/my_project"
    # you can override defaults
#    overrideOptions:
#      staleMRAfterDays: 1
#      closeMRAfterDays: 5

```