# Module Structure {#structure}

_Map of modules, their boundaries, and responsibilities._

## Module Map {#modules}

```
tg-gmail-parser-bot/
├── specs/                        # IPC buffer (source of truth)
│   ├── BOOT.md                   # Session entry point
│   ├── WAL.md                    # Continuation state
│   ├── WAL-PROTOCOL.md           # WAL maintenance rules
│   ├── SPEC-PROTOCOL.md          # Spec update and conflict rules
│   ├── common/
│   │   ├── main.md               # Architecture, stack, decisions
│   │   └── structure.md          # This file
│   └── modules/
│       ├── gmail/                # Gmail integration
│       ├── telegram/             # Telegram bot integration
│       ├── parser/               # Email parsing and formatting
│       └── config/               # Configuration and secrets
├── src/                          # Implementation (artifact)
├── tests/                        # Executable specs
├── tools/
│   └── spec-lint.sh              # Spec link integrity check
├── .human/                       # Human-only (AI-ignored)
├── .claudeignore
└── CLAUDE.md
```

## Module Responsibilities {#responsibilities}

### gmail {#gmail}
- Authenticate with Gmail API
- Poll for new emails
- Fetch email content (headers + body)

Specs: `specs/modules/gmail/`

### telegram {#telegram}
- Authenticate bot with Telegram API
- Send formatted messages to configured chats/channels
- Handle delivery errors and retries

Specs: `specs/modules/telegram/`

### parser {#parser}
- Extract relevant fields from raw email (from, subject, date, body)
- Format content for Telegram message constraints
- Handle encoding, HTML stripping, truncation

Specs: `specs/modules/parser/`

### claude {#claude}
- Authenticate with Claude API
- Summarizing prompt and output JSON structure

Specs: `specs/modules/parser/`

### config {#config}
- Load and validate configuration (Gmail credentials, Telegram token, chat IDs, prompts)
- Environment variable / file-based config

Specs: `specs/modules/config/`

## Module Boundaries {#boundaries}

- `gmail` → outputs raw email data structs; knows nothing about Telegram
- `parser` → pure transformation; no I/O
- `telegram` → consumes formatted strings; knows nothing about Gmail

## Changelog {#changelog}

_(empty — project not started)_
