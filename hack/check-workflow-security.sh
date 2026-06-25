#!/usr/bin/env bash

# Lint GitHub Actions workflows for "pwn request" vulnerabilities.
#
# pull_request_target workflows run with write access to the base repo
# and access to secrets, even when triggered by a fork. This script
# detects patterns that would let an attacker exploit that privilege.

set -euo pipefail

WORKFLOWS_DIR=".github/workflows"
errors=0

check_file() {
  local wf="$1"

  # 1. Checkout of PR code
  #    Any checkout in a prt workflow means running attacker-controlled
  #    code with elevated privileges.
  if grep -qE 'uses:\s*actions/checkout' "$wf"; then
    echo "ERROR: $wf: actions/checkout in a pull_request_target workflow (code from forks runs with write token)"
    errors=$((errors + 1))
  fi

  # 2. Expression injection
  #    ${{ github.event.* }} and ${{ github.head_ref }} contain attacker-
  #    controlled strings. In run:/script: blocks they expand before the
  #    shell sees them, enabling command injection.
  #    Safe uses (if: conditions, env: with integer fields like .number)
  #    are excluded below. When a new safe use is needed, add it to the
  #    grep -v allowlist rather than weakening the check.
  local injections
  injections=$(grep -nE '\$\{\{.*github\.(event\.|head_ref)' "$wf" \
    | grep -vE '^\s*#' \
    | grep -vE 'if\s*:' \
    | grep -vE 'github\.event\.pull_request\.number' \
    | grep -vE 'github\.event\.pull_request\.user\.login' \
    | grep -vE 'github\.event\.label\.name' \
    | grep -vE 'github\.event\.pull_request\.labels' \
    || true)
  if [[ -n "$injections" ]]; then
    echo "ERROR: $wf: potential expression injection in pull_request_target workflow:"
    echo "$injections" | sed 's/^/  /'
    errors=$((errors + 1))
  fi

  # 3. Custom secrets
  #    Only GITHUB_TOKEN is safe in prt workflows. Custom secrets (PATs,
  #    deploy keys) are high-value exfiltration targets from forks.
  local custom_secrets
  custom_secrets=$(grep -nE 'secrets\.' "$wf" \
    | grep -vE 'secrets\.GITHUB_TOKEN' \
    | grep -vE '^\s*#' \
    || true)
  if [[ -n "$custom_secrets" ]]; then
    echo "ERROR: $wf: custom secrets in a pull_request_target workflow:"
    echo "$custom_secrets" | sed 's/^/  /'
    errors=$((errors + 1))
  fi

  # 4. Unpinned actions
  #    Tag-based refs (@v6, @main) are vulnerable to supply-chain attacks
  #    via tag mutation. Pin to a full commit SHA.
  local unpinned
  unpinned=$(grep -nE 'uses:\s*\S+@' "$wf" \
    | grep -vE '@[0-9a-f]{40}' \
    | grep -vE '^\s*#' \
    || true)
  if [[ -n "$unpinned" ]]; then
    echo "WARNING: $wf: actions not pinned by SHA:"
    echo "$unpinned" | sed 's/^/  /'
  fi
}

found_prt=false
for wf in "$WORKFLOWS_DIR"/*.yml "$WORKFLOWS_DIR"/*.yaml; do
  [[ -f "$wf" ]] || continue
  grep -q 'pull_request_target' "$wf" || continue
  found_prt=true
  check_file "$wf"
done

if [[ "$found_prt" == false ]]; then
  echo "No pull_request_target workflows found."
fi

if [[ $errors -gt 0 ]]; then
  echo ""
  echo "Found $errors vulnerabilities. See https://securitylab.github.com/resources/github-actions-preventing-pwn-requests/"
  exit 1
fi

echo "Workflow security checks passed."
