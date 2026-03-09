# Telegram Gmail Parser Bot
`@tg_gmail_ai_parser_bot`

This is a Telegram bot that allows you to parse Gmail emails, set up filters and prompts on how to sum them up and receive notifications via Telegram.

## Usage

- Set up Gmail Oauth integration using Google Cloud. The bot will have quick instructions on that.
- Configure your prompts
<img width="722" height="340" alt="image" src="https://github.com/user-attachments/assets/fc9b286d-8866-477b-855a-1cbbfb396a6c" />

- Receive notifications after `/watch` commands based on your prompts
<img width="499" height="178" alt="image" src="https://github.com/user-attachments/assets/6b0bf0d6-7019-45b1-8d3a-b9eb2ce691db" />


## Local Development

In order to fork/run this bot locally, you will need to have:
 - golang
 - telegram bot token created via `@BotFather`. Env var: `TELEGRAM_BOT_TOKEN`
 - Claude api key

## Credential Security

Gmail OAuth credentials (client ID, client secret, refresh token) are stored encrypted in PostgreSQL.

### Encryption model

- **Algorithm**: AES-256-GCM with a random 12-byte nonce per write.
- **Per-user keys**: a unique 32-byte subkey is derived per user via HKDF-SHA256 using the master key and the user's Telegram ID as salt. A leaked subkey exposes only one user's data.
- **Versioned blobs**: ciphertext is stored as `v<N>.<nonce_b64>.<ciphertext_b64>`. The version identifies which master key was used, enabling zero-downtime rotation.

### Key configuration

Set these environment variables (see `.env.example`):

```
TOKEN_ENCRYPTION_KEY_1=<base64-encoded 32 bytes>   # openssl rand -base64 32
TOKEN_ENCRYPTION_KEY_CURRENT=1
```

### Key rotation

To rotate to a new master key without downtime:

1. Generate a new key: `openssl rand -base64 32`
2. Add it as `TOKEN_ENCRYPTION_KEY_2=<new key>` and set `TOKEN_ENCRYPTION_KEY_CURRENT=2`
3. Restart the bot (new credentials will use key 2 immediately)
4. Call `RegistrationService.RotateCredentials` (e.g. via an admin command) to re-encrypt all existing rows with key 2
5. Once rotation completes, `TOKEN_ENCRYPTION_KEY_1` can be removed

### Revoking access

To revoke a user's Gmail access:
- The user can run `/clearregistration` — this hard-deletes their credentials from the database
- Separately, revoke the OAuth token in [Google Account Permissions](https://myaccount.google.com/permissions) to prevent any cached token from being reused

## Known Trade-offs

### Email processing and `lastCheckedAt`

The poll timestamp (`lastCheckedAt`) is only advanced when Gmail fetch and message listing succeed. However, if Claude summarization fails for individual emails within an otherwise successful poll (e.g. transient Claude API error), those emails are **not retried** — the timestamp still advances and those emails are skipped on the next cycle.

The alternative (blocking the timestamp on any Claude failure) would cause duplicate Telegram notifications, since emails already successfully processed in the same batch would be re-sent on the next poll.

The correct long-term fix is to track processed Gmail message IDs in the database, enabling safe per-message retry. This is not currently implemented.

## Things to improve
 - Notes on the GCP registration process
 - Per-message ID tracking to allow safe retry on Claude failures
