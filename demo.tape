# Regenerate: make demo
Output demo.gif

Set Shell "bash"
Set FontFamily "Menlo"
Set FontSize 16
Set Width 900
Set Height 550
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 30ms
Set PlaybackSpeed 1
Set LetterSpacing 0
Set LineHeight 1.2

# --- Hidden env setup (demo-setup.sh runs before vhs via Makefile) ---
Hide

Type `export FLOW_HOME=/tmp/flow-demo/.flow`
Enter
Sleep 200ms

Type `export PATH="/tmp/flow-demo/bin:$(pwd):$PATH"`
Enter
Sleep 200ms

# Swap in local-only status spec (no gh needed for demo)
Type `cp "$FLOW_HOME/status-local.yaml" "$FLOW_HOME/status.yaml"`
Enter
Sleep 200ms

# Clear screen and print first section comment (hidden from recording)
Type@0ms `clear && printf '\033[2;3m# create a workspace — agent launches automatically\033[0m\n'`
Enter
Sleep 500ms

Show

# === Create a workspace — agent launches automatically ===

Sleep 500ms

Type "flow init dashboard"
Sleep 500ms
Enter
Sleep 3s

# The fake claude is now running and showing its prompt.
# Type the task description into the agent.
Type "dashboard charts broke after the auth migration, check the web frontend"
Sleep 500ms
Enter
Sleep 5s

# Agent has finished setting up. Exit the session.
Type "/exit"
Sleep 500ms
Enter
Sleep 2s

# === See all workstreams at a glance ===

# Clear screen and print second section comment (hidden from recording)
Hide
Type@0ms `clear && printf '\033[2;3m# see all workstreams at a glance\033[0m\n'`
Enter
Sleep 500ms
Show

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
