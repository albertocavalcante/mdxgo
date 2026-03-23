# mdxgo — Roslyn-style red-green CST parser for MDX v3

set dotenv-load := false

# Default: list available commands
default:
    @just --list --unsorted

# ═══════════════════════════════════════════════════════════════════════
#  Build & Test
# ═══════════════════════════════════════════════════════════════════════

# Build all packages
[group('go')]
build:
    go build ./...

# Run unit tests
[group('go')]
test:
    go test ./...

# Run tests with race detector
[group('go')]
test-race:
    go test -race -count=1 ./...

# Run tests via gotestsum (CI-identical)
[group('go')]
test-ci:
    go tool gotestsum \
        --format pkgname --junitfile junit.xml \
        -- -race -count=1 -coverprofile=coverage.out -covermode=atomic ./...

# Run tests with coverage profile
[group('go')]
coverage:
    go test -race -count=1 -coverprofile=coverage.out -covermode=atomic ./...
    go tool cover -func=coverage.out | tail -1
    @echo "  (run 'go tool cover -html=coverage.out' for HTML report)"

# ═══════════════════════════════════════════════════════════════════════
#  Bench
# ═══════════════════════════════════════════════════════════════════════

# Run benchmarks
[group('go')]
bench:
    go test -bench=. -benchmem -count=3 ./parser | tee bench.txt

# Compare benchmarks (requires benchstat)
[group('go')]
bench-compare old="bench-old.txt":
    go tool benchstat {{old}} bench.txt

# ═══════════════════════════════════════════════════════════════════════
#  Fuzz
# ═══════════════════════════════════════════════════════════════════════

# Run fuzz tests (30s per target)
[group('fuzz')]
fuzz:
    go test -fuzz='^FuzzRoundTrip$' -fuzztime=30s ./parser
    go test -fuzz='^FuzzRoundTripMDX$' -fuzztime=30s ./parser
    go test -fuzz='^FuzzWidthConsistency$' -fuzztime=15s ./parser
    go test -fuzz='^FuzzBothModes$' -fuzztime=15s ./parser
    go test -fuzz='^FuzzFromCorpus$' -fuzztime=30s ./parser

# ═══════════════════════════════════════════════════════════════════════
#  Lint & Format
# ═══════════════════════════════════════════════════════════════════════

# Run golangci-lint
[group('lint')]
lint:
    go tool golangci-lint run

# Run go vet
[group('lint')]
vet:
    go vet ./...

# Auto-format Go source files
[group('lint')]
fmt:
    gofmt -w .

# Check formatting (exits non-zero if unformatted)
[group('lint')]
[no-exit-message]
fmt-check:
    @test -z "$(gofmt -l .)" || { echo "Unformatted files:"; gofmt -l .; exit 1; }

# Check module tidy
[group('lint')]
[no-exit-message]
tidy-check:
    go mod tidy && git diff --exit-code go.mod go.sum

# Lint GitHub Actions workflows
[group('lint')]
actions-lint:
    cd tools/lint && go tool actionlint ../../.github/workflows/*.yml

# ═══════════════════════════════════════════════════════════════════════
#  CI & Quality Gates
# ═══════════════════════════════════════════════════════════════════════

# Full local CI: build → fmt-check → vet → lint → test
[group('ci')]
ci: build fmt-check vet lint test-race

# Quick sanity check: build + test + lint
[group('ci')]
check: build test lint
