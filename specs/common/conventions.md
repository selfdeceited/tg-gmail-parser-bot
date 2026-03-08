# Code Conventions {#conventions}

## Telegram Handlers {#telegram}

- Each command handler lives in its own file under `internal/telegram/` (e.g. `start.go`, `register.go`)
- `handlers.go` contains only `RegisterHandlers` and `DefaultHandler` — no logic
- Handlers that need DB access accept `*gorm.DB` and return `tgbot.HandlerFunc` (closure pattern):
  ```go
  func RegisterHandler(db *gorm.DB) tgbot.HandlerFunc {
      return func(ctx context.Context, b *tgbot.Bot, update *models.Update) { ... }
  }
  ```
- `DefaultHandler(db)` routes non-command messages to active conversation flows via `HandleConversation`

## Multi-Step Conversation Flows {#conversations}

- Per-user state stored in `internal/telegram/state.go` — in-memory `map[int64]*T` protected by `sync.Mutex`
- State structs hold the current step (enum) and any data accumulated across messages
- Steps are handled inside `HandleConversation`, branching on `s.step`
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
- `internal/db/crypto.go` — `EncryptToken` / `DecryptToken` (AES-256-GCM)

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
| `TOKEN_ENCRYPTION_KEY` | yes | Base64-encoded 32-byte key — generate with `openssl rand -base64 32` |

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
- Refresh token is AES-256-GCM encrypted before persistence via `EncryptToken`
