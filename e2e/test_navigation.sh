#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=./lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

setup_args "$@"
start_app navigation
trap cleanup_app EXIT

wait_for "input"

send_keys j
wait_for "input.b"
pane=$(capture)
assert_contains "$pane" "▶ b: 1"
assert_contains "$pane" "input.b"

send_keys j
wait_for "input.a"
pane=$(capture)
assert_contains "$pane" "▶ a: 2"
assert_contains "$pane" "input.a"

send_keys G
wait_for "input.longString"
pane=$(capture)
assert_contains "$pane" "▶ longString:"
assert_contains "$pane" "22/22"

send_keys g
wait_for "$E2E_FILE  1/22"
pane=$(capture)
assert_contains "$pane" "▼ (8)"
assert_contains "$pane" "input"

printf '%s\n' "$pane"
quit_app
trap - EXIT
