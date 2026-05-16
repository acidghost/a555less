#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=e2e/lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

setup_args "$@"
start_app help-expanded
trap cleanup_app EXIT

wait_for "input"
send_keys "?"
wait_for "collapse/parent"
pane=$(capture)
assert_contains "$pane" "k/↑   up"
assert_contains "$pane" "j/↓   down"
assert_contains "$pane" "h/←   collapse/parent"
assert_contains "$pane" "l/→   expand/child"
assert_contains "$pane" "space toggle"
assert_contains "$pane" "g    top"
assert_contains "$pane" "G    bottom"
assert_contains "$pane" "pgup page up"
assert_contains "$pane" "pgdn page down"
assert_contains "$pane" "C-u  half up"
assert_contains "$pane" "C-d  half down"
assert_contains "$pane" "H parent"
assert_contains "$pane" "J next sibling"
assert_contains "$pane" "K prev sibling"
assert_contains "$pane" "? toggle help"
assert_contains "$pane" "q quit"

printf '%s\n' "$pane"
quit_app
trap - EXIT
