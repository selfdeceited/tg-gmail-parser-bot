## DB
We'll use PostgreSQL with GORM package. The connection string is stored in `DATABASE_URL` env var.

### Migrations
We'll use GORM's built-in `AutoMigrate` (code-first).

### Package Layout

- `internal/db/` — connection setup
- `internal/db/entities/` — GORM model structs
- `internal/db/queries/` — read operations (SELECT)
- `internal/db/commands/` — write operations (INSERT/UPDATE/DELETE)

### Database Structure

1. `Users` table
  - `id` (int64, primary key) — Telegram user ID
  - `is_registered` (boolean) — has Gmail account linked
  - `is_active` (boolean) — /watch is running
  - `last_active_at` (timestamp)
  - `created_at`, `updated_at`, `deleted_at` — via `gorm.Model` fields (soft delete)

2. `Prompts` table
  - `id` (UUID, primary key)
  - `user_id` (int64, foreign key → Users.id)
  - `prompt` (text)
  - `filter` (text)
  - `created_at`, `updated_at`, `deleted_at` — via `gorm.Model` fields (soft delete replaces is_active)

3. `Credentials` table
  - `id` (UUID, primary key)
  - `user_id` (int64, foreign key → Users.id) — one credential per user
  - `encrypted_refresh_token` (text) — AES-256-GCM encrypted, format: `<base64(nonce)>.<base64(ciphertext)>`
  - `created_at`, `updated_at`, `deleted_at` — via `gorm.Model` fields

### Indexes
- `Prompts` — index on `user_id`

### Token Security

Only the OAuth **refresh token** is stored — access tokens (1h lifetime) are ephemeral and never persisted. Refresh tokens must be encrypted at rest using AES-256-GCM before storing and decrypted on read.

- Encryption key: 32-byte random value stored in `TOKEN_ENCRYPTION_KEY` env var (never in DB)
- Each token gets a unique random nonce; store as `<base64(nonce)>.<base64(ciphertext)>`
- If the DB leaks, tokens are useless without `TOKEN_ENCRYPTION_KEY`
- Refresh tokens are long-lived but invalidated if: user revokes access, unused for 6 months, or password changes
- UUID library: `github.com/google/uuid`

### Jobs
If `Users.last_active_at` is older than 30 days, soft-delete the user and hard-delete their credentials.
