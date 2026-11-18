# Balde

A budget manager CLI for the AI Agents era. Open-source, local-first, bucket budgeting method.

## Table of Contents

- [Overview](#overview)
- [Requirements](#requirements)
- [Getting Started](#getting-started)
- [AI Agent Integration](#ai-agent-integration)
- [Data privacy: local-first & LLMs](#data-privacy-local-first--llms)
- [Development](#development)
- [Architecture](#architecture)
- [Contributing](#contributing)
- [Maintainer & Support](#maintainer--support)
- [Usage](#usage)
- [License](#license)

## Overview

Balde implements the [bucket budgeting method](https://www.budgetwithbuckets.com/guide/everything/) in a CLI tool designed for both direct use and AI agent integration. All data stays local in SQLite. Amounts are stored as integer cents — no floats.

> **Status: alpha.** The core budget flow (accounts, buckets, transactions, allocation, encryption) is stable and tested, but the CLI surface may still change. The plugin system is experimental. See [Releases](https://github.com/egermano/balde/releases) for the current version.

## Installation

Install or upgrade `balde` with a single command.

#### Linux / macOS

```sh
curl -fsSL https://raw.githubusercontent.com/egermano/balde/main/install.sh | bash
```

#### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/egermano/balde/main/install.ps1 | iex
```

The script detects your platform, verifies the download against the release checksum (sha256), and installs the binary — leaving nothing behind:

- Linux/macOS: `~/.local/bin`
- Windows: `%USERPROFILE%\bin`

It is safe to re-run: if the latest version is already installed, it does nothing; otherwise it upgrades in place.

| Environment variable | Description |
|----------------------|-------------|
| `BALDE_VERSION` | Pin a specific release (e.g. `v0.1.0-alpha.3`) |
| `BALDE_INSTALL_DIR` | Custom install directory (must be on your `PATH`) |
| `BALDE_FORCE=1` | Reinstall even if the latest version is already installed |

If the install directory is not on your `PATH`, the script prints instructions to add it.

### Build from source

```sh
go install github.com/egermano/balde@latest
```

Requires Go 1.26+.

## Requirements

- Go 1.26+
- Make (optional, for convenience commands)

## Getting Started

The fastest way to see the whole method in action — a scripted flow with synthetic data:

```sh
git clone https://github.com/egermano/balde.git
cd balde
./examples/quickstart.sh
```

See [examples/](examples/) for the walk-through and expected output. To hack on Balde itself:

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

**Install the skill:**

```bash
npx skills add egermano/balde --skill balde
```

See `.agents/skills/balde/SKILL.md` for detailed agent instructions and workflows.

## Data privacy: local-first & LLMs

Balde itself is local-first: your budget lives in a SQLite file on your disk, and the CLI talks to no network service. **That is not a privacy guarantee against the tools you put around it.** If you let an AI agent operate your budget — with this skill or any other — the agent's tools and provider configuration decide what data is read and sent. Prompts, command outputs and `--json` snapshots can all reach the model provider's servers.

Practical consequences:

- Use synthetic or non-sensitive data when experimenting or writing demos
- Check your agent tooling's data-retention and training policies for anything you consider private
- Encrypted budgets (`balde encrypt`) protect the file at rest, not the data once an agent has unlocked and read it

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
core/      Domain model (Budget, Account, Bucket, Transaction, rain math). Pure Go, zero I/O.
store/     SQLite persistence + AES-256-GCM encryption and session handling (pure Go, no CGO).
app/       Application services: typed operations with the stable JSON wire schema (DTOs).
cli/       Cobra CLI. Parses args, calls app, formats output. Supports --json.
plugin/    Plugin runtime: manifest validation, lockfile pinning, verified subprocess execution.
plugins/   Bundled plugins (vacation planner) — examples of the plugin format.
version/   Build metadata injected at compile time.
docs/      Plugin docs (docs/plugins) and internal planning notes (docs/internal).
.agents/   The balde agent skill (skills/balde) installed by `npx skills add`.
scripts/   Installer smoke tests used by CI.
```

Dependency rule: `core` is a pure library — no SQLite, no filesystem, no user-facing strings. `store` implements the `Store` interface defined in `core`. `app` sits between `cli` and `core`, exposing DTOs so core types never cross the wire.

**Planned, not built yet:** CSV import, i18n locales, report formatting. The internal roadmap lives in [`docs/internal/`](docs/internal/).

## Contributing

1. Fork the repo
2. Create a feature branch (`feat/your-feature` or `fix/your-fix`)
3. Write tests first (TDD: red → green → refactor)
4. Ensure `make lint && make test` passes
5. Open a Pull Request against `main`

All changes to `main` require a PR. Max 8 buckets — simplicity is a design goal.

## Maintainer & Support

Balde is built and maintained by [Bruno Germano](https://github.com/egermano).

- 🐛 **Bugs and feature requests** — [open an issue](https://github.com/egermano/balde/issues)
- 🔒 **Security reports** — see [SECURITY.md](SECURITY.md) (private reporting only, please)
- 💬 **Questions** — [open an issue](https://github.com/egermano/balde/issues) with your question; there are no separate forums yet

## Usage

A real session — every command below runs exactly as shown (amounts in **integer cents**; negative = expense; buckets are referenced by **ID**, see `balde view buckets`):

```sh
balde init                                    # create budget: 6 default buckets, currency, encryption
balde account add checking checking 100000    # account with starting balance → id 1
balde transaction add 250000 "salary" 1 1     # income to account 1, bucket 1
balde transaction add -50000 "rent" 1 2       # expense from account 1, bucket 2
balde allocate 50000 2 --force                # move rain (unallocated) into bucket 2
balde rain                                    # → Rain (unallocated): 50000 cents
balde view buckets                            # id, target, balance, fill %
balde status --json                           # full snapshot — for scripts and agents
```

`balde allocate` asks for confirmation when the amount exceeds available rain; `--force` skips it. Most commands run in the current directory; only `init` and `unlock` accept `--dir`.

Encryption (optional, set at init or added later):

```sh
balde encrypt                                 # migrate a plain budget to AES-256-GCM
balde unlock                                  # decrypt for the session (password prompt)
balde lock                                    # invalidate the session
```

Plugins (experimental — opt in with `--experimental` or `BALDE_EXPERIMENTAL=1`):

```sh
balde plugin install github.com/egermano/balde --path plugins/vacation
BALDE_EXPERIMENTAL=1 balde plugin list
```

See [docs/plugins](docs/plugins/README.md) for the plugin format, protocol and bundled plugins.

## License

MIT — see [LICENSE](LICENSE).
