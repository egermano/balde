# Balde CLI Command Reference

Complete documentation for every Balde CLI command, including signatures, input/output formats, edge cases, and examples.

## Command Quick Reference

| Command | JSON Output | Interactive Prompts |
|---------|-------------|-------------------|
| `balde init` | No | Yes (password, encryption) |
| `balde unlock` | No | Yes (password) |
| `balde lock` | No | No |
| `balde encrypt` | No | Yes (password confirmation) |
| `balde account add` | No | No |
| `balde bucket add` | No | No |
| `balde transaction add` | No | No |
| `balde allocate` | No | No |
| `balde rain` | No | No |
| `balde view buckets` | **Yes** | No |
| `balde view transactions` | **Yes** | No |
| `balde status` | **Yes** | No |

---

## balde init

**Purpose:** Initialize a new budget database

**Signature:**
```bash
balde init [--password <pass>] [--dir <path>]
balde init -p <pass> -d <path>
```

**Arguments:** None

**Flags:**
- `--password`, `-p`: Password for encrypted database (optional)
- `--dir`, `-d`: Directory to initialize budget in (optional, default: current directory)

**Prerequisites:**
- Directory must not contain existing `balde.json` or `balde.db`
- If `--dir` provided, must be a valid absolute or relative path

**Creates:**
- `balde.json` — Budget configuration
- `balde.db` — SQLite database (or `balde.db.backup` if encrypting)

**Default buckets (created automatically):**
1. financial freedom
2. fixed costs
3. pleasures
4. comfort
5. knowledge
6. goals

**Plain text output:**
```
Budget initialized.
```

**JSON output:** None

**Examples:**

**Initialize in current directory (plain, unencrypted):**
```bash
balde init
# Creates: balde.json, balde.db
# Prompt: "Encrypt database? (y/n)"
```

**Initialize with encryption:**
```bash
balde init --password mysecretpassword
# Creates: balde.json (encrypted: true), balde.db (encrypted)
# Prompt: "Confirm password:"
```

**Initialize in specific directory:**
```bash
balde init --dir ~/finances
# Creates: ~/finances/balde.json, ~/finances/balde.db
```

**Edge cases:**

| Error | Condition |
|-------|-----------|
| `budget already initialized in this directory` | `balde.json` or `balde.db` already exists |
| `passwords do not match` | Confirmation password doesn't match initial |
| `create directory: <error>` | `--dir` path invalid or permissions denied |
| `create db: <error>` | SQLite initialization failed |

---

## balde unlock

**Purpose:** Unlock an encrypted budget database

**Signature:**
```bash
balde unlock [--password <pass>] [--db <path>]
balde unlock -p <pass> --db ~/finances/balde.db
```

**Arguments:** None

**Flags:**
- `--password`, `-p`: Password for encryption (optional, uses `BALDE_PASSWORD` env var as fallback)
- `--db`: Path to budget database (optional, default: `balde.db`)

**Prerequisites:**
- Database must be encrypted (`encrypted: true` in `balde.json`)
- Must have valid password

**Creates:**
- Session file at `/tmp/balde-session-<sha256hash>` (valid for 30 minutes)

**Plain text output:**
```
Successfully unlocked database. Session valid for 30 minutes.
Session file: /tmp/balde-session-abc123...
```

**JSON output:** None

**Examples:**

**Unlock with password flag:**
```bash
balde unlock --password mysecretpassword
# Session file created
```

**Unlock using environment variable:**
```bash
export BALDE_PASSWORD=mysecretpassword
balde unlock
# Uses BALDE_PASSWORD
```

**Unlock with interactive prompt:**
```bash
balde unlock
# Prompt: "Enter password:"
# Prompt: "Confirm password:"
```

**Unlock specific database:**
```bash
balde unlock --db ~/finances/balde.db
# Uses custom DB path
```

**Edge cases:**

| Error | Condition |
|-------|-----------|
| `database is not encrypted` | Plain database, unlock not needed |
| `password required` | No password provided and no env var |
| `unlock failed: <error>` | Wrong password or crypto error |

---

## balde lock

**Purpose:** Lock an encrypted database by invalidating the session

**Signature:**
```bash
balde lock [--db <path>]
```

**Arguments:** None

**Flags:**
- `--db`: Path to budget database (optional, default: `balde.db`)

**Prerequisites:**
- Database must be encrypted
- Active session must exist

**Creates:** Nothing (removes session file)

**Plain text output:**
```
Database locked. Session invalidated.
```

**JSON output:** None

**Examples:**

**Lock current database:**
```bash
balde lock
# Session file removed
```

**Lock specific database:**
```bash
balde lock --db ~/finances/balde.db
# Locks custom DB
```

**Edge cases:**

| Error | Condition |
|-------|-----------|
| `database is not encrypted` | Plain database, locking irrelevant |
| `no active session` | Already locked or never unlocked |
| `remove session: <error>` | File system permission error |

---

## balde encrypt

**Purpose:** Convert a plain budget database to encrypted

**Signature:**
```bash
balde encrypt [--password <pass>] [--db <path>]
balde encrypt -p <pass> --db ~/finances/balde.db
```

**Arguments:** None

**Flags:**
- `--password`, `-p`: Password for encryption (optional, uses `BALDE_PASSWORD` env var as fallback)
- `--db`: Path to budget database (optional, default: `balde.db`)

**Prerequisites:**
- Database must be plain (not encrypted)
- Must have password for encryption

**Creates:**
- Encrypted `balde.db` (replaces original)
- Backup at `balde.db.backup` (or `balde.db.backup.N` if collision)
- Session file (same as `unlock`)

**Plain text output:**
```
Database encrypted successfully.
Original backed up to: balde.db.backup
Session created for 30 minutes.
```

**JSON output:** None

**Examples:**

**Encrypt current database:**
```bash
balde encrypt --password mysecretpassword
# Prompt: "Confirm password:"
# Creates: balde.db (encrypted), balde.db.backup
```

**Encrypt with environment variable:**
```bash
export BALDE_PASSWORD=mysecretpassword
balde encrypt
# Uses BALDE_PASSWORD
```

**Edge cases:**

| Error | Condition |
|-------|-----------|
| `database is already encrypted` | Already encrypted, no action needed |
| `passwords do not match` | Confirmation password doesn't match |

---

## balde account add

**Purpose:** Add a new financial account

**Signature:**
```bash
balde account add <name> <type> <balance>
```

**Arguments:**
1. `name`: Account name (string, quoted if contains spaces)
2. `type`: Account type (one of: `checking`, `savings`, `credit`)
3. `balance`: Initial balance (integer cents)

**Flags:** None

**Prerequisites:**
- Budget must be initialized
- Database must be unlocked (if encrypted)

**Creates:** New account in database

**Plain text output:**
```
Account created: My Checking (checking) balance=5000000 id=1
```

**JSON output:** None

**Examples:**

**Add checking account:**
```bash
balde account add "My Checking" checking 5000000
# Balance: $5,000.00 (5,000,000 cents)
```

**Add savings account:**
```bash
balde account add "Emergency Fund" savings 10000000
# Balance: $10,000.00
```

**Add credit card (starting at $0):**
```bash
balde account add "Visa" credit 0
# Credit cards start at 0, expenses become negative balance
```

**Edge cases:**

| Error | Condition |
|-------|-----------|
| `invalid balance: <input>` | Not an integer |
| `invalid type: <input>` | Not `checking`, `savings`, or `credit` |

---

## balde bucket add

**Purpose:** Add a new bucket (envelope) for budgeting

**Signature:**
```bash
balde bucket add <name> <target>
```

**Arguments:**
1. `name`: Bucket name (string, quoted if contains spaces)
2. `target`: Target allocation amount (integer cents)

**Flags:** None

**Prerequisites:**
- Budget must be initialized
- Database must be unlocked
- Must have fewer than 8 buckets

**Creates:** New bucket in database

**Plain text output:**
```
Bucket created: Travel fund target=2000000 id=3
```

**JSON output:** None

**Examples:**

**Add bucket:**
```bash
balde bucket add "Travel fund" 2000000
# Target: $2,000.00
```

**Add with specific ID** (not possible, ID auto-generated):
```bash
balde bucket add "Car repairs" 1500000
# Auto-generated ID: 4
```

**Edge cases:**

| Error | Condition |
|-------|-----------|
| `maximum of 8 buckets exceeded` | Already have 8 buckets (free tier limit) |
| `invalid target: <input>` | Not an integer |

**Note:** Free version has 6 default buckets, allowing 2 custom buckets.

---

## balde transaction add

**Purpose:** Record a financial transaction

**Signature:**
```bash
balde transaction add <amount> <description> <account_id> <bucket_id>
```

**Arguments:**
1. `amount`: Transaction amount in cents (negative = expense, positive = income)
2. `description`: Transaction description (quoted if contains spaces)
3. `account_id`: Account ID (integer, string representation)
4. `bucket_id`: Bucket ID (integer, string representation)

**Flags:**
- `cmd.DisableFlagParsing = true` — all flags disabled, everything is positional

**Prerequisites:**
- Budget must be initialized
- Database must be unlocked
- Account ID must exist
- Bucket ID must exist

**Creates:** New transaction with `Date: time.Now()` and `Categorized: false`

**Plain text output:**
```
Transaction created: amount=-50000 desc="Coffee" id=1
```

**JSON output:** None

**Examples:**

**Record expense:**
```bash
balde transaction add -50000 "Coffee" 1 4
# $50.00 from account 1 to bucket 4
```

**Record income:**
```bash
balde transaction add 5000000 "Salary" 1 1
# $5,000.00 to account 1 and bucket 1
```

**Edge cases:**

| Error | Condition |
|-------|-----------|
| `invalid amount: <input>` | Not an integer |
| `account not found: <id>` | Account ID doesn't exist |
| `bucket not found: <id>` | Bucket ID doesn't exist |

**Important:** Date is always `time.Now()` — cannot set custom date via CLI.

---

## balde allocate

**Purpose:** Allocate money from rain to a specific bucket

**Signature:**
```bash
balde allocate <amount> <bucket_id>
```

**Arguments:**
1. `amount`: Amount to allocate (integer cents, typically positive)
2. `bucket_id`: Target bucket ID (integer, string representation)

**Flags:** None

**Prerequisites:**
- Budget must be initialized
- Database must be unlocked
- Bucket ID must exist

**Effect:** Increases bucket balance, does not check rain availability (can over-allocate)

**Plain text output:**
```
Allocated 500000 cents to bucket 3
```

**JSON output:** None

**Examples:**

**Allocate to bucket:**
```bash
balde allocate 500000 3
# Allocate $5,000.00 to bucket 3
```

**Withdraw from bucket** (negative amount):
```bash
balde allocate -500000 3
# Remove $5,000.00 from bucket 3 (returns to rain)
```

**Edge cases:**

| Error | Condition |
|-------|-----------|
| `invalid amount: <input>` | Not an integer |
| `bucket not found: <id>` | Bucket ID doesn't exist |

**Note:** No rain checking — you can allocate more than available, resulting in negative rain.

---

## balde rain

**Purpose:** Show unallocated money available

**Signature:**
```bash
balde rain
```

**Arguments:** None

**Flags:** None

**Prerequisites:**
- Budget must be initialized
- Database must be unlocked

**Effect:** Computes `rain = sum(account balances) - sum(bucket balances)`

**Plain text output:**
```
Rain (unallocated): 123456 cents
```

**JSON output:** None

**Examples:**

**Check rain:**
```bash
balde rain
# Output: Rain (unallocated): 123456 cents
# Meaning: $1,234.56 available to allocate
```

**Negative rain** (over-allocated):
```bash
balde rain
# Output: Rain (unallocated): -50000 cents
# Meaning: Buckets have $500.00 more than accounts
```

**Edge cases:** None (always succeeds if DB is unlocked)

---

## balde view buckets

**Purpose:** List all buckets with their allocation status

**Signature:**
```bash
balde view buckets [--json]
```

**Arguments:** None

**Flags:**
- `--json`: Output as JSON array (default: false)

**Prerequisites:**
- Budget must be initialized
- Database must be unlocked

**Effect:** Queries all buckets

**Plain text output** (tab-separated):
```
1	financial freedom	target=0	balance=0
2	fixed costs	target=2000000	balance=1500000
3	pleasures	target=500000	balance=0
```

**JSON output:**
```json
[
  {
    "ID": "1",
    "Name": "financial freedom",
    "Target": 0,
    "Balance": 0,
    "BudgetID": "default"
  },
  {
    "ID": "2",
    "Name": "fixed costs",
    "Target": 2000000,
    "Balance": 1500000,
    "BudgetID": "default"
  }
]
```

**Examples:**

**View buckets (plain):**
```bash
balde view buckets
# Tab-separated, one per line
```

**View buckets (JSON):**
```bash
balde view buckets --json
# Prettified JSON array
```

**Edge cases:** None (always succeeds if DB is unlocked)

---

## balde view transactions

**Purpose:** List all transactions

**Signature:**
```bash
balde view transactions [--json]
```

**Arguments:** None

**Flags:**
- `--json`: Output as JSON array (default: false)

**Prerequisites:**
- Budget must be initialized
- Database must be unlocked

**Effect:** Queries all transactions

**Plain text output** (tab-separated):
```
1	-50000	Coffee	2026-05-29T00:00:00Z
2	5000000	Salary	2026-05-28T00:00:00Z
```

**JSON output:**
```json
[
  {
    "ID": "1",
    "Amount": -50000,
    "Description": "Coffee",
    "Date": "2026-05-29T00:00:00Z",
    "AccountID": "1",
    "BucketID": "4",
    "Categorized": false
  },
  {
    "ID": "2",
    "Amount": 5000000,
    "Description": "Salary",
    "Date": "2026-05-28T00:00:00Z",
    "AccountID": "1",
    "BucketID": "1",
    "Categorized": false
  }
]
```

**Examples:**

**View transactions (plain):**
```bash
balde view transactions
# Tab-separated, one per line
```

**View transactions (JSON):**
```bash
balde view transactions --json
# Prettified JSON array
```

**Edge cases:** None (always succeeds if DB is unlocked)

---

## balde status

**Purpose:** Get complete budget snapshot

**Signature:**
```bash
balde status [--json]
```

**Arguments:** None

**Flags:**
- `--json`: Output as JSON object (default: false)

**Prerequisites:**
- Budget must be initialized
- Database must be unlocked

**Effect:** Returns accounts, buckets, transactions, and rain

**Plain text output:**
```
Accounts: 2
Buckets: 6
Transactions: 10
Rain: 500000 cents
```

**JSON output:**
```json
{
  "accounts": [
    {
      "ID": "1",
      "Name": "Checking",
      "Type": "checking",
      "Balance": 5000000
    },
    {
      "ID": "2",
      "Name": "Savings",
      "Type": "savings",
      "Balance": 10000000
    }
  ],
  "buckets": [
    {
      "ID": "1",
      "Name": "financial freedom",
      "Target": 0,
      "Balance": 0,
      "BudgetID": "default"
    }
  ],
  "transactions": [
    {
      "ID": "1",
      "Amount": -50000,
      "Description": "Coffee",
      "Date": "2026-05-29T00:00:00Z",
      "AccountID": "1",
      "BucketID": "4",
      "Categorized": false
    }
  ],
  "rain": 500000
}
```

**Examples:**

**Status overview:**
```bash
balde status
# Counts + rain in cents
```

**Full JSON export:**
```bash
balde status --json
# Complete snapshot with all objects
```

**Edge cases:** None (always succeeds if DB is unlocked)

---

## Not Yet Implemented (from spec)

These commands are specified in `AGENTS.md` but not yet implemented:

| Command | Spec |
|---------|------|
| `balde config set locale <code>` | Not implemented — no `config` command |
| `balde transaction import <file> --format csv` | Not implemented — `import/` directory empty |
| `balde transaction categorize` | Not implemented |
| `--quiet` flag | Not implemented anywhere |

When user asks for these, inform them they're not yet implemented and suggest alternative workflows.

---

## Config File Format (`balde.json`)

Located in the budget directory, contains currency and locale settings.

**Example:**
```json
{
  "locale": "en",
  "currency_symbol": "$",
  "decimal_separator": ".",
  "thousands_separator": ",",
  "frequency": "monthly",
  "encrypted": false
}
```

**Fields:**

| Field | Type | Default | Notes |
|-------|------|---------|-------|
| `locale` | string | `en` | Language code for i18n (MVP: en only) |
| `currency_symbol` | string | `$` | Currency symbol (`$`, `R$`, `€`, `£`, etc.) |
| `decimal_separator` | string | `.` | Decimal point separator (`.` or `,`) |
| `thousands_separator` | string | `,` | Thousands grouping separator (`,` or `.`) |
| `frequency` | string | `monthly` | Budget cycle: `weekly`, `fortnightly`, `monthly` |
| `encrypted` | boolean | `false` | Whether database is encrypted |

**Currency Formatting Examples:**

| Locale | Symbol | Decimal | Thousands | `500000` cents → |
|--------|--------|---------|-----------|-----------------|
| US (en) | `$` | `.` | `,` | `$5,000.00` |
| Brazil | `R$` | `,` | `.` | `R$ 5.000,00` |
| EU (de) | `€` | `,` | `.` | `5.000,00 €` |

---

## Frequency Conversion

When converting bucket targets between cycles:

| Source | To Weekly | To Fortnightly | To Monthly |
|--------|-----------|----------------|------------|
| Monthly | ÷ 4.33 | ÷ 2.17 | × 1 |
| Fortnightly | × 2 | × 1 | × 2.17 |
| Weekly | × 1 | ÷ 2 | × 4.33 |

**Note:** These conversions are not implemented in CLI yet — for skill knowledge only.

---

## Rain Calculation

**Formula:** `rain = sum(account balances) - sum(bucket balances)`

**Scenarios:**

| Scenario | Account Sum | Bucket Sum | Rain | Meaning |
|----------|-------------|------------|------|---------|
| Fresh budget | $5,000 | $0 | $5,000 | All money available to allocate |
| Partially allocated | $5,000 | $2,000 | $3,000 | $3,000 remaining |
| Fully allocated | $5,000 | $5,000 | $0 | All money in buckets |
| Over-allocated | $5,000 | $6,000 | -$1,000 | Buckets exceed accounts (warning) |

**When rain is negative:** The user has allocated more money to buckets than exists in accounts. This is a warning state — either a mistake or intentional over-budgeting.

---

## CLI Conventions

### Amount Input

- **Always integer cents:** `50000` = $500.00
- **Negative = expense:** `-50000` = -$50.00
- **Positive = income:** `50000` = +$50.00
- **No commas in input:** Use `50000`, not `50,000`

### Date Format

- **RFC3339 in JSON:** `2026-05-29T00:00:00Z`
- **CLI output:** ISO 8601 (`YYYY-MM-DD`)
- **Note:** `transaction add` always uses `time.Now()` — custom dates not supported

### ID Format

- **IDs are strings:** `"1"`, `"2"`, `"3"` (even though internally integers)
- **Auto-generated:** No manual ID assignment
- **Unique per type:** Account IDs, bucket IDs, and transaction IDs are separate sequences

---

## Testing Your Integration

To verify the skill works correctly:

1. **Initialize a test budget:**
   ```bash
   cd /tmp
   mkdir test-balde
   cd test-balde
   balde init
   ```

2. **Add test data:**
   ```bash
   balde account add "Test Checking" checking 10000000
   balde bucket add "Test Bucket" 5000000
   balde transaction add -500000 "Test Expense" 1 1
   ```

3. **Verify JSON outputs:**
   ```bash
   balde status --json | jq .
   balde view buckets --json | jq .
   balde view transactions --json | jq .
   ```

4. **Clean up:**
   ```bash
   cd /tmp
   rm -rf test-balde
   ```

This gives you a reproducible test environment for skill development.