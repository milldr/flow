#!/bin/bash
# Creates a temporary demo environment for the VHS tape.
# Called from Makefile: `make demo` runs this, then `vhs demo.tape`.
set -e

rm -rf /tmp/flow-demo
mkdir -p /tmp/flow-demo

export FLOW_HOME=/tmp/flow-demo/.flow

for name in vpc-service subnet-manager; do
  dir="/tmp/flow-demo/repos/$name"
  mkdir -p "$dir"
  git -C "$dir" init -q
  echo "# $name" > "$dir/README.md"
  echo "package main" > "$dir/main.go"
  git -C "$dir" add .
  git -C "$dir" commit -q -m "initial commit"
  git -C "$dir" checkout -q -b feature/ipv6
  echo "// IPv6 support" >> "$dir/main.go"
  git -C "$dir" add .
  git -C "$dir" commit -q -m "add ipv6 support"
done

# Pre-create a workspace so the demo can show list/render/exec
./flow init --name vpc-ipv6

# Find the generated workspace ID and populate its state file with repos
WS_ID=$(ls "$FLOW_HOME/workspaces/")
cat > "$FLOW_HOME/workspaces/$WS_ID/state.yaml" << 'EOF'
apiVersion: flow/v1
kind: State
metadata:
    name: vpc-ipv6
    created: "2025-01-15T10:00:00Z"
spec:
    repos:
        - url: /tmp/flow-demo/repos/vpc-service
          branch: feature/ipv6
          path: vpc-service
        - url: /tmp/flow-demo/repos/subnet-manager
          branch: feature/ipv6
          path: subnet-manager
EOF
