# BMAD — mainichi repo understanding brief

## Mission
Understand **mainichi** well enough to onboard a new contributor quickly without making the repo feel bigger than it is.

## Objective
Give a grounded map of:
- what the app does
- how the code is split
- how user data flows
- where the important behavior lives
- where a safe first change would land

## Constraints
- Keep the app **single-purpose**: daily writing.
- Keep the UX **quiet** and **low-friction**.
- Favor **plaintext-first** storage.
- Avoid turning the app into a tracker / analytics tool.
- Prefer deterministic behavior where possible.

## Non-goals
- No streak system
- No dashboarding
- No social / collaboration features
- No cloud sync layer
- No metrics-heavy productization

## One-line summary
**mainichi is a terminal writing ritual app**: it opens a dated entry, optionally supplies a prompt, and stores the result as readable markdown on disk.

## Repo map

### Entry point
- `cmd/mainichi/main.go`
  - parses CLI commands
  - loads config and storage
  - selects the initial UI view
  - wires the Bubble Tea app together

### Application layer
- `app/session.go`
  - opens today's entry or a specific date
  - saves entries
  - draws prompts from the deck
  - loads recent entries and calendar data

### Storage / external adapters
- `adapters/fs.go`
  - filesystem-backed store
  - reads/writes entries, config, prompts, and prompt deck state
- `adapters/openai.go`
  - optional AI prompt generation

### Domain logic
- `core/entry.go`
  - entry parsing / serialization
  - word counting
- `core/prompt.go`
  - prompt deck shuffle / draw logic
- `core/config.go`
  - config defaults
- `core/stoic.go`
  - maps dates to stoic headings

### UI layer
- `ui/app.go`
  - top-level Bubble Tea model and view switching
- `ui/writer.go`
  - writing view
- `ui/calendar.go`
  - calendar view
- `ui/recent.go`
  - recent entries view
- `ui/config.go`
  - config editor view

### Assets / embedded data
- `prompts_embed.go`
- `stoic_embed.go`
- `daily_stoic_headings.json`

### Planning notes
- `plan.md`
  - captures the product philosophy and intentional non-features

## Data model

The default data directory is:

```text
~/.mainichi/
```

Expected contents:

```text
~/.mainichi/
  entries/
    YYYY-MM-DD.md
  prompts.txt
  prompt_state.json
  config.toml
```

### Entries
Each day is stored as markdown with frontmatter:

```md
---
date: 2026-02-12
prompt: "Where are you mistaking motion for progress?"
minimum: 300
---

Body text here.
```

Important detail: the body is plain markdown text, but the app does not try to be fancy about formatting.

### Config
`config.toml` currently supports:
- `minimum`
- `auto_prompt`
- `prompt_source`

### Prompt deck state
`prompt_state.json` tracks the shuffled prompt deck so prompts are drawn in a finite cycle.

## Primary user flows

### 1) Open the app
- `main.go` loads the store and config.
- `Session.OpenToday()` creates or loads today’s entry.
- The writer view becomes the default screen.

### 2) Open with a prompt
- `mainichi prompt` draws from the default prompt deck.
- If `auto_prompt` is on, the startup flow may also assign a prompt.
- If `prompt_source = "ai"`, the app can fetch an AI-generated prompt when `OPENAI_API_KEY` is present.
- If `prompt_source = "stoic"`, it uses the embedded Daily Stoic heading for the date.

### 3) Write and save
- entry text lives in the current session
- save writes a markdown file into `entries/`
- minimum word count is stored alongside the entry

### 4) Browse the archive
- `date` opens a calendar view
- `recent` lists recent entries
- a date argument like `2025-03-15` opens that day directly

### 5) Configure behavior
- `config` opens the config view
- user can change minimums, auto-prompt behavior, prompt source, or reset deck state

## Core behavior worth knowing

### Prompt deck
- prompts are loaded from `prompts.txt`
- the deck is shuffled
- each prompt is used once before reshuffling
- the deck state is persisted separately from entries

### Frontmatter parsing
`core.ParseEntry()` expects YAML-style frontmatter bounded by `---` lines.
If the frontmatter is malformed, loading fails.

### Word counting
`core.WordCount()` is just whitespace splitting.
That keeps the behavior simple and predictable.

### Calendar / recent views
Those views are derived from entries on disk, not from a separate database.
That is a strong design choice: the filesystem is the source of truth.

## Architecture shape

This repo is split in a way that is easy to explain to new contributors:

- **core** = pure-ish domain logic
- **app** = use cases / orchestration
- **adapters** = filesystem + optional external APIs
- **ui** = terminal presentation and interaction
- **cmd** = startup wiring

That means most meaningful changes should probably happen in `core/` or `app/`, not in the UI first.

## What makes the app feel “mainichi”

The philosophy is visible in `plan.md` and the implementation:
- calm terminal UI
- daily ritual instead of task management
- prompt as a nudge, not a goal
- plain files that stay readable outside the app
- no analytics theater

## Good first-change areas

If you want to improve the repo without disturbing the core idea, safe places include:
- better onboarding text / help copy
- minor writer UI polish
- prompt source handling
- config ergonomics
- entry loading / validation hardening
- calendar/recent view readability

## Things to watch for

- `tmp/` is not part of the product surface; ignore it.
- The repo should stay easy to inspect with normal file tools.
- Any feature that introduces hidden state should be justified carefully.

## Short handoff for another agent

> mainichi is a Go + Bubble Tea daily writing app. Focus on `cmd/mainichi/main.go`, `app/session.go`, `adapters/fs.go`, `core/entry.go`, `core/prompt.go`, and the UI models. The app stores each day as markdown in `~/.mainichi/entries/YYYY-MM-DD.md`, uses a shuffled prompt deck, and intentionally avoids streaks/analytics. A useful next step is to improve understanding/onboarding without changing the app’s calm, plaintext-first feel.
