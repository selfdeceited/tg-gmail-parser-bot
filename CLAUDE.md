# Claude Instructions

## Session Start Protocol

1. Read `specs/BOOT.md` — entry point, session context
2. Read `specs/WAL.md` — current state, what to do next
3. Read `specs/common/main.md` — architecture and stack decisions
4. Read relevant module specs before touching any module code

## Spec System

- `specs/` contains the **source of truth** — implementation is a compiled artifact
- Losing implementation is recoverable; losing specs is catastrophic
- Do NOT modify specs unless explicitly asked
- When you disagree with a spec: implement as written, then add a REVIEW marker

## Conflict Protocol

When disagreeing with a spec decision:

```
<!-- REVIEW: §X.Y could benefit from [alternative] instead of [current],
because [reason] -->
```

Do not silently deviate from specs.

## Commit Convention

Each commit maps to one logical change linked to a spec item:

```
[module] short description
Implements: spec://module/document#section
```

## End of Session

Update `specs/WAL.md` with:
- What was completed
- What is in progress (with file locations)
- Any new known issues or pending decisions

## Off-Limits

- `.human/` — human-only buffer, AI must not read or modify
