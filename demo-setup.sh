#!/bin/bash
# Creates a temporary demo environment for the VHS tape.
# Called from demo.tape's hidden section.
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
