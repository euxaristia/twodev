# Creates stacked Go rewrite PRs. Run from repo root.
$ErrorActionPreference = "Stop"

$stack = @(
    @{ Branch = "go/01-foundation"; Base = "main"; Title = "feat(go): twodev foundation"; Body = "Ports bootstrap config loading and boots a health-checked HTTP server."; Paths = @("go.mod", "go.sum", "internal/version", "internal/config", "internal/server", "cmd/twodev") }
    @{ Branch = "go/02-buildspec"; Base = "go/01-foundation"; Title = "feat(go): buildspec parser"; Body = "Ports buildspec YAML parsing for .onedev-buildspec.yml."; Paths = @("internal/buildspec") }
    @{ Branch = "go/03-agent-protocol"; Base = "go/02-buildspec"; Title = "feat(go): agent websocket protocol"; Body = "Ports agent Message framing and adds twodev-agent client skeleton."; Paths = @("internal/agent/protocol", "internal/agent/client", "cmd/twodev-agent") }
    @{ Branch = "go/04-job-executor"; Base = "go/03-agent-protocol"; Title = "feat(go): local job executor"; Body = "Ports local CommandStep execution."; Paths = @("internal/job") }
    @{ Branch = "go/05-git-core"; Base = "go/04-job-executor"; Title = "feat(go): git service"; Body = "Ports git clone and checkout via git CLI."; Paths = @("internal/git") }
    @{ Branch = "go/06-store"; Base = "go/05-git-core"; Title = "feat(go): sqlite store"; Body = "Replaces Hibernate with embedded SQLite schema."; Paths = @("internal/store") }
    @{ Branch = "go/07-domain-models"; Base = "go/06-store"; Title = "feat(go): domain models"; Body = "Ports core model structs."; Paths = @("internal/model") }
    @{ Branch = "go/08-auth"; Base = "go/07-domain-models"; Title = "feat(go): bearer auth"; Body = "Ports agent bearer token validation."; Paths = @("internal/auth") }
    @{ Branch = "go/09-rest-api"; Base = "go/08-auth"; Title = "feat(go): REST API"; Body = "Adds JSON API under /~api/twodev/."; Paths = @("internal/api", "internal/server", "cmd/twodev") }
    @{ Branch = "go/10-scheduler"; Base = "go/09-rest-api"; Title = "feat(go): job scheduler"; Body = "In-memory job queue skeleton."; Paths = @("internal/scheduler") }
    @{ Branch = "go/11-agent-server"; Base = "go/10-scheduler"; Title = "feat(go): agent websocket server"; Body = "Websocket endpoint at /~server."; Paths = @("internal/agent/server", "internal/server", "cmd/twodev") }
    @{ Branch = "go/12-githttp"; Base = "go/11-agent-server"; Title = "feat(go): git smart HTTP"; Body = "Git smart HTTP route skeleton."; Paths = @("internal/githttp", "internal/server", "cmd/twodev") }
    @{ Branch = "go/13-sshserver"; Base = "go/12-githttp"; Title = "feat(go): embedded SSH server"; Body = "SSH listener skeleton for git over SSH."; Paths = @("internal/sshserver", "cmd/twodev") }
    @{ Branch = "go/14-search"; Base = "go/13-sshserver"; Title = "feat(go): bleve search index"; Body = "Bleve full-text search index."; Paths = @("internal/search") }
    @{ Branch = "go/15-issues-prs"; Base = "go/14-search"; Title = "feat(go): issues and pull requests"; Body = "Issue and pull request services."; Paths = @("internal/issue", "internal/pullrequest") }
    @{ Branch = "go/16-platform"; Base = "go/15-issues-prs"; Title = "feat(go): platform services"; Body = "Notification, pack registry, cluster, event bus, workspace, buildspec triggers."; Paths = @("internal/notification", "internal/pack", "internal/cluster", "internal/event", "internal/workspace", "internal/buildspec", "internal/server", "cmd/twodev", "go.mod", "go.sum") }
)

git checkout main
git branch -D go/wall 2>$null
git checkout -b go/wall
git add go.mod go.sum cmd internal
git commit -m "feat(go): cumulative twodev rewrite snapshot"

$parent = "main"
foreach ($item in $stack) {
    git checkout -B $item.Branch $parent
    foreach ($path in $item.Paths) {
        git checkout go/wall -- $path
    }
    git add -A
    git commit -m $item.Title
    git push -u origin $item.Branch --force
    $parent = $item.Branch
}

git checkout main
git branch -D go/wall
Write-Host "Branches pushed. Creating PRs..."