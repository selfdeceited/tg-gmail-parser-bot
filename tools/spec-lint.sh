#!/usr/bin/env bash
# spec-lint.sh — verify referential integrity of spec:// URIs
#
# Checks that every spec://module/document#section reference in specs/
# points to an anchor that actually exists in the target file.
#
# Usage: ./tools/spec-lint.sh [--verbose]

set -euo pipefail

SPECS_DIR="$(cd "$(dirname "$0")/../specs" && pwd)"
VERBOSE=0
[[ "${1:-}" == "--verbose" ]] && VERBOSE=1

errors=0
checked=0

log() { [[ $VERBOSE -eq 1 ]] && echo "$@" || true; }

# Extract all spec:// URIs from markdown files
while IFS= read -r file; do
  while IFS= read -r uri; do
    # Parse: spec://module/document#section
    # Strip markdown link syntax if present
    uri="${uri%%)*}"
    uri="${uri##*(}"

    module_doc="${uri#spec://}"
    module="${module_doc%%/*}"
    rest="${module_doc#*/}"
    document="${rest%%#*}"
    anchor=""
    [[ "$rest" == *"#"* ]] && anchor="${rest#*#}"

    # Resolve file path
    target_file="$SPECS_DIR/modules/$module/${document}.md"
    # Also check common/
    if [[ ! -f "$target_file" ]]; then
      target_file="$SPECS_DIR/common/${document}.md"
    fi

    checked=$((checked + 1))

    if [[ ! -f "$target_file" ]]; then
      echo "ERROR: $file → $uri — target file not found: $target_file"
      errors=$((errors + 1))
      continue
    fi

    log "OK file: $uri → $target_file"

    if [[ -n "$anchor" ]]; then
      # Check anchor exists (as {#anchor} or as generated slug)
      if ! grep -qE "\{#${anchor}\}|^#{1,6}[[:space:]].*${anchor}" "$target_file"; then
        echo "ERROR: $file → $uri — anchor '#${anchor}' not found in $target_file"
        errors=$((errors + 1))
      else
        log "OK anchor: #${anchor}"
      fi
    fi

  done < <(grep -oE 'spec://[a-zA-Z0-9/_.-]+#?[a-zA-Z0-9._-]*' "$file" || true)
done < <(find "$SPECS_DIR" -name "*.md" -type f)

echo ""
echo "spec-lint: checked $checked URI(s), $errors error(s)"
[[ $errors -eq 0 ]] && exit 0 || exit 1
