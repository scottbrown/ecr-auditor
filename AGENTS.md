# CLAUDE.md

This file provides guidance to AI coding assistants when working with code in this repository.

## Build and Test Commands

```bash
task build              # Build binary to .build/ecr-auditor
task test               # Run tests with verbose output
task test:coverage      # Run tests with coverage report
task test:coverage:html # Generate HTML coverage report
task fmt                # Format code
task vet                # Run go vet
task all                # Run fmt, vet, test, and build
```

Run a single test:
```bash
go test -v -run TestFunctionName ./...
```

Build output goes to `.build/`, test coverage to `.test/`.

## Architecture

This is a CLI tool for auditing AWS ECR repositories. The package name is `ecrauditor`.

**Core Components:**

- `types.go` - Domain types: `AuditResult`, `Repository`, `Image`, `LifecyclePolicy`. Contains methods like `MedianImagesPerDay()` and aggregate calculations.
- `ecr.go` - AWS ECR client wrapper. `Auditor` uses an `ECRClient` interface for testability. `NewAuditorWithClient()` allows injecting mock clients.
- `cache.go` - JSON file-based caching at `~/.ecr-auditor/cache/{profile}_{region}.json`. `NewCacheWithDir()` allows test directory injection.
- `report.go` - HTML report generation using Go templates with embedded CSS/JS.
- `errors.go` - Sentinel errors (e.g., `ErrCacheNotFound`).

**CLI Layer (`cmd/ecr-auditor/`):**

- `root.go` - Cobra root command with global `--region` (required) and `--profile` flags
- `collect.go` - Fetches ECR data and caches it
- `summary.go` - Displays cached audit summary in terminal
- `report.go` - Generates HTML report from cached data

**Testing Pattern:**

Tests use interface-based mocking without third-party libraries. The `ECRClient` interface in `ecr.go` enables testing without AWS credentials. Create mock structs implementing the interface methods.
