program := 'a555less'

version := 'SNAPSHOT-'+`git describe --tags --always --dirty 2>/dev/null || printf 'unknown'`
commit_sha := `(git rev-parse HEAD 2>/dev/null || printf 'unknown') | tr -d '\n'`
build_time := `date -u '+%Y-%m-%d_%H:%M:%S'`

ldflags := '-s -w -X main.buildVersion='+version \
        +' -X main.buildCommit='+commit_sha \
        +' -X main.buildDate='+build_time

goos := if os() == 'macos' { 'darwin' } else { os() }
goarch := if arch() == 'aarch64' { 'arm64' } else if arch() == 'x86_64' { 'amd64' } else { arch() }

smoke_json := 'testdata/basic.json'
tmux_session := program+'-smoke'
tmux_width := '100'
tmux_height := '30'
tmux_bin := 'build/'+program+'-tmux'

_help:
    @just --list

# Cross-compile
build-all: (build 'darwin' 'arm64') (build 'linux' 'arm64') (build 'linux' 'amd64')

build os=goos arch=goarch: _build-dir
    CGO_ENABLED=0 GOOS={{os}} GOARCH={{arch}} \
        go build \
            -ldflags '{{ldflags}}' \
            -o build/{{program}}-{{os}}-{{arch}}

_build-dir:
    mkdir -p build

# Fast development run without building build/ artifact.
run *args:
    go run . {{args}}

# Run with stdin input, matching: cat file.json | a555less
dev-stdin file=smoke_json:
    go run . - < {{file}}

# Unit tests for parser/rendering/navigation helpers.
test:
    go test ./...

# Vendor dependencies and tidy module
vendor:
    go mod tidy
    go mod vendor

# Format Go files
fmt:
    go fmt ./...

# Check linter
lint:
    golangci-lint run

# Start the app in a detached fixed-size tmux session for manual TUI testing.
tmux-dev file=smoke_json session=tmux_session width=tmux_width height=tmux_height:
    @tmux kill-session -t "{{session}}" 2>/dev/null || true
    tmux new-session -d -s "{{session}}" -x "{{width}}" -y "{{height}}" -c "$PWD" "go run . {{file}}"
    @echo "Started {{session}}. Attach with: tmux attach -t {{session}}"
    @echo "Capture with: just tmux-capture"

# Capture the current TUI pane as plain text.
tmux-capture session=tmux_session:
    tmux capture-pane -p -t "{{session}}"

# Capture the current TUI pane including ANSI escapes, useful for style checks.
tmux-capture-ansi session=tmux_session:
    tmux capture-pane -e -p -t "{{session}}"

# Send keystrokes to the tmux TUI session, e.g. `just tmux-keys j Space Enter`.
tmux-keys session=tmux_session *keys:
    tmux send-keys -t "{{session}}" {{keys}}

# Stop the tmux TUI session.
tmux-kill session=tmux_session:
    tmux kill-session -t "{{session}}" 2>/dev/null || true

# Scriptable smoke check for core TUI render/navigation. Intended for use once the Bubble Tea model exists.
tmux-smoke file=smoke_json session=tmux_session width=tmux_width height=tmux_height:
    #!/usr/bin/env bash
    set -euo pipefail
    target="{{session}}"
    tmux kill-session -t "{{session}}" 2>/dev/null || true
    trap 'tmux kill-session -t "{{session}}" 2>/dev/null || true' EXIT

    tmux new-session -d -s "{{session}}" -x "{{width}}" -y "{{height}}" -c "$PWD" "go run . {{file}}"

    deadline=$((SECONDS + 5))
    pane=""
    until pane=$(tmux capture-pane -p -t "$target" 2>/dev/null) && grep -q 'users' <<<"$pane"; do
      if (( SECONDS >= deadline )); then
        echo "Timed out waiting for TUI tree to render" >&2
        printf '%s\n' "$pane" >&2
        exit 1
      fi
      sleep 0.1
    done

    tmux send-keys -t "$target" j Space Enter Right Left PageDown C-u
    sleep 0.1
    tmux capture-pane -p -t "$target"
    tmux send-keys -t "$target" q

# Install into GOBIN
install: build
    cp -v './build/{{program}}-{{goos}}-{{goarch}}' "$(go env GOBIN)/{{program}}"

# Clean up build output
clean:
    rm -rf build
