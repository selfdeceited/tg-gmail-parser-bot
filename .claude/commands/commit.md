Create a git commit following the Conventional Commits specification.

1. Run `git status` and `git diff --staged` to review what is staged. If nothing is staged, run `git diff` to see unstaged changes and stage the relevant files first.
2. Determine the commit type based on the changes:
   - `feat` — new feature
   - `fix` — bug fix
   - `docs` — documentation only
   - `refactor` — code change that neither fixes a bug nor adds a feature
   - `test` — adding or updating tests
   - `chore` — build process, tooling, dependencies
   - `perf` — performance improvement
   - `ci` — CI/CD configuration
   - `style` — formatting, whitespace (no logic change)
3. Determine an optional scope (module name, e.g. `gmail`, `parser`, `telegram`).
4. Write a short imperative subject line (max 72 chars): `type(scope): description`
5. If the change links to a spec item, add a footer: `Implements: spec://module/document#section`
6. Commit using:

```bash
git commit -m "$(cat <<'EOF'
type(scope): short description

Optional longer body if needed.

Implements: spec://...
Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
EOF
)"
```
