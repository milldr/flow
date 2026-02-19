# Regenerate: make demo
Output demo.gif

Set Shell "bash"
Set Theme "Catppuccin Mocha"
Set FontFamily "JetBrains Mono"
Set FontSize 14
Set Width 1000
Set Height 550
Set Padding 20
Set TypingSpeed 40ms

# --- Hidden setup ---
Hide

Type `bash demo-setup.sh`
Enter
Sleep 2s

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

# === flow init ===

Sleep 300ms

Type "flow init"
Sleep 400ms
Enter
Sleep 1500ms

# Form: "What are you working on?"
Type "VPC IPv6"
Sleep 400ms
Enter
Sleep 1500ms

# Form: "Workspace name" (pre-filled with "vpc-ipv6")
# Accept the default
Enter
Sleep 1500ms

# Form: "Add a repo (blank to finish)" — first repo
Type "/tmp/flow-demo/repos/vpc-service"
Sleep 300ms
Enter
Sleep 1500ms

# Form: "Branch for /tmp/flow-demo/repos/vpc-service?"
Type "feature/ipv6"
Sleep 300ms
Enter
Sleep 1500ms

# Form: "Add a repo (blank to finish)" — second repo
Type "/tmp/flow-demo/repos/subnet-manager"
Sleep 300ms
Enter
Sleep 1500ms

# Form: "Branch for /tmp/flow-demo/repos/subnet-manager?"
Type "feature/ipv6"
Sleep 300ms
Enter
Sleep 1500ms

# Form: "Add a repo (blank to finish)" — done
Enter
Sleep 2500ms

# === flow list ===

Type "flow list"
Sleep 400ms
Enter
Sleep 2s

# === flow render ===

Type "flow render vpc-ipv6"
Sleep 400ms
Enter
Sleep 4s

# === flow exec ===

Type "flow exec vpc-ipv6 -- ls"
Sleep 400ms
Enter
Sleep 3s

Sleep 2s
