# Regenerate: make demo
Output demo.gif

Set Shell "bash"
Set FontFamily "Menlo"
Set FontSize 16
Set Width 900
Set Height 450
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 60ms
Set PlaybackSpeed 2
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

Type "clear"
Enter
Sleep 500ms

Show

# === flow list ===

Sleep 500ms

Type "flow list"
Sleep 500ms
Enter
Sleep 2s

# === flow render (by name) ===

Type "flow render vpc-ipv6"
Sleep 500ms
Enter
Sleep 4s

# === flow exec (by name) ===

Type "flow exec vpc-ipv6 -- ls"
Sleep 500ms
Enter
Sleep 3s

Sleep 5s
