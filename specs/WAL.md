# WAL — Write-Ahead Log

_Max ~3000 tokens. Collapse completed work into single lines._

## Current Phase

**IMPLEMENTATION** — /watch feature complete

## In Progress

_(none)_

## Completed

- [2026-03-07] spec scaffold: architecture, stack, module structure (spec://common/main, spec://common/structure)
- [2026-03-07] module specs: /start, /register, /configure, /addprompt, /watch flows (spec://modules/config, spec://modules/telegram)
- [2026-03-07] bot server + /start handler + /register stub (spec://modules/config#start)
- [2026-03-07] handlers split into per-file convention; documented in BOOT.md
- [2026-03-08] /register + configure button handler (spec://config/config#register)
- [2026-03-09] /watch feature (spec://config/telegram#watch)
- [2026-03-07] persistence layer (spec://modules/db)
- [2026-03-13] configurable Gmail account index (`GmailAccountIndex` on User entity; `/configure` UI with inline button flow; `SetGmailAccountIndex` command + service method; threaded through `FetchNewMessages` and watch poll)
- [2026-03-13] CI/CD: bumped GitHub Actions to latest versions; added Dependabot config (gomod, docker, github-actions weekly)
- [2026-03-13] services wiring refactored into `cmd/bot/services.go` (`services` struct + `wireServices`); convention documented in `specs/common/conventions.md`
- [2026-03-13] `StartBot` moved to `internal/telegram/setup.go` (package `telegram`)
- [2026-03-13] handlers restructured into `internal/telegram/handlers/` subfolder:
  - `handlers/` (package `handlers`) — helpers.go, start_handler.go, clearregistration_handler.go, watch_handler.go, configure_handler.go
  - `handlers/register/` — register_handler.go, register_state.go
  - `handlers/addprompt/` — addprompt_handler.go, addprompt_state.go
  - `handlers/gmailaccount/` — gmailaccount_handler.go, gmailaccount_state.go
  - `internal/telegram/setup.go` (renamed from handlers.go) — RegisterHandlers, DefaultHandler, StartBot

## Known Issues

1. `specs/common/structure.md` references `src/` as implementation root — incompatible with Go module layout. Implementation uses `cmd/`+`internal/` at project root. REVIEW marker added in `handlers.go`.

## Decisions Pending

- Gmail integration: OOB flow (`urn:ietf:wg:oauth:2.0:oob`) is deprecated by Google — may need device flow or hosted redirect in future
- Deployment: target environment (VPS, serverless, container)
- Email parsing strategy: which fields to extract, formatting rules

## Watch Out

- `internal/telegram/handlers.go`: `models.ParseModeMarkdown` = MarkdownV2 in go-telegram/bot (counterintuitive naming — do NOT change to ParseModeMarkdownV2, it does not exist)
- `TOKEN_ENCRYPTION_KEY` must be base64-encoded 32-byte key — generate with `openssl rand -base64 32`
- Claude model: use `anthropic.ModelClaudeHaiku4_5_20251001` — `claude-3-5-haiku-latest` returns 404
- `.human/` — off-limits, never read or modify

## Session Notes

_(cleared each session)_
