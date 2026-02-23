#!/bin/bash
# Creates a temporary demo environment for the VHS tape.
# Called from Makefile: `make demo` runs this, then `vhs demo.tape`.
set -e

rm -rf /tmp/flow-demo /tmp/demo
mkdir -p /tmp/demo

# Create a local git repo with a feature branch
dir="/tmp/demo/app"
mkdir -p "$dir"
git -C "$dir" init -q
echo "# app" > "$dir/README.md"
echo "package main" > "$dir/main.go"
git -C "$dir" add .
git -C "$dir" commit -q -m "initial commit"
git -C "$dir" checkout -q -b feature/ipv6
echo "// IPv6 support" >> "$dir/main.go"
git -C "$dir" add .
git -C "$dir" commit -q -m "add ipv6 support"

# Pre-populate the bare clone cache so tapes can use a realistic URL.
# Flow looks up repos by URL → ~/.flow/repos/<url>.git
# The bare clone's origin still points at the local repo, so fetches work.
bare="/tmp/flow-demo/.flow/repos/github.com/acme/app.git"
mkdir -p "$(dirname "$bare")"
git clone --bare "$dir" "$bare" -q
