#!/usr/bin/env bash

# Shared helpers for tmux-based end-to-end tests.

PROJECT_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
DEFAULT_BIN="$PROJECT_ROOT/build/a555less-$(go env GOOS)-$(go env GOARCH)"
BIN=${BIN:-$DEFAULT_BIN}

E2E_FILE=${E2E_FILE:-testdata/basic.json}
E2E_SESSION_BASE=${E2E_SESSION_BASE:-a555less-e2e}
E2E_WIDTH=${E2E_WIDTH:-100}
E2E_HEIGHT=${E2E_HEIGHT:-30}
E2E_SESSION=""
E2E_TARGET=""

setup_args() {
    E2E_FILE=${1:-$E2E_FILE}
    E2E_SESSION_BASE=${2:-$E2E_SESSION_BASE}
    E2E_WIDTH=${3:-$E2E_WIDTH}
    E2E_HEIGHT=${4:-$E2E_HEIGHT}

    if [[ ! -x "$BIN" ]]; then
        cat >&2 <<EOF
Missing executable: $BIN
Build before running e2e tests, e.g.:
  just build
Or set BIN to a built binary.
EOF
        exit 1
    fi
}

start_app() {
    local name=$1
    E2E_SESSION="${E2E_SESSION_BASE}-${name}-$$"
    E2E_TARGET="$E2E_SESSION"

    tmux kill-session -t "$E2E_SESSION" 2>/dev/null || true

    local cmd
    printf -v cmd '%q %q' "$BIN" "$E2E_FILE"
    tmux new-session -d -s "$E2E_SESSION" -x "$E2E_WIDTH" -y "$E2E_HEIGHT" -c "$PROJECT_ROOT" "$cmd"
}

cleanup_app() {
    if [[ -n "$E2E_SESSION" ]]; then
        tmux kill-session -t "$E2E_SESSION" 2>/dev/null || true
    fi
}

capture() {
    tmux capture-pane -p -t "$E2E_TARGET"
}

send_keys() {
    tmux send-keys -t "$E2E_TARGET" "$@"
}

wait_for() {
    local needle=$1
    local timeout=${2:-5}
    local deadline=$((SECONDS + timeout))
    local pane=""

    until pane=$(capture 2>/dev/null) && grep -Fq -- "$needle" <<<"$pane"; do
        if ((SECONDS >= deadline)); then
            echo "Timed out waiting for: $needle" >&2
            printf '%s\n' "$pane" >&2
            exit 1
        fi
        sleep 0.1
    done
}

assert_contains() {
    local haystack=$1
    local needle=$2
    if ! grep -Fq -- "$needle" <<<"$haystack"; then
        echo "Expected capture to contain: $needle" >&2
        printf '%s\n' "$haystack" >&2
        exit 1
    fi
}

assert_not_contains() {
    local haystack=$1
    local needle=$2
    if grep -Fq -- "$needle" <<<"$haystack"; then
        echo "Expected capture not to contain: $needle" >&2
        printf '%s\n' "$haystack" >&2
        exit 1
    fi
}

quit_app() {
    tmux send-keys -t "$E2E_TARGET" q 2>/dev/null || true
    local deadline=$((SECONDS + 5))
    while tmux has-session -t "$E2E_SESSION" 2>/dev/null; do
        if ((SECONDS >= deadline)); then
            echo "Timed out waiting for app to quit" >&2
            capture >&2 || true
            exit 1
        fi
        sleep 0.1
    done
    return 0
}
