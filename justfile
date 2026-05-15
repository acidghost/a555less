program := 'a555less'

version := 'SNAPSHOT-'+`git describe --tags --always --dirty 2>/dev/null || printf 'unknown'`
commit_sha := `(git rev-parse HEAD 2>/dev/null || printf 'unknown') | tr -d '\n'`
build_time := `date -u '+%Y-%m-%d_%H:%M:%S'`

ldflags := '-s -w -X main.buildVersion='+version \
        +' -X main.buildCommit='+commit_sha \
        +' -X main.buildDate='+build_time

goos := if os() == 'macos' { 'darwin' } else { os() }
goarch := if arch() == 'aarch64' { 'arm64' } else if arch() == 'x86_64' { 'amd64' } else { arch() }

shell_files := shell("find . -name '*.sh' -not -path '*/vendor/*' | tr '\\n' ' '")

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

watch *args:
    watchexec --restart -- 'date; go run . {{args}}'

# Unit tests for parser/rendering/navigation helpers.
test:
    go test ./...

# Subcommand for e2e testing with tmux. Run `just build` first.
e2e:
    @just e2e/

# Vendor dependencies and tidy module
vendor:
    go mod tidy
    go mod vendor

# Format Go files
fmt:
    go fmt ./...

# Check Go linter
lint:
    golangci-lint run

# Format shell files
fmt-sh:
    shfmt --write {{shell_files}}

# Check shell linter
lint-sh:
    shfmt --diff {{shell_files}}
    shellcheck --external-sources {{shell_files}}

# Install into GOBIN
install: build
    cp -v './build/{{program}}-{{goos}}-{{goarch}}' "$(go env GOBIN)/{{program}}"

# Clean up build output
clean:
    rm -rf build
