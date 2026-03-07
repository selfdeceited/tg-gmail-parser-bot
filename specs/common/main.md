# Architecture & Key Decisions {#main}

*Actors*
- Gmail (external API)
- Telegram bot (hosted backend)
- Claude SDK (external API)

*Commands*
 - `start` - user registers in the bot and sends /start. It registers the user and starts the installation guide telling how to set up Gmail integration. More details at `spec://config/config#start`.
 - `/register` - This command is used to link new Gmail account for monitoring. More details at `spec://config/config#register`.
 - `/configure` - This command is used to configure filters and prompts for summarization. It's not available if `register` command result is not successful. More details at `spec://config/config#configure`.
 - `/watch` - This command is used to start the integration. It's not available if `register` command result is not successful or no prompts are configured. More details at `spec://config/telegram#watch`.


_Cross-cutting decisions that apply to the whole project._

## Project Overview {#overview}

**tg-gmail-parser-bot** — the Telegram bot that monitors one or more Gmail inboxes and
forwards parsed job feedback email summaries back to you.

## Tech Stack {#stack}


| Concern | Choice | Rationale |
|---------|--------|-----------|
| Runtime | Go | |
| Language | Golang | |
| Gmail integration | [API package](https://pkg.go.dev/google.golang.org/api/gmail/v1) | |
| Telegram integration | [go-telegram Package](https://github.com/go-telegram/bot) | |
| Persistence | PostgreSQL | |
| Deployment | Netlify | |

## Authentication {#auth}

### Gmail {#auth.gmail}
> use user OAuth flow from GCP

### Telegram {#auth.telegram}
> Bot token via `@BotFather`

## Email Polling Strategy {#polling}

> Use periodic polling (polling interval: 15 minutes, configurable)


## Email Parsing Rules {#parsing}

> TBD: which fields to extract and how to format them

Fields under consideration:
- `from`, `subject`, `date`
- Body: full
- Attachments: skip

## Telegram Message Formatting {#formatting}

> Will use Markdown parse mode

## Error Handling {#errors}

> TBD

## Changelog {#changelog}

_(empty — project not started)_
