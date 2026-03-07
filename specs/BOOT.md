# BOOT — Session Entry Point

**Project:** tg-gmail-parser-bot
**Purpose:** Telegram bot that monitors Gmail and forwards parsed emails to Telegram chats

## On Session Start

1. Read this file
2. Read `WAL.md` — continuation state and what to do next
3. Read `common/main.md` — architecture, stack, key decisions
4. Read `common/structure.md` — module map
5. Read relevant module specs before working on a module

## Spec URI Scheme

```
spec://<module>/<document>#<section>[.<subsection>]
```

Example: `spec://common/main#stack.runtime`

## Critical Rules

- Specs in `specs/` are **source of truth**; `src/` is a compiled artifact
- Do NOT modify specs unless explicitly asked
- Add `<!-- REVIEW: ... -->` for suggestions, never silently deviate
- `.human/` is off-limits — never read or write it
- Update `WAL.md` at the end of every session

## Module Index

See `common/structure.md` for full module map.

## Code Conventions

- Each Telegram handler lives in its own file under `internal/telegram/` (e.g. `start.go`, `register.go`)
- `handlers.go` contains only `RegisterHandlers` and `DefaultHandler`

## Database Migrations

`AutoMigrate` runs inside `db.Connect()` on every bot startup — synchronous, before the bot accepts updates.

- Only **adds** columns/tables — never drops or renames
- Renaming a struct field silently creates a new empty column; old column stays — manual `ALTER TABLE` required for renames
- No migration history or rollback support
- Acceptable for this project's scale; revisit if schema churn increases

## Environment Notes

- Go binary is at `/opt/homebrew/bin/go` (may need `export PATH=$PATH:/opt/homebrew/bin` in non-interactive shells)
- Dependencies: `github.com/go-telegram/bot`, `github.com/joho/godotenv`
- Run `go mod tidy` after adding/removing dependencies

## Protocols

- WAL maintenance: `WAL-PROTOCOL.md`
- Spec updates and conflict resolution: `SPEC-PROTOCOL.md`
