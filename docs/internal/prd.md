# PRD: Balde — MVP (Minimum Viable Product)

## Problem Statement

Personal finance tools are either too complex (spreadsheets with 20+ categories), too simple (no methodology), or locked behind subscriptions and cloud services that handle sensitive financial data. There is no open-source CLI tool that strictly implements the bucket budgeting method while being designed from the ground up for AI agent integration. People who want to manage their money using the bucket/envelope method currently have to either use the paid Budget with Buckets desktop app ($64), manually track buckets in spreadsheets, or use general-purpose budgeting apps that don't enforce the methodology.

## Solution

Balde is a free, open-source CLI tool written in Go that implements the bucket budgeting method. It stores all data locally in SQLite, imports transactions from bank/credit card CSV exports, and is designed to be operated both directly by users and programmatically by AI agents (like Opencode). The MVP delivers the core loop: receive income → "make it rain" (distribute to buckets) → spend from buckets → review and adjust.

## User Stories

### Setup & Onboarding

1. As a new user, I want to run `balde init` to create a new budget database with 6 default buckets, so that I can start tracking my finances immediately
2. As a new user, I want to be prompted to set my global frequency (weekly/fortnightly/monthly), so that all bucket allocations match my pay cycle
3. As a new user, I want the default buckets to be: financial freedom, fixed costs, pleasures, comfort, knowledge, goals, so that I can start with a proven bucket budgeting structure
4. As a new user, I want to be able to add up to 2 custom buckets (free tier limit), so that I can personalize my budget without overcomplicating it
5. As a new user, I want to be prompted to configure my currency format (decimal separator, thousands separator, currency symbol) during init, so that amounts display correctly for my locale
6. As a new user, I want to run `balde config set locale pt-BR` to change the interface language, so that I can use Balde in my preferred language

### Accounts

5. As a user, I want to add a bank account with a name, type (checking, savings, credit), and starting balance, so that Balde knows where my money lives
6. As a user, I want to view all my accounts and their current balances, so that I can verify my data is correct
7. As a user, I want to close an account when I stop using it in real life, so that my budget stays clean without losing transaction history

### Buckets

8. As a user, I want to create a bucket with a name and monthly allocation target, so that I can allocate money to a spending/savings purpose
9. As a user, I want to see all my buckets with their name, allocation target, current balance, and how much of the target is filled, so that I know my financial position at a glance
10. As a user, I want the system to warn me when I have more than 8 buckets, so that I don't fall into the overcomplication trap that leads to abandoning budgets
11. As a user, I want to edit a bucket's name or target, so that I can adjust as my life changes
12. As a user, I want to delete a bucket, so that I can remove ones I no longer need

### Rain (Income & To Be Budgeted)

13. As a user, I want to see my current "rain" (money available to budget), so that I know how much I can distribute to buckets
14. As a user, I want the rain to be calculated as: total account balances minus total bucket balances, so that it accurately reflects unallocated money
15. As a user, I want to run `balde rain` and see the rain amount plus a list of my buckets with their current fill status, so that I can make informed allocation decisions
16. As a user, I want to allocate a specific amount from rain to a bucket via `balde allocate <amount> <bucket>`, so that each centavo has a job
17. As a user, I want to see the rain drop to zero when all money is allocated, so that I know every centavo is accounted for
18. As a user, I want to record income by adding a positive transaction categorized as "income", so that it feeds into the rain for that month

### Transactions

19. As a user, I want to add a manual transaction with amount, description, account, and bucket, so that I can record spending or income
20. As a user, I want negative amounts to represent expenses and positive amounts to represent income, so that the convention is intuitive
21. As a user, I want to view all transactions with date, description, amount, account, and bucket, so that I can audit my financial activity
22. As a user, I want to filter transactions by bucket, account, or date range, so that I can find specific transactions quickly
23. As a user, I want every expense transaction to reduce the balance of its assigned bucket, so that bucket balances stay accurate in real time

### CSV Import

24. As a user, I want to import transactions from a CSV file exported by my bank or credit card, so that I don't have to enter every transaction manually
25. As a user, I want to specify column mapping during import (which column is date, description, amount, etc.), so that Balde can handle different bank formats
26. As a user, I want imported transactions to be flagged as "uncategorized" if no bucket mapping is provided, so that I can categorize them later
27. As a user, I want Balde to skip duplicate imports (based on date + amount + description), so that I don't get double entries if I import the same file twice

### Negative Balances & Warnings

28. As a user, I want to see a visual warning when a bucket goes negative, so that I know I've overspent in that category
29. As a user, I want to see a warning when my rain goes negative (total allocated exceeds total available), so that I know I've over-budgeted

### Currency & Number Formatting

30. As a user, I want to configure whether my amounts use `.` or `,` as the decimal separator, so that Balde matches my local convention (e.g., 1.234,56 vs 1,234.56)
31. As a user, I want to configure my currency symbol (R$, $, €, £, etc.), so that amounts display with the correct symbol
32. As a user, I want to enter amounts in my configured format and have Balde parse them correctly, so that I don't have to think about format conversion
33. As a user, I want all displayed amounts to follow my configured formatting, so that output is consistently readable for me

### Internationalization (i18n)

34. As a user, I want the MVP interface to be in English, so that it reaches the widest audience
35. As a user, I want Balde to support loading different language dictionaries, so that future versions can display the interface in my language
36. As a developer, I want all user-facing strings to be externalized in locale files (not hardcoded), so that adding new translations is straightforward

### AI-Friendly Output

37. As an AI agent, I want all `view` commands to accept a `--json` flag that outputs structured JSON, so that I can parse the data programmatically
38. As an AI agent, I want command arguments to be predictable and positional, so that I can construct commands without ambiguity
39. As an AI agent, I want error messages to be machine-readable (structured format with error code and description), so that I can handle failures gracefully
40. As an AI agent, I want a `balde status` command that outputs a JSON snapshot of the entire budget state (accounts, buckets with balances, rain, recent transactions), so that I can understand the user's financial position in a single call

### AI Harness Integration

41. As a developer integrating Balde with an AI harness, I want the CLI to have a modular architecture with clear separation between core logic and CLI presentation, so that I can reuse the core logic as a library
42. As a developer, I want the CLI to support structured logging with a `--quiet` flag for agent-driven usage, so that only the requested data is output without decorative text

## Implementation Decisions

### Module Architecture

The project is organized into the following deep modules, each with a narrow public interface:

- **`core`** — Pure business logic for buckets, accounts, transactions, and rain calculation. No knowledge of CLI, SQLite, or file formats. This is the most testable module. Key interfaces:
  - `Budget` — aggregate root that owns accounts, buckets, transactions, and provides rain calculation
  - `Account` — represents a real-world financial account (checking, savings, credit)
  - `Bucket` — represents a budgeting envelope with a name, target amount, and current balance
  - `Transaction` — represents a financial movement linked to an account and a bucket

- **`store`** — SQLite persistence layer. Implements a `Store` interface used by `core` to persist and retrieve data. The interface allows swapping implementations (e.g., in-memory for tests, future cloud sync). Uses `modernc.org/sqlite` (pure Go, no CGO) for zero-dependency binaries.

- **`import`** — File parsing module. Each format (CSV, future OFX/QIF) is a separate parser behind a common `Importer` interface. CSV import in MVP; others post-MVP.

- **`cli`** — Thin presentation layer using `cobra`. Builds commands, parses arguments, calls `core`, formats output. Supports `--json` flag on all read commands. Handles currency formatting based on user config.

- **`i18n`** — Internationalization module. Loads locale dictionaries from JSON files. MVP ships with `en` only. Architecture supports adding new locales without code changes (just add a new JSON file). Uses `go-i18n` or similar. All user-facing strings (help text, error messages, bucket suggestions, warnings) go through the i18n layer.

- **`report`** — Report generation module. Formats data from `core` into tables, JSON, or Markdown. Uses `i18n` for labels and `config` for currency formatting. MVP scope is limited to the `view` commands and `status`.

### Data Model (SQLite Schema)

Key design decisions:

- **Amounts stored as integers (cents)** — Following the Budget with Buckets convention, avoiding floating-point arithmetic issues. Display layer converts to decimal using user-configured separators.
- **Currency formatting is config-driven, not locale-assumed** — The user explicitly sets: decimal separator (`.` or `,`), thousands separator (`,` or `.`), and currency symbol (R$, $, €, etc.). These are stored in the project config. This avoids incorrect assumptions (e.g., a Brazilian user on an English-language OS still needs `,` as decimal separator).
- **Input parsing respects config** — When a user types `1.234,56` or `1,234.56`, the input parser uses the configured separators to correctly interpret the amount. Internally stored as `123456` cents.
- **Global frequency stored in config** — A single setting (weekly/fortnightly/monthly) that all bucket targets are expressed in. Conversion utilities for pay cycles (monthly ÷ 2.17 for fortnightly, monthly ÷ 4.33 for weekly).
- **Locale and currency config stored in project config** — Settings include: `locale` (default `en`), `currency_symbol` (default `$`), `decimal_separator` (default `.`), `thousands_separator` (default `,`). Config is per-project (stored alongside the SQLite database), not global, so a user can have different budgets with different currencies.
- **Rain is computed, not stored** — Rain = sum of all account balances minus sum of all bucket balances. This is always recalculated, never persisted, ensuring consistency.
- **Transactions are immutable once categorized** — A transaction can be edited before categorization, but once assigned to a bucket, only the bucket assignment can be changed (which triggers a balance recalculation).

### Rain Calculation

Rain follows the "rain barrel" metaphor from Budget with Buckets:

```
rain = (sum of all account balances) - (sum of all bucket balances)
```

- Income transactions increase account balances → rain goes up
- Allocating to a bucket increases bucket balances → rain goes down
- Expense transactions decrease bucket balances (when categorized) → rain is unaffected
- Positive rain = money available to distribute. Negative rain = over-allocated.

### CSV Import Pipeline

1. Read file → parse rows into raw records
2. Map columns using user-provided or default mapping (date, description, amount, maybe account)
3. Deduplicate against existing transactions (date + amount + description match)
4. Create transactions as "uncategorized" (no bucket assignment)
5. User (or agent) assigns buckets afterward via `balde transaction categorize`

### Frequency Conversion Factors

| Source | To Weekly | To Fortnightly | To Monthly |
|--------|-----------|----------------|------------|
| Monthly | ÷ 4.33 | ÷ 2.17 | 1 |
| Fortnightly | × 2 | 1 | × 2.17 |
| Weekly | 1 | ÷ 2 | × 4.33 |

These are used to display bucket targets in the user's chosen frequency regardless of the original billing cycle.

## Testing Decisions

### What makes a good test

Tests should verify **external behavior** (inputs → outputs, state transitions), not implementation details (which internal function was called, how data is structured internally). A test should continue to pass if the internal implementation is refactored, as long as the observable behavior remains the same.

### Modules to test

- **`core`** (highest priority) — This is the business logic heart. Every behavior should be tested:
  - Rain calculation (positive, negative, zero, after income, after allocation, after expense)
  - Bucket balance tracking (allocate, spend, overspend)
  - Account balance tracking (income, expense, transfer between accounts)
  - Frequency conversion accuracy
  - Overspend detection and warnings
  - 8+ bucket warning threshold
  - Duplicate transaction detection
  - Currency amount parsing with different separators (1.234,56 vs 1,234.56)
  - Currency amount formatting for display

- **`store`** — Test the SQLite implementation of the Store interface:
  - CRUD operations for accounts, buckets, transactions
  - Schema migration on init
  - Data integrity (balances stay consistent after operations)

- **`import`** — Test CSV parser with various bank formats:
  - Correct column mapping
  - Edge cases (empty rows, negative amounts, different date formats)
  - Deduplication logic

- **`cli`** — Integration-level tests only:
  - Command parsing and routing
  - `--json` flag produces valid JSON
  - Error messages are structured and readable

- **`i18n`** — Verify locale loading and string resolution:
  - English locale loads and resolves all required keys
  - Missing key falls back gracefully (no crash)
  - Adding a new locale file works without code changes

### Prior art

As a greenfield Go project, tests will follow standard Go conventions (`_test.go` files, table-driven tests). The `core` module can be tested with an in-memory store implementation, making tests fast and isolated.

## Out of Scope (for MVP)

- Locale dictionaries beyond English (pt-BR, es, fr, etc.) — architecture supports them but only `en` ships in MVP
- Multi-currency budgets (one budget, multiple currencies) — MVP assumes one currency per budget
- OFX and QIF import formats (post-MVP, Phase 1)
- Tag/taxonomy system for transactions (post-MVP, Phase 1)
- Rule-based auto-categorization (post-MVP, Phase 1)
- Debt accounts and debt payment buckets (Phase 2)
- Recurring transaction scheduling (Phase 2)
- Savings goals with target dates (Phase 2)
- Reports beyond basic `view` commands (Phase 1)
- Export to CSV/Markdown (Phase 1)
- Transfer between buckets (post-MVP — MVP only supports allocate-from-rain)
- Multi-user or shared budgets (Phase 3)
- Bank API integration / Open Banking (Phase 3)
- Web or desktop UI (Phase 3)
- Plugin system (Phase 3)
- Cloud sync (monetization tier, not MVP)

## Further Notes

### Methodology References

The bucket budgeting method is documented in three primary sources that informed the design:
- [Budget with Buckets — Everything Guide](https://www.budgetwithbuckets.com/guide/everything/) — the original "make it rain" metaphor, 4-step monthly loop, bucket types, debt handling
- [Budget Buckets App — Method Guide](https://budgetbucket.app/guide/budget-buckets-method) — 5-8 bucket guidance, 50/30/20 rule, frequency alignment, common mistakes
- [Spaceship — How to Bucket Your Money](https://www.spaceship.com.au/learn/how-to-bucket-your-money/) — practical examples (Barefoot Investor 3-bucket model), real user setups, auto-transfer patterns, rebalancing advice

### Future Monetization

While the CLI is free and open-source, monetization paths include: hosted sync service, advanced AI categorization, web/mobile UI, consulting/onboarding, and educational content. See `monetization.md` for full details.

### AI Harness Design Principle

The single most important architectural decision is: **`core` has zero dependencies on `cli` or `store`**. This means:
- An AI agent skill can import `core` as a Go library and programmatically manage a budget
- Tests can use an in-memory store
- A future web API can reuse `core` without changes
- The CLI is just one possible interface to the same business logic

### i18n & Currency Design Principle

All user-facing strings live in the `i18n` module behind locale JSON files. The `core` module never produces human-readable text — it returns structured data. The `cli` and `report` modules use `i18n` to translate data into the user's language. Similarly, all amount formatting (decimal separator, thousands separator, currency symbol) is handled exclusively at the presentation layer. `core` works only in integer cents with no knowledge of display conventions.
