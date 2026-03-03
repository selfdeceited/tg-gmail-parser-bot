# SPEC Protocol — How to Update Specifications

## Spec as Source of Truth

Specifications are not documentation. They are the **source of truth** and the primary IPC channel between human and AI. Implementation is a compiled artifact derived from specs.

- Losing implementation: recoverable (re-derive from specs)
- Losing specs: catastrophic (intent is gone)

## Addressability

Every spec element must be precisely locatable:

```
spec://<module>/<document>#<section>[.<subsection>]
```

Examples:
- `spec://common/main#stack`
- `spec://gmail/PROP-001#polling.interval`
- `spec://telegram/FEAT-001#formatting.subject`

Use `{#anchor-id}` in Markdown headings to enable direct linking:

```markdown
## Polling Interval {#polling.interval}
```

## Who Can Modify Specs

- **Human:** can modify any spec at any time
- **AI:** implements specs as written; proposes changes via REVIEW markers only

## Conflict Protocol

When AI disagrees with a spec decision:

1. **Implement as written** — do not deviate silently
2. **Add a REVIEW marker** immediately after the implemented section:

```
<!-- REVIEW: §polling.interval — fixed 60s interval may cause rate-limit
issues under high email volume. Consider exponential backoff starting at
30s, because Gmail API quota is per-user per-second. -->
```

3. **Do not resolve REVIEW markers yourself** — leave them for human review

## Spec Update Rules

### Adding a new section
- Add to the appropriate `PROP-*.md` or `FEAT-*.md` file
- Include anchor: `## Section Name {#section.name}`
- Reference in WAL under "In Progress"

### Changing an existing decision
- Update the spec text
- Add a changelog entry in the spec file:

```markdown
## Changelog
- 2026-03-01: Changed polling interval from 60s to 30s — rationale: [reason]
```

### Removing a section
- Mark as deprecated first: `## Section Name ~~(deprecated)~~`
- Remove in the next logical commit

## Commit Convention

Each commit = one logical change tied to a spec item:

```
[module] implement gmail polling

Implements: spec://gmail/PROP-001#polling
```

## Spec File Types

| File | Purpose |
|------|---------|
| `PROP-*.md` | Foundational decisions, protocols, algorithms |
| `FEAT-*.md` | Feature specifications, user-facing behavior |
| `common/main.md` | Architecture, stack, cross-cutting decisions |
| `common/structure.md` | Module map and boundaries |
