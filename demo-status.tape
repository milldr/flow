# Regenerate: make demo
Output demo-status.gif

Set Shell "bash"
Set FontFamily "Menlo"
Set FontSize 16
Set Width 1100
Set Height 650
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 30ms
Set PlaybackSpeed 1
Set LetterSpacing 0
Set LineHeight 1.2

# --- Hidden env setup ---
Hide

Type `export FLOW_HOME=/tmp/flow-demo/.flow`
Enter
Sleep 200ms

Type `export PATH="$(pwd):$PATH"`
Enter
Sleep 200ms

# Swap in local-only status spec (no gh needed for demo)
Type `cp "$FLOW_HOME/status-local.yaml" "$FLOW_HOME/status.yaml"`
Enter
Sleep 200ms

Type@0ms `clear`
Enter
Sleep 500ms

Show

# === See all workstreams at a glance ===

Sleep 500ms

Type "flow status"
Sleep 500ms
Enter
Sleep 5s

Type "flow status api-refactor"
Sleep 500ms
Enter
Sleep 4s

Sleep 4s
