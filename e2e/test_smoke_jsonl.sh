#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=e2e/lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

setup_args "$@"
start_app smoke-jsonl
trap cleanup_app EXIT

wait_for "input[0]"
pane=$(capture)
assert_contains "$pane" "▼ [0]: (3)"
assert_contains "$pane" "type: \"create\""
assert_contains "$pane" "id: 1"
assert_contains "$pane" "ok: true"
assert_contains "$pane" "▽ [1]: (3)"
assert_contains "$pane" "[2]: (3)"
assert_contains "$pane" "input[0]"
assert_contains "$pane" "$E2E_FILE"
assert_not_contains "$pane" "▼ (3)"

printf '%s\n' "$pane"
quit_app
trap - EXIT
