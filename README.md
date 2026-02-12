# mainichi

A minimal daily writing habit tool for the terminal.

## Install

```
go install mainichi/cmd/mainichi@latest
```

## Usage

```
mainichi                # open today's entry
mainichi prompt         # open today with a writing prompt
mainichi ai             # open today with an AI-generated writing prompt
mainichi config         # settings
mainichi date           # calendar view
mainichi recent         # browse recent entries
mainichi 2025-03-15     # open a specific date
```

## Keys

| Key | Action |
|---|---|
| `ctrl+s` | save |
| `ctrl+c` | save and quit |
| `esc` | command palette |

In the command palette, use `j`/`k` or arrows to navigate and `enter` to select. Press `/` to search.

## Environment

| Variable | Description |
|---|---|
| `MAINICHI_DIR` | Custom data directory (default: `~/.mainichi`) |
| `OPENAI_API_KEY` | Required for `mainichi ai` and `prompt_source = "ai"` |

To sync entries via iCloud, add to your `.zshrc`:

```
export MAINICHI_DIR="$HOME/Library/Mobile Documents/com~apple~CloudDocs/mainichi"
```

## Data

Entries are stored in `~/.mainichi/` (or `$MAINICHI_DIR` if set).
