# Regenerate: make demo
Output demo.gif

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

# --- Hidden env setup (demo-setup.sh runs before vhs via Makefile) ---
Hide

Type `export FLOW_HOME=/tmp/flow-demo/.flow`
Enter
Sleep 200ms

Type `export PATH="$(pwd):$PATH"`
Enter
Sleep 200ms

Type@0ms `clear`
Enter
Sleep 500ms

Show

# === flow init — launches Claude automatically ===

Sleep 500ms

Type "flow init"
Sleep 500ms
Enter

# Wait for Claude to launch and show the workspace trust dialog
Sleep 3s

# Accept the workspace trust dialog
Hide
Enter
Show
Sleep 3s

# Type the task into Claude
Type "Our auth migration broke prod — API errors in github.com/acme/apps, wrong docs in github.com/acme/docs, and infra config drift in github.com/acme/infrastructure. Set up a workspace for this, then ask me what to look at first."
Sleep 500ms
Enter

# Wait for Claude to read skills, edit state.yaml, run flow render
Sleep 40s
