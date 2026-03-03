# Architecture & Key Decisions {#main}

*Actors*
- Gmail (external API)
- Telegram bot (hosted backend)
- Claude SDK (external API)

*Scenario*
1) Gmail --> Telegram bot // new email arrives in inbox
2) Telegram bot -> Gmail // read new email content
3) Telegram bot --> Claude SDK // summarize content by specified prompt, return as JSON
4) Telegram bot --> Telegram chat // parsed summary with the letter result

Example:
- User sends /start bot command in the chat and enters his gcp oauth token
- New job feedback email arrives in Gmail inbox
- Telegram bot reads the email content and sends it to Claude SDK for summarization
- Claude SDK understands the following:
  - link to the email
  - whether the result is successful
  - if successful, the summary for the next steps.
- Telegram bot forwards parsed summary to configured chat to render it as a well-formatted message

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
| Persistence | TBD | |
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
