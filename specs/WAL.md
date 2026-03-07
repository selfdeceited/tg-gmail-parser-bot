# WAL — Write-Ahead Log

_Max ~3000 tokens. Collapse completed work into single lines._

## Current Phase

**IMPLEMENTATION** — bot server skeleton with /start command

## In Progress

_(none — /start complete, next: /register)_

## Completed

- [2026-03-07] spec scaffold: architecture, stack, module structure (spec://common/main, spec://common/structure)
- [2026-03-07] module specs: /start, /register, /configure, /addprompt, /watch flows (spec://modules/config, spec://modules/telegram)
- [2026-03-07] bot server + /start handler + /register stub (spec://modules/config#start)
  - `cmd/bot/main.go` — entry point, godotenv, SetMyCommands, signal shutdown
  - `internal/telegram/handlers.go` — StartHandler (GCP OAuth guide), RegisterHandler (stub), DefaultHandler (no-op)
  - `go.mod` — module `github.com/selfdeceited/tg-gmail-parser-bot`, deps: go-telegram/bot, godotenv
  - `.env.example`, `.gitignore`

## Known Issues

1. `specs/common/structure.md` references `src/` as implementation root — incompatible with Go module layout. Implementation uses `cmd/`+`internal/` at project root. REVIEW marker added in `handlers.go`.

## Decisions Pending

- Gmail integration: OAuth2 flow details, polling interval configuration
- Persistence: PostgreSQL schema (needed for /register — storing user + Gmail credentials)
- Deployment: target environment (VPS, serverless, container)
- Email parsing strategy: which fields to extract, formatting rules

## Watch Out

- `internal/telegram/handlers.go`: `models.ParseModeMarkdown` = MarkdownV2 in go-telegram/bot (counterintuitive naming — do NOT change to ParseModeMarkdownV2, it does not exist)
- `.human/` — off-limits, never read or modify

## Session Notes

_(cleared each session)_
