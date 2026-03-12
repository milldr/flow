#!/bin/bash
# Creates a temporary demo environment for the VHS tape.
# Called from Makefile: `make demo` runs this, then `vhs demo.tape`.
set -e

export FLOW_HOME="/tmp/flow-demo/.flow"
FLOW="$(pwd)/flow"
rm -rf /tmp/flow-demo /tmp/demo
mkdir -p /tmp/demo "$FLOW_HOME/workspaces" "$FLOW_HOME/repos"

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

create_repo "app" "feature/ipv6"   "main.go"   "add ipv6 support"
create_repo "api" "feat/auth"      "main.go"   "add auth endpoints"
create_repo "docs" "update/guides" "README.md" "update setup guide"
create_repo "web" "feat/dashboard" "main.go"   "add dashboard page"

# --- Pre-populate bare clone cache for realistic URLs ---

for name in app api docs web; do
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

# Render both workspaces so exec and status work
# Use --reset=false to skip interactive prompt when branches already exist
$FLOW render bold-creek --reset=false
$FLOW render swift-pine --reset=false

# --- Add local commits so status checks detect diffs ---

add_local_change() {
  local ws_dir="$FLOW_HOME/workspaces/$1/$2"
  git -C "$ws_dir" fetch -q origin 2>/dev/null || true
  echo "# local change" >> "$ws_dir/README.md"
  git -C "$ws_dir" add .
  git -C "$ws_dir" commit -q -m "wip: local changes"
}

add_local_change "bold-creek" "api"
add_local_change "bold-creek" "docs"
add_local_change "swift-pine" "app"

# --- Create status specs ---
# The default spec (written by EnsureDirs on first flow command) uses gh + jq.
# For the demo, we keep the default for the edit-status tape (so it shows the
# real commands), then the status tape swaps in a local-only spec before running.

# Write a local-only spec that the status tape will swap in (no gh needed).
cat > "$FLOW_HOME/status-local.yaml" <<YAML
apiVersion: flow/v1
kind: Status
spec:
  statuses:
    - name: closed
      description: All PRs merged or closed
      check: 'false'
    - name: in-review
      description: Non-draft PR open
      check: 'false'
    - name: in-progress
      description: Local diffs or draft PR
      check: git -C "\$FLOW_REPO_PATH" diff --name-only "origin/\$FLOW_REPO_BRANCH" 2>/dev/null | grep -q .
    - name: open
      description: Workspace created, no changes yet
      default: true
YAML
