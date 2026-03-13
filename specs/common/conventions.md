# Code Conventions {#conventions}

## Service Layer {#service}

Business logic lives in `internal/service/`, between handlers (delivery) and DB/external adapters (infrastructure).

- Services are defined as **interfaces** in `internal/service/` and implemented in the same package using injected dependencies
- Handlers accept service interfaces — never `*gorm.DB` or infrastructure types directly
- Services call `internal/db/commands`, `internal/db/queries`, and `internal/gmail` packages
- All service construction and wiring lives in `cmd/bot/services.go` — never in `main.go` or anywhere else:
  - Declare every service as a field in the `services` struct
  - Instantiate every service inside `wireServices(db, cfg) services`
  - `main.go` calls `wireServices` once and passes fields to the bot/handlers
  ```go
  // services.go
  type services struct {
      registrationService service.RegistrationService
      promptService       service.PromptService
      watchService        service.WatchService
  }

  func wireServices(db *gorm.DB, cfg config) services { ... }
  ```
- Variable names use the full service name — never abbreviated suffixes like `regSvc`, `promptSvc`. Use `registrationService`, `promptService`, `watchService`.
- Avoid single-letter variable names. Use descriptive names: `bot` instead of `b`, `gmailService` instead of `svc`, etc.

## Development Workflow {#workflow}
- Run `make check` before committing — it executes `golangci-lint` then `go build` in order.
- Lint failures block the build. All lint issues must be resolved, not suppressed.

## Reliability {#reliability}
- Use timeouts and context cancellation to handle long-running operations gracefully. I/O timeout should be configurable (default to be 10s)

## Telegram Handlers {#telegram}

- Each command handler lives in its own file under `internal/telegram/`, named `<command>_handler.go` (e.g. `start_handler.go`, `register_handler.go`)
- `handlers.go` contains only `RegisterHandlers` and `DefaultHandler` — no logic
- Handlers accept a service interface and return `tgbot.HandlerFunc` (closure pattern):
  ```go
  func RegisterHandler(svc service.RegistrationService) tgbot.HandlerFunc {
      return func(ctx context.Context, b *tgbot.Bot, update *models.Update) { ... }
  }
  ```
- `ConversationHandler(svc)` routes non-command messages to active conversation flows via `HandleConversation`

## Multi-Step Conversation Flows {#conversations}

- Per-user state stored in `internal/telegram/state.go` — in-memory `map[int64]*T` protected by `sync.Mutex`
- State structs hold the current step (enum) and any data accumulated across messages
- Steps are handled inside `HandleRegisterConversation`, branching on `s.step`
- Always call `setState(userID, nil)` to clear state on both success and terminal failure

## Markdown Formatting {#markdown}

- Use `models.ParseModeMarkdown` — this is **MarkdownV2** in go-telegram/bot (counterintuitive naming)
- `models.ParseModeMarkdownV2` does **not exist** — do not attempt to use it
- Special characters in dynamic strings must be escaped; use `escapeMarkdown()` in `register.go`
- Static message strings use raw backtick literals with manual escaping (`\.`, `\!`, etc.)

## Database Layer {#db}

### Package structure
- `internal/db/entities/` — GORM model structs only
- `internal/db/commands/` — write operations (INSERT, UPDATE, DELETE)
- `internal/db/queries/` — read operations (SELECT)
- `internal/db/crypto.go` — `EncryptCredentials` / `DecryptCredentials` (AES-256-GCM)

### Migrations
- `AutoMigrate` runs inside `db.Connect()` on every startup — synchronous, before the bot accepts updates
- Only **adds** columns/tables — never drops or renames
- Renaming a struct field silently creates a new empty column; old column stays — manual `ALTER TABLE` required
- No migration history or rollback support

## Environment Variables {#env}

| Variable | Required | Description |
|---|---|---|
| `TELEGRAM_BOT_TOKEN` | yes | Bot token from @BotFather |
| `DATABASE_URL` | yes | PostgreSQL connection string |
| `TOKEN_ENCRYPTION_KEY_<N>` | yes* | Base64-encoded 32-byte key for version N — generate with `openssl rand -base64 32` |
| `TOKEN_ENCRYPTION_KEY_CURRENT` | yes* | Integer — which version to use for new encryptions (e.g. `1`) |
| `TOKEN_ENCRYPTION_KEY` | legacy | Single-key fallback, treated as version 1 — no rotation support |

\* Use the versioned form (`TOKEN_ENCRYPTION_KEY_<N>` + `TOKEN_ENCRYPTION_KEY_CURRENT`) for all new deployments. To rotate: add `TOKEN_ENCRYPTION_KEY_<N+1>`, bump `CURRENT`, then call `RegistrationService.RotateCredentials`.

## Logging {#logging}

- Use **[logrus](https://github.com/sirupsen/logrus)** (`github.com/sirupsen/logrus`) for all logging — never `log` from the standard library
- Configure once at startup in `main.go` (format, level); use package-level functions everywhere else (`logrus.Info`, `logrus.Error`, etc.)
- **All significant successful outcomes must be logged at `Info` level**, including but not limited to:
  - Gmail smoke test pass
  - Credentials saved to DB
  - User marked as registered
  - Bot startup
- Errors use `logrus.WithError(err).Error(...)` or `logrus.WithError(err).Fatal(...)`
- Include structured fields for user-scoped events:
  ```go
  logrus.WithField("user_id", userID).Info("gmail smoke test passed")
  ```

## Gmail OAuth {#gmail-oauth}

- Desktop app OOB redirect: `urn:ietf:wg:oauth:2.0:oob` — user copies code from browser page
- Always use `oauth2.AccessTypeOffline` + `oauth2.ApprovalForce` to guarantee a refresh token is returned
- Smoke test (read last message) is run before saving any credentials
- Refresh token is AES-256-GCM encrypted before persistence via `EncryptCredentials`


## General development
 - enhance centralizing logging with scopes and enrichment by tgId, session, etc
