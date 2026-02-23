# Regenerate: make docs
Output docs/commands/tapes/status.gif

Set Shell "bash"
Set FontFamily "Menlo"
Set FontSize 16
Set Width 900
Set Height 500
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 60ms
Set PlaybackSpeed 1
Set LetterSpacing 0
Set LineHeight 1.2

Hide

Type `export FLOW_HOME=/tmp/flow-demo/.flow`
Enter
Sleep 200ms

Type `export PATH="$(pwd):$PATH"`
Enter
Sleep 200ms

# Overwrite status spec with demo-friendly checks that work locally
Type `cat > /tmp/flow-demo/.flow/status.yaml << 'SPECEOF'`
Enter
Type `apiVersion: flow/v1`
Enter
Type `kind: Status`
Enter
Type `statuses:`
Enter
Type `  - name: closed`
Enter
Type `    description: Work is merged`
Enter
Type `    check: 'test "$FLOW_REPO_BRANCH" = "main"'`
Enter
Type `  - name: in-review`
Enter
Type `    description: Feature branch active`
Enter
Type `    check: 'test "$FLOW_REPO_BRANCH" != "main"'`
Enter
Type `  - name: in-progress`
Enter
Type `    description: Active development`
Enter
Type `    default: true`
Enter
Type `SPECEOF`
Enter
Sleep 200ms

Type "clear"
Enter
Sleep 500ms

Show

Sleep 500ms

# Show all workspace statuses
Type "flow status"
Sleep 500ms
Enter
Sleep 4s

# Show per-repo breakdown
Type "flow status demo"
Sleep 500ms
Enter
Sleep 4s

Sleep 3s
