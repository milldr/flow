# Regenerate: make docs
Output docs/commands/tapes/state.gif

Set Shell "bash"
Set FontFamily "Menlo"
Set FontSize 16
Set Width 800
Set Height 400
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

Type "flow state demo"
Sleep 500ms
Enter
Sleep 1500ms

# Vim opens state.yaml. Navigate to repos line and add a repo.
Type@0ms "G$F[C"
Sleep 500ms

Enter
Type@30ms "        - url: /tmp/demo/app"
Enter
Type@30ms "          branch: feature/ipv6"
Enter
Type@30ms "          path: app"
Sleep 2s

Escape
Sleep 300ms
Type ":wq"
Enter
Sleep 1500ms

Sleep 2s
