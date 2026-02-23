# Regenerate: make docs
Output docs/commands/tapes/edit_status.gif

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

Type `export EDITOR="vim -u NONE"`
Enter
Sleep 200ms

Type "clear"
Enter
Sleep 500ms

Show

Sleep 500ms

# Creates starter spec and opens in vim
Type "flow edit status"
Sleep 500ms
Enter
Sleep 2s

# Vim opens the status spec. Browse the content.
Sleep 3s

# Quit
Type ":q"
Enter
Sleep 1500ms

Sleep 2s
