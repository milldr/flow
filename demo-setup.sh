#!/bin/bash
# Creates a temporary demo environment for the VHS tape.
# Called from Makefile: `make demo` runs this, then `vhs demo.tape`.
set -e

export FLOW_HOME="/tmp/flow-demo/.flow"
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

# --- Pre-populate bare clone cache for realistic URLs ---

for name in app api docs; do
  bare="$FLOW_HOME/repos/github.com/acme/${name}.git"
  mkdir -p "$(dirname "$bare")"
  git clone --bare "/tmp/demo/$name" "$bare" -q
done

# --- Create additional workspaces (the "demo" workspace is created by the init tape) ---

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
    - url: github.com/acme/api
      branch: feat/auth
      path: api
    - url: github.com/acme/docs
      branch: update/guides
      path: docs
YAML
)"
