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

Type `export EDITOR="vim -u NONE"`
Enter
Sleep 200ms

Type "clear"
Enter
Sleep 500ms

Show

# === flow init ===

Sleep 500ms

Type "flow init --name demo"
Sleep 500ms
Enter
Sleep 2s

# === flow state (add a repo in vim) ===

Type "flow state demo"
Sleep 500ms
Enter
Sleep 1500ms

# Vim opens state.yaml. Last line is "    repos: []"
# G=last line, $=end of line, F[=back to [, C=change to EOL (deletes [])
Type@0ms "G$F[C"
Sleep 500ms

# Now in insert mode after "    repos: " — type the repo entry
Enter
Type@30ms "        - url: /tmp/demo/app"
Enter
Type@30ms "          branch: feature/ipv6"
Enter
Type@30ms "          path: app"
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
Sleep 4s

# === flow exec ===

Type "flow exec demo -- ls"
Sleep 500ms
Enter
Sleep 3s

Sleep 8s
