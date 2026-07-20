#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=e2e/lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

setup_args "$@"
start_app print
trap cleanup_app EXIT

wait_for "input"
tmux resize-window -t "$E2E_TARGET" -x 40 -y 15
send_keys p
wait_for "p█"
send_keys p
wait_for "Press any key to continue."
pane=$(capture_history)
assert_contains "$pane" '"b": 1,'
assert_contains "$pane" '"score": 1e-9'
assert_contains "$pane" '"longString": "abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789"'

send_keys x
wait_for "input"
pane=$(capture)
assert_not_contains "$pane" "Press any key to continue."

printf '%s\n' "$pane"
quit_app
trap - EXIT
