#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=./lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

setup_args "$@"
start_app smoke
trap cleanup_app EXIT

wait_for "input"
pane=$(capture)
assert_contains "$pane" "▼ (8)"
assert_contains "$pane" "b: 1"
assert_contains "$pane" "a: 2"
assert_contains "$pane" "dup: \"first\""
assert_contains "$pane" "dup: \"second\""
assert_contains "$pane" "users: (2)"
assert_contains "$pane" "[0]: (4)"
assert_contains "$pane" "input"
assert_contains "$pane" "$E2E_FILE"

printf '%s\n' "$pane"
quit_app
trap - EXIT
