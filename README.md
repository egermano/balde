# Balde

[![CI](https://github.com/egermano/balde/actions/workflows/ci.yml/badge.svg)](https://github.com/egermano/balde/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/egermano/balde?include_prereleases&sort=semver)](https://github.com/egermano/balde/releases)
[![License: MIT](https://img.shields.io/github/license/egermano/balde)](LICENSE)

A local-first budget manager that implements the [bucket (envelope) method](https://www.budgetwithbuckets.com/guide/everything/) in a single SQLite file — usable straight from the terminal, and safe for AI agents to operate on your behalf.

## Table of Contents

- [Overview](#overview)
- [Why Balde](#why-balde)
- [Installation](#installation)
- [Usage](#usage)
- [Backup & data safety](#backup--data-safety)
- [AI Agent Integration](#ai-agent-integration)
- [Demo](#demo)
- [Data privacy: local-first & LLMs](#data-privacy-local-first--llms)
- [Development](#development)
- [Architecture](#architecture)
- [Contributing](#contributing)
- [Maintainer & Support](#maintainer--support)
- [Roadmap](#roadmap)
- [License](#license)

## Overview

Balde implements the [bucket budgeting method](https://www.budgetwithbuckets.com/guide/everything/) in a CLI tool designed for both direct use and AI agent integration. All data stays local in SQLite. Amounts are stored as integer cents — no floats.

> **Status: alpha.** The core budget flow (accounts, buckets, transactions, allocation, encryption) is stable and tested, but the CLI surface may still change. The plugin system is experimental. See [Releases](https://github.com/egermano/balde/releases) for the current version.

## Why Balde

**Compared to a spreadsheet** — formulas rot quietly. Balde encodes the method as invariants: rain is *computed* from accounts minus buckets (never stored, never stale), over-allocation warns, transactions keep history. Everything is a command or `--json` away — scriptable, diffable, agent-readable.

**Compared to a budgeting app/SaaS** — your financial life doesn't need an account and a cloud. The whole budget is two files in a directory you own (`balde.db` + `balde.json`), offline forever, [encryptable at rest](docs/backup.md) with AES-256-GCM. No telemetry, no sync, no subscription.

**Compared to a generic ledger CLI** — double-entry ledgers are powerful but unforgiving. Balde is deliberately narrow: buckets, targets, rain, fill percentages. The "make it rain" loop answers the actual question — *how much money do I still need to put somewhere?* — in one command.

**And it works with zero AI.** Every command is a plain shell command. The agent layer is optional: a [documented skill](docs/agent-demo.md) teaches an agent to run the same CLI you would — resolving names to IDs, checking rain before allocating, formatting cents into your currency. You get automation without giving up the manual path (or vice versa).

Balde is probably **not** for you if you need bank sync, multi-user budgets, charts or mobile apps — those are design boundaries, not gaps; see the [roadmap](ROADMAP.md).

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
| `BALDE_VERSION` | Pin a specific release (e.g. `v0.1.0-alpha.4`) |
| `BALDE_INSTALL_DIR` | Custom install directory (must be on your `PATH`) |
| `BALDE_FORCE=1` | Reinstall even if the latest version is already installed |

If the install directory is not on your `PATH`, the script prints instructions to add it.

### Build from source

```sh
go install github.com/egermano/balde@latest
```

Requires Go 1.26+.

## Requirements

Nothing, if you install the prebuilt binary — it's static (pure Go, no CGO). Go 1.26+ and Make are only needed to build from source or hack on Balde.

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

## Backup & data safety

A budget is two files (`balde.db` + `balde.json`) in a directory you own — back them up together. `balde transaction delete` is a hard delete that rolls back balances; `balde bucket delete` archives instead of destroying. A forgotten encryption password is unrecoverable **by design**.

Details, restore steps and the full destructive-operations table: [docs/backup.md](docs/backup.md). A built-in `balde backup` command is on the [roadmap](ROADMAP.md).

## AI Agent Integration

Balde is designed to work seamlessly with AI agents through its JSON output mode. The skill at `.agents/skills/balde/` provides comprehensive instructions for agents to manage budgets.

**Key features for agents:**
- All commands support `--json` output for structured data
- Deterministic, parseable errors (`Error: <operation>: <cause>`)
- Safety rails baked into the CLI: allocation warns when it exceeds rain, destructive commands confirm
- Currency formatting via config-driven rules (symbol, separators)

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

## Demo

**Terminal walkthrough** — recorded from a real session with synthetic data:

![balde quickstart demo](demo/balde-quickstart.gif)

The same recording is available as an
[asciinema cast](demo/balde-quickstart.cast) — playable at full fidelity
straight from this repository, no cloning or upload needed:

```sh
asciinema play https://raw.githubusercontent.com/egermano/balde/main/demo/balde-quickstart.cast
```

A textual equivalent with expected output lives in [examples/](examples/).

**Agent integration** — a real OpenCode session using the balde skill to check rain, allocate and report status: [docs/agent-demo.md](docs/agent-demo.md).

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

Contributions follow a test-first workflow and a strict dependency rule — the short version: fork, branch, red → green → refactor, `make lint && make test-short`, PR against `main`.

Full guide: [CONTRIBUTING.md](CONTRIBUTING.md) · Conduct: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) · What's next: [ROADMAP.md](ROADMAP.md)

All changes to `main` require a PR. Max 8 buckets — simplicity is a design goal.

## Maintainer & Support

Balde is built and maintained by [Bruno Germano](https://github.com/egermano).

- 🐛 **Bugs and feature requests** — [open an issue](https://github.com/egermano/balde/issues)
- 🔒 **Security reports** — see [SECURITY.md](SECURITY.md) (private reporting only, please)
- 💬 **Questions** — [open an issue](https://github.com/egermano/balde/issues) with your question; there are no separate forums yet

## Roadmap

What's being built next, each item with its problem and a testable done-criterion — CSV import, encrypted-migration integrity, backup & restore, pt-BR locale, and the road to v0.1.0 stable: [ROADMAP.md](ROADMAP.md)

## License

MIT — see [LICENSE](LICENSE).
