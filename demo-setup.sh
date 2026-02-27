#!/bin/bash
# Creates a temporary demo environment for the VHS tape.
# Called from Makefile: `make demo` runs this, then `vhs demo.tape`.
set -e

export FLOW_HOME="/tmp/flow-demo/.flow"
FLOW="$(pwd)/flow"
rm -rf /tmp/flow-demo /tmp/demo
mkdir -p /tmp/demo "$FLOW_HOME/workspaces" "$FLOW_HOME/repos"

# --- Create a fake claude binary that simulates an interactive agent session ---

mkdir -p /tmp/flow-demo/bin
cat > /tmp/flow-demo/bin/claude <<'SCRIPT'
#!/bin/bash
DIM='\033[2m'
CYAN='\033[36m'
GREEN='\033[32m'
BOLD='\033[1m'
RESET='\033[0m'

echo ""
echo -e "  ${DIM}[mocked agent session]${RESET}"
echo ""
echo -e "  ${DIM}Flow includes skills that teach your agent to manage workspaces.${RESET}"
echo -e "  ${DIM}Add your own skills for repo discovery, PR lookup, or any custom workflow.${RESET}"
echo -e "  ${DIM}Paste a Slack thread, dictate a bug report — the agent handles the rest.${RESET}"
echo ""
echo -e "  ${BOLD}Enter your prompt:${RESET}"
echo ""
printf "  > "
read -r task

echo ""
sleep 0.5
echo -e "  ${CYAN}● Reading skill: flow-cli${RESET}"
sleep 0.6
echo -e "  ${CYAN}● Reading skill: find-repo${RESET}"
echo -e "      ${DIM}→ github.com/acme/web${RESET}"
sleep 0.6
echo -e "  ${CYAN}● Editing state.yaml${RESET}"
echo -e "      ${DIM}name: fix-dashboard-charts${RESET}"
echo -e "      ${DIM}repo: github.com/acme/web @ fix/dashboard-charts${RESET}"
sleep 0.6
echo -e "  ${CYAN}● Running flow render...${RESET}"
sleep 0.8
echo ""
echo -e "  ${GREEN}✓ Workspace ready${RESET}"
echo ""
echo -e "  ${CYAN}● Analyzing web/src/components/Charts.tsx...${RESET}"
sleep 0.8
echo ""
printf "  > "
read -r cmd
echo ""
SCRIPT
chmod +x /tmp/flow-demo/bin/claude

# --- Create local git repos with feature branches ---

create_repo() {
  local name="$1" branch="$2" file="$3" msg="$4"
  local dir="/tmp/demo/$name"
  mkdir -p "$dir"
  git -C "$dir" init -q
  echo "# $name" > "$dir/README.md"
  echo "package main" > "$dir/main.go"
  git -C "$dir" add .
  git -C "$dir" commit -q -m "initial commit"
  git -C "$dir" checkout -q -b "$branch"
  echo "// $msg" >> "$dir/$file"
  git -C "$dir" add .
  git -C "$dir" commit -q -m "$msg"
}

create_repo "app"     "feature/ipv6"     "main.go"   "add ipv6 support"
create_repo "api"     "feat/auth"        "main.go"   "add auth endpoints"
create_repo "docs"    "update/guides"    "README.md" "update setup guide"
create_repo "web"     "feat/dashboard"   "main.go"   "add dashboard page"
create_repo "billing" "feat/billing-v2"  "main.go"   "billing v2 migration"
create_repo "gateway" "feat/rate-limits" "main.go"   "add rate limiting"
create_repo "config"  "feat/env-vars"    "main.go"   "add env var support"

# --- Pre-populate bare clone cache for realistic URLs ---

for name in app api docs web billing gateway config; do
  bare="$FLOW_HOME/repos/github.com/acme/${name}.git"
  mkdir -p "$(dirname "$bare")"
  git clone --bare "/tmp/demo/$name" "$bare" -q
  # Add fetch refspec so worktrees can resolve origin/* refs
  git -C "$bare" config remote.origin.fetch "+refs/heads/*:refs/remotes/origin/*"
  git -C "$bare" fetch -q origin
done

# --- Create and render pre-existing workspaces ---

create_workspace() {
  local id="$1" yaml="$2"
  mkdir -p "$FLOW_HOME/workspaces/$id"
  echo "$yaml" > "$FLOW_HOME/workspaces/$id/state.yaml"
}

# Workspace: api-refactor — will be "in-progress" (has local diffs)
create_workspace "bold-creek" "$(cat <<'YAML'
apiVersion: flow/v1
kind: State
metadata:
  name: api-refactor
  description: Refactor authentication API
  created: "2026-02-21T14:30:00Z"
spec:
  repos:
    - url: github.com/acme/api
      branch: feat/auth
      path: api
    - url: github.com/acme/docs
      branch: update/guides
      path: docs
YAML
)"

# Workspace: infra-update — will be "in-review" (marker file)
create_workspace "swift-pine" "$(cat <<'YAML'
apiVersion: flow/v1
kind: State
metadata:
  name: infra-update
  description: Update infrastructure across services
  created: "2026-02-19T09:00:00Z"
spec:
  repos:
    - url: github.com/acme/app
      branch: feature/ipv6
      path: app
YAML
)"

# Workspace: billing-v2 — will be "closed" (marker file)
create_workspace "calm-ridge" "$(cat <<'YAML'
apiVersion: flow/v1
kind: State
metadata:
  name: billing-v2
  description: Migrate billing system to v2
  created: "2026-02-15T10:00:00Z"
spec:
  repos:
    - url: github.com/acme/billing
      branch: feat/billing-v2
      path: billing
YAML
)"

# Workspace: rate-limits — will be "in-progress" (has local diffs)
create_workspace "dry-fog" "$(cat <<'YAML'
apiVersion: flow/v1
kind: State
metadata:
  name: rate-limits
  description: Add rate limiting to API gateway
  created: "2026-02-23T16:00:00Z"
spec:
  repos:
    - url: github.com/acme/gateway
      branch: feat/rate-limits
      path: gateway
YAML
)"

# Workspace: env-config — will be "open" (no changes)
create_workspace "iron-vale" "$(cat <<'YAML'
apiVersion: flow/v1
kind: State
metadata:
  name: env-config
  description: Add environment variable support
  created: "2026-02-26T08:00:00Z"
spec:
  repos:
    - url: github.com/acme/config
      branch: feat/env-vars
      path: config
YAML
)"

# Render all workspaces
$FLOW render bold-creek
$FLOW render swift-pine
$FLOW render calm-ridge
$FLOW render dry-fog
$FLOW render iron-vale

# --- Set up different statuses via local changes and marker files ---

# api-refactor: "in-progress" — add local commits so git diff detects changes
add_local_change() {
  local ws_dir="$FLOW_HOME/workspaces/$1/$2"
  git -C "$ws_dir" fetch -q origin 2>/dev/null || true
  echo "# local change" >> "$ws_dir/README.md"
  git -C "$ws_dir" add .
  git -C "$ws_dir" commit -q -m "wip: local changes"
}

add_local_change "bold-creek" "api"
add_local_change "bold-creek" "docs"

# rate-limits: "in-progress" — local commits
add_local_change "dry-fog" "gateway"

# infra-update: "in-review" — marker file
touch "$FLOW_HOME/workspaces/swift-pine/app/.flow-review"

# billing-v2: "closed" — marker file
touch "$FLOW_HOME/workspaces/calm-ridge/billing/.flow-closed"

# --- Create status specs ---
# The default spec (written by EnsureDirs on first flow command) uses gh + jq.
# For the demo, we keep the default for the edit-status tape (so it shows the
# real commands), then the status tape swaps in a local-only spec before running.

# Write a local-only spec that the status tape will swap in (no gh needed).
# Checks are evaluated in order — first match wins per repo.
cat > "$FLOW_HOME/status-local.yaml" <<YAML
apiVersion: flow/v1
kind: Status
spec:
  statuses:
    - name: closed
      description: All PRs merged or closed
      color: "131"
      check: test -f "\$FLOW_REPO_PATH/.flow-closed"
    - name: stale
      description: Workspace inactive
      color: magenta
      check: 'false'
    - name: in-review
      description: Non-draft PR open
      color: purple
      check: test -f "\$FLOW_REPO_PATH/.flow-review"
    - name: in-progress
      description: Local diffs or draft PR
      color: yellow
      check: git -C "\$FLOW_REPO_PATH" diff --name-only "origin/\$FLOW_REPO_BRANCH" 2>/dev/null | grep -q .
    - name: open
      description: Workspace created, no changes yet
      color: green
      default: true
YAML
