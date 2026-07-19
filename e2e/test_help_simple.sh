#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=e2e/lib.sh
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

setup_args "$@"
start_app help-simple
trap cleanup_app EXIT

wait_for "input"
pane=$(capture)
assert_contains "$pane" "k/↑ up • j/↓ down • space toggle • / search • ? toggle help • q quit"
assert_not_contains "$pane" "collapse/parent"
assert_not_contains "$pane" "page up"

printf '%s\n' "$pane"
quit_app
trap - EXIT
