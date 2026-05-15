#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=./lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

setup_args "$@"
start_app collapse
trap cleanup_app EXIT

wait_for "input"

send_keys j j j j j Space
wait_for "▶ users: (2)"
pane=$(capture)
assert_contains "$pane" "▶ users: (2)"
assert_contains "$pane" "input.users"
assert_contains "$pane" "6/9"
assert_not_contains "$pane" "name: \"alice\""

send_keys Right Right
wait_for "input.users[0]"
pane=$(capture)
assert_contains "$pane" "▼ [0]: (4)"
assert_contains "$pane" "name: \"alice\""
assert_contains "$pane" "input.users[0]"

printf '%s\n' "$pane"
quit_app
trap - EXIT
