# WAL — Write-Ahead Log

_Max ~3000 tokens. Collapse completed work into single lines._

## Current Phase

**IMPLEMENTATION** — configure button verification complete, next: /clearregistration or full /configure

## In Progress

_(none)_

## Completed

- [2026-03-07] spec scaffold: architecture, stack, module structure (spec://common/main, spec://common/structure)
- [2026-03-07] module specs: /start, /register, /configure, /addprompt, /watch flows (spec://modules/config, spec://modules/telegram)
- [2026-03-07] bot server + /start handler + /register stub (spec://modules/config#start)
- [2026-03-07] handlers split into per-file convention; documented in BOOT.md
- [2026-03-08] /register + configure button handler (spec://config/config#register)
  - `internal/telegram/state.go` — per-user in-memory conversation state machine
  - `internal/gmail/oauth.go` — ParseCredentials, BuildAuthURL, ExchangeCode (OOB redirect)
  - `internal/gmail/smoketest.go` — SmokeTest + VerifyRefreshToken (reads last email via Gmail API)
  - `internal/telegram/register.go` — 3-step flow + double-registration guard w/ live smoke test; clears stale creds
  - `internal/telegram/configure.go` — ConfigureButtonHandler: live smoke test; clearRegistration helper
  - `internal/telegram/handlers.go` — wired DB; DefaultHandler routes to HandleConversation; registered "⚙️ Configure"
  - `cmd/bot/main.go` — passes db to RegisterHandlers and DefaultHandler
  - Credential storage redesigned: single `EncryptedCredentials` JSON blob (client_id + client_secret + refresh_token)
  - `internal/db/entities/credential.go` — renamed EncryptedRefreshToken → EncryptedCredentials
  - `internal/db/commands/credential.go` — UpsertCredentials (JSON marshal + encrypt)
  - `internal/db/queries/credential.go` — GetCredentials (decrypt + unmarshal → StoredCredentials)
  - New deps: golang.org/x/oauth2, google.golang.org/api/gmail/v1, github.com/sirupsen/logrus
  - All logging migrated to logrus; success events logged at Info with structured fields
- [2026-03-07] persistence layer (spec://modules/db)
  - `internal/db/db.go` — Connect(), AutoMigrate
  - `internal/db/crypto.go` — EncryptToken/DecryptToken (AES-256-GCM, TOKEN_ENCRYPTION_KEY env var)
  - `internal/db/entities/` — User (int64 PK), Prompt (UUID PK, soft delete), Credential (UUID PK, uniqueIndex user_id)
  - `internal/db/queries/` — GetUser, GetActivePrompts, GetRefreshToken
  - `internal/db/commands/` — UpsertUser, SetRegistered, SetActive, AddPrompt, DeletePrompt, UpsertRefreshToken, DeleteCredential
  - `.env.example` — added TOKEN_ENCRYPTION_KEY example

## Known Issues

1. `specs/common/structure.md` references `src/` as implementation root — incompatible with Go module layout. Implementation uses `cmd/`+`internal/` at project root. REVIEW marker added in `handlers.go`.

## Decisions Pending

- Gmail integration: OOB flow (`urn:ietf:wg:oauth:2.0:oob`) is deprecated by Google — may need device flow or hosted redirect in future
- Deployment: target environment (VPS, serverless, container)
- Email parsing strategy: which fields to extract, formatting rules

## Watch Out

- `internal/telegram/handlers.go`: `models.ParseModeMarkdown` = MarkdownV2 in go-telegram/bot (counterintuitive naming — do NOT change to ParseModeMarkdownV2, it does not exist)
- `TOKEN_ENCRYPTION_KEY` must be base64-encoded 32-byte key — generate with `openssl rand -base64 32`
- `.human/` — off-limits, never read or modify

## Session Notes

_(cleared each session)_
