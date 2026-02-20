---
name: updating-docs
description: How to update documentation for this project. Covers updating the root README, creating and updating VHS tape recordings, running gendocs, and the make commands that tie it together.
---

# Updating Docs

Documentation lives in three places: the root `README.md`, `docs/commands/` (auto-generated), and VHS tape recordings (GIF demos).

## Root README

`README.md` follows a fixed structure:

1. Title + badge + one-line description
2. Motivation paragraph
3. Hero GIF (`demo.gif`)
4. Quickstart (install, init, state, render)
5. State file reference
6. How it works (directory layout)
7. Requirements
8. Development
9. Support + License

When adding a new command or changing behavior, update the relevant sections. Keep the quickstart concise — detailed docs belong in `docs/commands/`.

## Command Reference Docs

Auto-generated from cobra command definitions. Never edit `docs/commands/*.md` by hand.

Regenerate with:

```bash
make gendocs
```

This runs `cmd/gendocs/main.go`, which:
1. Calls `cobra/doc.GenMarkdownTree` to produce one `.md` per command
2. Injects GIF demos where a matching `tapes/<name>.gif` exists

## VHS Tape Recordings

VHS tapes produce GIF demos from terminal recordings. There are two types:

### Hero Demo (`demo.tape`)

Full walkthrough shown in the root README. Regenerate with:

```bash
make demo
```

### Per-Command Tapes (`docs/commands/tapes/*.tape`)

One tape per command, embedded in generated docs. Regenerate all with:

```bash
make vhs
```

### Tape Template

All tapes share these standard settings:

```tape
Output <output-path>.gif

Set Shell "bash"
Set FontFamily "Menlo"
Set FontSize 16
Set Width 700          # 900 for hero demo
Set Height 300         # 500 for hero demo
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 60ms
Set PlaybackSpeed 1
Set LetterSpacing 0
Set LineHeight 1.2
```

### Hidden Setup Block

Every tape starts with a hidden block that configures the environment:

```tape
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
```

This assumes `demo-setup.sh` has already run (the Makefile handles this). The script creates a temp git repo at `/tmp/demo/app` with a `feature/ipv6` branch.

### Key VHS Commands

- `Type "text"` — types text at `TypingSpeed`
- `Type@0ms "text"` — instant typing (for vim commands)
- `Enter` — press enter
- `Sleep 500ms` — pause
- `Hide` / `Show` — toggle recording visibility
- `Escape` — press escape key

### Adding a New Command Tape

1. Create `docs/commands/tapes/<command>.tape`
2. Use the template above with `Width 700`, `Height 300`
3. Output to `docs/commands/tapes/<command>.gif`
4. Add the `vhs` call to the Makefile `vhs` target
5. Run `make docs` to regenerate everything

## Make Commands

| Command | What it does |
|---------|-------------|
| `make gendocs` | Regenerates `docs/commands/*.md` from cobra definitions |
| `make vhs` | Runs `demo-setup.sh` then all per-command tapes |
| `make demo` | Runs `demo-setup.sh` then the hero `demo.tape` |
| `make docs` | Runs `gendocs` then `vhs` (full doc rebuild) |

### Full Docs Rebuild

After any change to commands, flags, or demos:

```bash
make docs
```

This regenerates all markdown docs and re-records all per-command GIFs.
