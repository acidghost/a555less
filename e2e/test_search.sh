#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=e2e/lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

setup_args "$@"
start_app search
trap cleanup_app EXIT

wait_for "input"
send_keys "/"
send_keys "dup"
wait_for "/dup"
send_keys Enter
wait_for "[1/2]"

pane_ansi=$(capture_ansi)
assert_contains "$pane_ansi" "48;5;220"
assert_contains "$pane_ansi" "48;5;229"

# Manual cursor movement keeps the results highlighted.
send_keys g
wait_for "1/22"
pane_ansi=$(capture_ansi)
assert_contains "$pane_ansi" "48;5;220"
assert_contains "$pane_ansi" "48;5;229"

# Next is relative to the manual cursor, not the last jumped-to result.
send_keys n
wait_for "4/22"
pane=$(capture)
assert_contains "$pane" "[1/2]"

# Previous is also relative to a manually positioned cursor.
send_keys n
wait_for "[2/2]"
send_keys G
wait_for "22/22"
send_keys N
wait_for "5/22"
pane=$(capture)
assert_contains "$pane" "[2/2]"
pane_ansi=$(capture_ansi)
assert_contains "$pane_ansi" "48;5;220"
assert_contains "$pane_ansi" "48;5;229"

pane=$(capture)
assert_contains "$pane" "dup: \"first\""
assert_contains "$pane" "dup: \"second\""

# Searches ignore case unless the query ends in /s.
send_keys "/"
send_keys "DUP"
wait_for "/DUP"
send_keys Enter
wait_for "[2/2]"
send_keys "/"
send_keys "DUP/s"
wait_for "/DUP/s"
send_keys Enter
wait_for "[0/0]"
pane_ansi=$(capture_ansi)
assert_not_contains "$pane_ansi" "48;5;220"
assert_not_contains "$pane_ansi" "48;5;229"

pane=$(capture)
printf '%s\n' "$pane"

quit_app
trap - EXIT
