# Telegram Gmail Parser Bot

This is a Telegram bot that allows you to parse Gmail emails, set up filters and prompts on how to interpret them and receive notifications via Telegram.

## Usage

TBD

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

## Things to improve
 - Notes on the GCP registration process
