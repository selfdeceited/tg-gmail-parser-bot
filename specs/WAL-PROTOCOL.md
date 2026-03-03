# WAL Protocol — How to Maintain the Write-Ahead Log

## Purpose

WAL captures continuation state between sessions. It answers: "where did we stop, what's next?"

## Structure

```markdown
## Current Phase
<phase name> — <one-line description>

## In Progress
- <spec-uri>: description
  - DONE: subtask (spec://...)
  - TODO: subtask (spec://...)

## Completed
- [date] short description (spec://...)

## Known Issues
1. <issue description>

## Decisions Pending
- spec://... : question to resolve

## Watch Out
<anti-instructions: what NOT to touch>

## Session Notes
<cleared each session>
```

## Rules

### On Session Start
1. Read WAL to recover context
2. Treat "Watch Out" as hard constraints
3. Do NOT act on stale "In Progress" items without verifying current file state

### During Work
- When starting a sub-task, add it to "In Progress" with its spec URI
- When completing a sub-task, mark it `DONE`

### On Session End
1. Move completed items from "In Progress" to "Completed" (collapse to one line)
2. Update "In Progress" with exact file locations for the next session
3. Add any new known issues to "Known Issues"
4. Add any unresolved design questions to "Decisions Pending"
5. Add anti-instructions to "Watch Out" for stable code that must not be touched

## Size Limit

Keep WAL under ~3000 tokens. Aggressively collapse "Completed" entries:

```markdown
## Completed
- [2026-03-01] scaffold, auth, gmail-polling (spec://common/main, spec://gmail/PROP-001)
```

## Signal vs Noise

WAL should be scannable in ~30 seconds. If it takes longer, it's too long.
