# Balde

A budget manager CLI for the AI Agents era. Open-source, local-first, bucket budgeting method.

## Table of Contents

- [Overview](#overview)
- [Requirements](#requirements)
- [Getting Started](#getting-started)
- [AI Agent Integration](#ai-agent-integration)
- [Development](#development)
- [Architecture](#architecture)
- [Contributing](#contributing)
- [Usage](#usage)
- [License](#license)

## Overview

Balde implements the [bucket budgeting method](https://www.budgetwithbuckets.com/guide/everything/) in a CLI tool designed for both direct use and AI agent integration. All data stays local in SQLite. Amounts are stored as integer cents — no floats.

## Requirements

- Go 1.26+
- Make (optional, for convenience commands)

## Getting Started

```sh
git clone https://github.com/egermano/balde.git
cd balde
go mod download
make test
```

## AI Agent Integration

Balde is designed to work seamlessly with AI agents through its JSON output mode. The skill at `.agents/skills/balde/` provides comprehensive instructions for agents to manage budgets.

**Key features for agents:**
- All commands support `--json` output for structured data
- Currency formatting via config-driven rules (symbol, separators)
- Clear error messages with structured error codes
- TDD-ready architecture with testable core logic

**Example agent workflow:**

```bash
# Initialize non-interactive budget
echo "N" | balde init --dir /path/to/budget

# Get complete budget state as JSON
balde status --json

# Parse with jq for structured data
echo "$(balde status --json)" | jq '.accounts // []'

# Navigate and execute commands
cd /path/to/budget && balde allocate 50000 1
```

**Agent skill capabilities:**
- Budget initialization (with encryption support)
- Account and bucket management
- Transaction recording with categorization
- "Make it rain" allocation workflow
- Real-time budget status monitoring

See `.agents/skills/balde/SKILL.md` for detailed agent instructions and workflows.

## Development

```sh
make build         # build CLI binary to bin/balde
make test          # run all tests verbose
make test-short    # run tests without -v
make lint          # go vet + gofmt
make tidy          # go mod tidy
```

### Run a single test

```sh
go test ./core/ -run TestRain -v
go test ./core/ -v -count=1
```

## Architecture

```
core/     Business logic (Budget, Account, Bucket, Transaction). Zero I/O deps.
store/    SQLite persistence (Store interface). Uses modernc.org/sqlite (no CGO).
import/   File parsers (CSV now, OFX/QIF later).
cli/      Cobra CLI layer. Supports --json output.
i18n/     Locale dictionaries (JSON). MVP = English only.
report/   Output formatting (tables, JSON, Markdown).
```

`core` has zero imports from `cli`, `store`, or any I/O. It is a pure Go library.

## Contributing

1. Fork the repo
2. Create a feature branch (`feat/your-feature` or `fix/your-fix`)
3. Write tests first (TDD: red → green → refactor)
4. Ensure `make lint && make test` passes
5. Open a Pull Request against `main`

All changes to `main` require a PR. Max 8 buckets — simplicity is a design goal.

## Usage

```sh
balde init                                    # create budget DB, set frequency & currency
balde account add checking checking 100000    # add account (amount in cents)
balde bucket add housing 50000                # add bucket with monthly target
balde transaction add -50000 "rent" acc1 bkt1  # record expense
balde allocate 50000 housing                  # allocate rain to bucket
balde rain                                     # show unallocated money
balde status --json                            # full budget snapshot
```

## License

MIT
