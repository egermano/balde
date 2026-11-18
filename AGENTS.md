# Balde — Agent Instructions

## Workflow: XP Pair Programming (You = Driver, User = Guide)

The user sets direction and priorities. You execute. Don't over-explain — do the work, show results, ask only when blocked.

## Mandatory: TDD Mindset

Every feature MUST follow the red-green-refactor cycle:
1. **Red** — write a failing test first
2. **Green** — write the minimum code to pass
3. **Refactor** — clean up while keeping tests green

No production code without a test. No refactoring without passing tests.

## Dev Environment

```sh
make test          # run all tests verbose
make test-short    # run tests without -v
make build         # build CLI binary to bin/balde
make lint          # go vet + gofmt
make tidy          # go mod tidy
go test ./core/ -run TestRain -v   # run single test
go test ./core/ -v -count=1        # run single package
```

- Go 1.26+, module: `github.com/egermano/balde`
- Dependencies: `cobra` (CLI), `modernc.org/sqlite` (pure Go, no CGO)

## Project Context

Budget manager CLI written in **Go** using the **bucket budgeting method**. Key references:
- Budget with Buckets methodology
- 50/30/20 rule as initial suggestion
- "Make it Rain" allocation loop

See `docs/internal/` for full PRD, roadmap, and project spec (internal working notes — may describe planned features).

## Architecture

```
core/     — Business logic (Budget, Account, Bucket, Transaction). Zero deps on CLI/store. Pure Go.
store/    — SQLite persistence + AES-256-GCM encryption (Store interface). Pure Go, no CGO.
app/      — Application services: typed operations with the stable JSON wire schema (DTOs).
cli/      — Cobra CLI layer. Thin — parses args, calls app, formats output. Supports --json.
plugin/   — Plugin runtime: manifest validation, lockfile pinning, verified subprocess execution.
plugins/  — Bundled plugins (vacation planner).
docs/     — Plugin docs (docs/plugins) and internal planning notes (docs/internal).
```

### Dependency rule (violating this breaks the project)

```
cli ──→ app ──→ core ←── store
              ↑
       (core has ZERO imports from cli, app, store, or any I/O)
```

- `core` is a pure Go library — no filesystem, no SQLite, no user-facing strings
- `store` implements a `Store` interface defined in `core` (in-memory mock for tests, SQLite for production)
- `app` wraps core with DTOs so core types never cross the wire; `cli` and the plugin runtime are the only layers that touch user I/O
- An AI agent can import `core` as a standalone library to manage a budget programmatically

### Planned, not built yet (do not reference as existing)

CSV/OFX/QIF import (`import/`), i18n locale dictionaries (`i18n/`), report formatting (`report/`). The internal roadmap lives in `docs/internal/`.

### Core domain types

- `Budget` — aggregate root; owns accounts, buckets, transactions; provides rain calculation
- `Account` — real-world financial account (checking, savings, credit)
- `Bucket` — budgeting envelope with name, target amount, and current balance
- `Transaction` — financial movement linked to one account and one bucket

## Data Model

- Amounts stored as **integer cents** — never float, anywhere
- Currency formatting is config-driven (decimal sep, thousands sep, symbol) — per-project, not locale-assumed
- Rain is **computed, never stored**: `rain = sum(account balances) - sum(bucket balances)`
- Global frequency: weekly/fortnightly/monthly — all targets expressed in this cycle
- Transactions immutable once categorized (only bucket assignment can change after categorization, triggering recalculation)

### Config is per-project, not global

Stored alongside the SQLite DB. A user can have different budgets with different currencies:

| Setting | Default | Notes |
|---------|---------|-------|
| `locale` | `en` | MVP ships `en` only |
| `currency_symbol` | `$` | R$, €, £, etc. |
| `decimal_separator` | `.` |`,` for BR |
| `thousands_separator` | `,` |`.` for BR |
| `frequency` | `monthly` | weekly / fortnightly / monthly |
| `encrypted` | `false` | AES-256-GCM encryption; set by `balde init --password` or `balde encrypt` |

### Frequency conversion factors

Agent will need these when converting bucket targets between cycles:

| Source | To Weekly | To Fortnightly | To Monthly |
|--------|-----------|----------------|------------|
| Monthly | ÷ 4.33 | ÷ 2.17 | 1 |
| Fortnightly | × 2 | 1 | × 2.17 |
| Weekly | 1 | ÷ 2 | × 4.33 |

### Rain calculation rules

```
rain = sum(account balances) - sum(bucket balances)
```

| Action | Account balance | Bucket balance | Rain |
|--------|----------------|----------------|------|
| Record income | ↑ | — | ↑ |
| Allocate to bucket | — | ↑ | ↓ |
| Spend from bucket | — | ↓ | unchanged |
| Over-allocate | — | ↑↑ | goes negative (warn) |

## Commands (CLI surface)

```
balde init [--dir <path>]            # Create budget DB, optionally encrypted
balde unlock [--password <pass>]     # Unlock encrypted DB (or use BALDE_PASSWORD env var)
balde lock                           # Lock encrypted DB by invalidating session
balde encrypt [--password <pass>]    # Migrate plain DB to encrypted
balde account add <name> <type> <balance>
balde bucket add <name> <target>
balde bucket delete <id>
balde transaction add <amount> <desc> <account_id> <bucket_id>
balde allocate <amount> <bucket_id> [--force]
balde rain                           # Show unallocated money
balde view buckets [--json]
balde view transactions [--json]
balde status [--json]                # JSON snapshot of entire budget state
balde plugin install|list|remove     # Experimental (--experimental)
```

Commands other than `init`/`unlock` run from the budget directory.

### Amount input conventions

- Negative = expense, Positive = income
- Input respects configured separators: `1.234,56` or `1,234.56` → internally `123456` cents
- Parsing and formatting happen only at the CLI layer; `core` only sees integer cents

## Testing

### Priority order

1. **`core`** — highest priority, pure logic, no I/O. Use in-memory store. Table-driven tests.
   Must cover: rain calculation (positive/negative/zero), bucket balance tracking, account balance tracking, frequency conversion, overspend detection, 8+ bucket warning, duplicate transaction detection, amount parsing with different separators
2. **`store`** — SQLite CRUD, schema migration, balance integrity after operations, encryption round-trips
3. **`plugin`** — manifest validation, lockfile pinning, subprocess runtime safety
4. **`app`/`cli`** — integration only: command routing, `--json` produces valid JSON, structured error messages

### Testing principles

- Test **external behavior** (inputs → outputs, state transitions), never implementation details
- A test must survive internal refactoring as long as observable behavior is unchanged
- In-memory `Store` implementation lives in `core` test package for isolated, fast tests
- Tests must be environment-independent: no reliance on current date, network, or host VCS state

## Constraints

- Free/opensource version has **6 fixed buckets**: financial freedom, fixed costs, pleasures, comfort, knowledge, goals. Created on `balde init`.
- Max 8 buckets — system must warn if exceeded (2 custom buckets available in free tier).
- MVP language: English only (i18n infrastructure is planned, not built)
- One currency per budget
- CSV import is planned, not built; OFX/QIF post-MVP
- No transfer between buckets (allocate-from-rain only)
- No multi-user, no cloud sync, no web UI
- Plugin system exists but is **experimental** (gated behind `--experimental`)

## CSV Import Pipeline (planned, not built)

1. Read file → parse rows into raw records
2. Map columns (date, description, amount) using user-provided or default mapping
3. Deduplicate against existing transactions (match on date + amount + description)
4. Create transactions as "uncategorized" (no bucket)
5. User/agent assigns buckets via `balde transaction categorize`

## i18n & Currency Design

- `core` never produces human-readable text — it returns structured data
- Currency formatting (symbol, decimal/thousands separators) is config-driven, read from `balde.json`
- Locale dictionaries (`i18n/`) are planned: adding a language should require only a JSON file, zero code changes

## AI-Friendly Output

All `view` and `status` commands support `--json` for structured output. Error messages are plain-text but deterministic (`Error: <operation>: <cause>`), making them parseable by agents.
