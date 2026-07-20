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
normalized=$(sed -E 's/[[:space:]]+/ /g' <<<"$pane")
assert_contains "$normalized" "k/↑ up"
assert_contains "$normalized" "j/↓ down"
assert_contains "$normalized" "h/← collapse/parent"
assert_contains "$normalized" "l/→ expand/child"
assert_contains "$normalized" "space toggle"
assert_contains "$normalized" "g top"
assert_contains "$normalized" "G bottom"
assert_contains "$normalized" "pgup page up"
assert_contains "$normalized" "pgdn page down"
assert_contains "$normalized" "C-u half up"
assert_contains "$normalized" "C-d half down"
assert_contains "$normalized" "H parent"
assert_contains "$normalized" "J next sibling"
assert_contains "$normalized" "K prev sibling"
assert_contains "$normalized" "/ search"
assert_contains "$normalized" "n next match"
assert_contains "$normalized" "N previous match"
assert_contains "$normalized" "pp print pretty value"
assert_contains "$normalized" "ps print string contents"
assert_contains "$normalized" "pq print jq query"
assert_contains "$normalized" "? toggle help"
assert_contains "$normalized" "q quit"

printf '%s\n' "$pane"
quit_app
trap - EXIT
