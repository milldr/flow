# Regenerate: make demo
Output demo.gif

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

# --- Hidden env setup (demo-setup.sh runs before vhs via Makefile) ---
Hide

Type `export FLOW_HOME=/tmp/flow-demo/.flow`
Enter
Sleep 200ms

Type `export PATH="$(pwd):$PATH"`
Enter
Sleep 200ms

Type `export EDITOR="vim -u NONE"`
Enter
Sleep 200ms

Type "clear"
Enter
Sleep 500ms

Show

# === flow status (pre-existing workspaces) ===

Sleep 500ms

Type "flow status"
Sleep 500ms
Enter
Sleep 4s

# === flow exec ===

Type "flow exec api-refactor -- ls"
Sleep 500ms
Enter
Sleep 3s

# === Clear and start creating a new workspace ===

Type "clear"
Enter
Sleep 500ms

# === flow init ===

Type "flow init demo"
Sleep 500ms
Enter
Sleep 3s

# === flow edit state (add a repo in vim) ===

Type "flow edit state demo"
Sleep 500ms
Enter
Sleep 1500ms

# Vim opens state.yaml. Last line is "    repos: []"
# G=last line, $=end of line, F[=back to [, C=change to EOL (deletes [])
Type@0ms "G$F[C"
Sleep 500ms

# Now in insert mode after "    repos: " — type the repo entry
Enter
Type@30ms "        - url: github.com/acme/web"
Enter
Type@30ms "          branch: feat/dashboard"
Enter
Type@30ms "          path: web"
Sleep 3s

# Exit insert mode, save and quit
Escape
Sleep 300ms
Type ":wq"
Enter
Sleep 1500ms

# === flow render ===

Type "flow render demo"
Sleep 500ms
Enter
Sleep 5s

# === flow list ===

Type "flow list"
Sleep 500ms
Enter
Sleep 4s

Sleep 6s
