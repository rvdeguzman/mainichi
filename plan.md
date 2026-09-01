# mainichi — Plan

## Philosophy

**mainichi** is a daily writing ritual.

It is not:
- A productivity tracker
- A streak counter
- A quantified self tool
- A coaching app
- A mood analytics dashboard

It is:
- A quiet container
- A daily writing practice
- A plaintext-first system
- A long-term archive of thought

Prompts are steering devices, not the goal.  
Metrics are hidden.  
Presence matters more than performance.

---

## Core Principles

1. **One purpose only** — daily writing.
2. **Plaintext first** — human-readable, grep-able, durable.
3. **No visible metrics** — no word counts, no streaks.
4. **Calm UI** — centered card layout.
5. **Soft constraints** — minimum is guidance, not enforcement.
6. **Ritual over optimization.**
7. **Durability over cleverness.**

---

## Tech Stack

- **Language:** Go
- **TUI Framework:** Bubble Tea
- **Architecture Style:** Core logic separated from UI (clean separation of concerns)

### Architectural Layers

- `core/` — journaling logic, prompt deck logic, minimum/progress logic
- `app/` — application services (open entry, save entry, draw prompt)
- `adapters/` — filesystem implementation
- `ui/` — Bubble Tea models and rendering logic
- `cmd/mainichi/` — CLI entrypoint and wiring

The UI must remain a thin layer over core functionality.

---

## File Structure

```
~/.mainichi/
    entries/
        YYYY-MM-DD.md
    prompts.txt
    prompt_state.json
    config.toml
```

### entries/
Daily Markdown files with frontmatter and raw body content.

### prompts.txt
Plaintext file. One prompt per line.

### prompt_state.json
Tracks shuffled deck state (remaining + used).

### config.toml
User configuration (minimum, adaptive mode, etc).

---

## Entry Format

All entries are Markdown files with frontmatter.

Example:

```md
---
date: 2026-02-12
prompt: "Where are you mistaking motion for progress?"
minimum: 300
---

I woke up today thinking about...
```

### Rules

- Frontmatter is **always written**.
- Prompt stored in metadata only.
- Body is raw stream-of-consciousness text.
- No auto-formatting.
- No hard wrapping.
- Soft wrapping handled in UI only.

The file must remain readable with `cat` or any editor.

---

## Writing View (TUI Design)

### Layout

- Centered writing card
- Fixed width (~70 characters)
- Fixed height (~6–8 visible lines)
- Soft wrap
- Vertical scroll beyond visible area
- Minimal border
- No heavy UI chrome

### Mockup

```
                      mainichi — 2026-02-12

             ┌──────────────────────────────────────┐
             │ I woke up today with a strange       │
             │ tension in my chest. It wasn't work  │
             │ exactly. It was the feeling that I   │
             │ keep chasing intensity instead of    │
             │ consistency.                         │
             │                                      │
             └──────────────────────────────────────┘

                 ▓▓▓▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░
```

### Prompt Display (Session Only)

Prompt appears above the card but is not written into body.

```
Prompt (steering): What are you avoiding that you already understand?
```

Prompt must not dominate visually.

---

## Progress Bar

- Based on configured minimum.
- No numeric display.
- Fills proportionally.
- When minimum reached → bar fully filled.
- No celebration or gamification.

### Example

```
before: ▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░░░░░
after : ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
```

---

## Minimum Logic

- Default minimum: 300 words (configurable).
- User may exit anytime.
- Minimum is a soft target.

### Optional Adaptive Mode (Future)

- If user consistently exceeds minimum,
  the full-bar threshold may gradually increase.
- No visible announcement.
- Never shrinks automatically.

---

## Prompts

### Prompt Deck (Default)

- Random draw from `prompts.txt`.
- Prompts shuffled into a finite deck.
- Each prompt used once before reshuffle.
- Prompt counts as used only if entry is saved.

### Prompt Lifecycle

```
prompts.txt
     ↓
shuffle indices
     ↓
remaining prompts
     ↓
draw one
     ↓
save entry → mark used
     ↓
deck empty → reshuffle
```

## CLI Interface

```
mainichi              # open today's entry
mainichi prompt       # open with the configured prompt source (stoic by default)
mainichi stoic        # open with the Daily Stoic heading
mainichi date         # open calendar view
mainichi YYYY-MM-DD   # open specific date
```

No todo system.

---

## Calendar View

- Monthly view
- Weekly view
- Presence markers only

Markers:
- ● written
- ◐ below minimum
- · empty

No streaks.
No totals.
No analytics.

### Monthly Mockup

```
February 2026

Mo Tu We Th Fr Sa Su
                   ·
·  ●  ●  ◐  ●  ·  ·
●  ●  ·  ·  ●  ●  ·
·  ●  ●  ●  ·  ·  ·
●  ·  ·  ·  ●  ·  ·
```

---

## Deliberate Non-Features

- No streak tracking
- No word count display
- No charts
- No stats
- No tagging system
- No attachments
- No collaboration
- No cloud sync
- No performance metrics

If users want analytics, they can grep the directory.

---

## Long-Term Vision

mainichi should:

- Still work in 10 years.
- Still open with `cat`.
- Still be readable without the app.
- Feel like a notebook.
- Age well.

It should feel like a daily object.  
Not like a product.
