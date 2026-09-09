---
name: balde
description: Use this whenever the user wants to manage their budget, check finances, allocate money between buckets, view transactions or accounts, run "make it rain" allocations, or interacts with the Balde bucket-budgeting CLI. Also triggers for any financial management task involving the Balde tool.
---

# Balde Budget Manager Skill

A skill for managing a Balde budget through the CLI. This skill runs `balde` commands, parses outputs (preferring JSON where available), converts raw cents to human-readable currency, and presents results as clean markdown tables.

## Prerequisites

First, verify `balde` is installed and accessible:

```bash
if ! command -v balde &> /dev/null; then
    echo "Balde CLI not found. Install it from: https://github.com/egermano/balde"
    exit 1
fi
```

All commands must run from a directory containing `balde.json` and `balde.db`. Use one of these patterns:

```bash
# Pattern 1: cd + command (recommended for bash tool)
cd /path/to/budget && balde <command>

# Pattern 2: Use workdir parameter (if available)
balde <command>  # with workdir=/path/to/budget
```

If budget files are missing in the current working directory, tell the user:

```
No budget found in this directory. Navigate to a directory with balde.json and balde.db,
or run 'balde init [--dir <path>]' to create a new budget.
```

**Important:** Only `balde init` and `balde unlock` support a `--dir` flag. All other commands must be run from within the budget directory.

## Currency Formatting

The CLI stores all amounts as **integer cents** (e.g., `50000` = $500.00). To format properly:

1. Read `balde.json` to get currency settings:
   - `currency_symbol` (e.g., `$`, `R$`, `€`)
   - `decimal_separator` (e.g., `.` or `,`)
   - `thousands_separator` (e.g., `,` or `.`)

2. Convert cents to display value: `value = cents / 100.0`

3. Apply separators:
   - Add thousands separator every 3 digits from right
   - Insert decimal separator for the 2 decimal places
   - Handle zero: show `0.00`
   - Handle negative: prepend `-` before symbol or after (locale-dependent)

4. Prepend currency symbol

**Examples:**
- `cents=123456` → `$1,234.56` (US)
- `cents=123456` → `R$ 1.234,56` (Brazil)
- `cents=0` → `$0.00`
- `cents=-50000` → `-$50.00`

**Why format this way:** The CLI outputs raw integers which are hard for users to parse. By reading the config and formatting correctly, we present data in a way that matches the user's locale and expectations.

## JSON Parsing Guidelines

When parsing JSON from `balde status --json`, `balde view buckets --json`, or `balde view transactions --json`:

**Handle null fields:**
```bash
# Use jq to convert null to empty array
accounts=$(echo "$json" | jq '.accounts // []')
transactions=$(echo "$json" | jq '.transactions // []')
```

**Why:** When no accounts or transactions exist, the CLI returns `null` instead of `[]`, which would crash code expecting an array.

## Command Workflows

### balde status
Get a complete snapshot of the budget state.

```bash
cd /path/to/budget && balde status --json
```

Parse the JSON and present as:

```markdown
## Budget Status

- **Accounts:** N
- **Buckets:** N (max 8)
- **Transactions:** N
- **Rain (unallocated):** $X,XXX.XX

### Accounts
| ID | Name     | Type      | Balance   |
|----|----------|-----------|-----------|
| 1  | Checking | checking  | $5,000.00 |

### Buckets
| ID | Name             | Target    | Balance   | Fill % | Status      |
|----|------------------|-----------|-----------|--------|-------------|
| 1  | financial freedom| $0.00     | $500.00   | -      | No target   |
| 2  | fixed costs      | $2,000.00 | $1,500.00 | 75%    | On track    |

### Recent Transactions (last 10)
| ID | Date                 | Description  | Amount    | Account | Bucket     |
|----|----------------------|--------------|-----------|---------|------------|
| 1  | 2026-05-29 17:17:14  | Rent         | -$1,500.00| 1       | 2          |
```

**Fill % calculation:**
- If target == 0: show "-" (no target set)
- If target > 0: `fill % = (balance / target) * 100`

**Date format:** Show full RFC3339 timestamp (`YYYY-MM-DD HH:MM:SS`) from JSON.

**Why show all data:** `status` is the "dashboard" command — users want the full picture in one glance.

### balde view buckets
List all buckets with their allocation status.

```bash
cd /path/to/budget && balde view buckets --json
```

Present as markdown table with fill percentage:

```markdown
## Buckets Overview

| ID | Name             | Target    | Balance   | Fill % | Status      |
|----|------------------|-----------|-----------|--------|-------------|
| 1  | financial freedom| $0.00     | $0.00     | -      | No target   |
| 2  | fixed costs      | $2,000.00 | $1,500.00 | 75%    | On track    |
| 3  | pleasures        | $500.00   | $0.00     | 0%     | Empty       |
```

Status logic:
- target == 0, balance == 0: "No target"
- target > 0, balance == 0: "Empty"
- target > 0, balance > 0, fill < 20%: "Low"
- target > 0, 20% ≤ fill < 80%: "On track"
- target > 0, 80% ≤ fill < 100%: "Near full"
- target > 0, fill == 100%: "Full ✓"

### balde view transactions
List all transactions chronologically.

```bash
cd /path/to/budget && balde view transactions --json
```

Present with formatted amounts (use minus sign for expenses, no color if markdown doesn't support):

```markdown
## All Transactions

| ID | Date                 | Description    | Amount    | Account | Bucket     |
|----|----------------------|----------------|-----------|---------|------------|
| 1  | 2026-05-29 17:17:14  | Monthly salary | +$5,000.00| 1       | (none)     |
| 2  | 2026-05-28 10:30:00  | Rent           | -$1,500.00| 1       | 2          |
```

Note: Expenses are negative — format with a minus sign. If BucketID is empty string, show "(none)".

**Important:** Transaction IDs in the CLI response may be incomplete (`id=` without value). Don't rely on the CLI output for the ID; get it from the JSON response.

### balde rain
Show unallocated money available for distribution.

```bash
cd /path/to/budget && balde rain
```

Output format:

```markdown
## Rain (Unallocated Money)

**Available:** $1,234.56

This amount is the difference between your account balances and bucket allocations.
Use 'balde allocate <amount> <bucket_id>' to distribute to buckets.
```

**Why separate command:** Users often check rain before deciding allocations — this gives them a quick number without the full status overhead.

### balde allocate <amount> <bucket_id>
Move money from rain into a specific bucket.

**CRITICAL:** The CLI does NOT prevent over-allocation. You MUST check rain first.

**Workflow:**
1. Parse the amount — if ambiguous whether income/expense, ask: "Is this an allocation (positive) or withdrawal (negative)?"
2. **Check rain first:**
   ```bash
   rain=$(cd /path/to/budget && balde rain | grep -oP '\d+(?= cents)')
   ```
   If rain < amount, warn: "Rain is only $X, you're trying to allocate $Y. This will result in negative rain. Continue? (y/N)"
3. Validate bucket_id — if invalid or ambiguous, ask user to specify
4. Resolve duplicate bucket names if user provides name instead of ID (see Clarification Rules)

```bash
cd /path/to/budget && balde allocate <cents> <bucket_id>
```

After running, show confirmation:

```markdown
✓ Allocated $500.00 to bucket "fixed costs" (ID: 2)

**New rain balance:** $1,234.56
```

### balde account add <name> <type> <balance>
Add a new financial account.

**Clarification needed:**
- If balance is ambiguous (e.g., "1000" vs "1000.00"), ask for cents or decimal
- If type is not one of `checking`, `savings`, `credit`, ask user to pick

```bash
cd /path/to/budget && balde account add "<name>" "<type>" <cents>
```

Confirmation (parse CLI output for ID):
```markdown
✓ Created account: "My Checking" (checking) with balance $10,000.00
Account ID: 1
```

**Note:** Account IDs start at 1, not 0.

### balde bucket add <name> <target>
Add a new bucket (envelope) for budgeting.

**Clarification needed:**
- Check bucket count first by running `balde status --json` — if already 8, warn: "Maximum 8 buckets reached. Free version supports 8; cannot delete buckets in current CLI."
- If target amount ambiguous, ask user to confirm

```bash
cd /path/to/budget && balde bucket add "<name>" <cents>
```

Confirmation:
```markdown
✓ Created bucket: "Emergency Fund" with target $5,000.00
Bucket ID: 3
```

**Why warn about 8 buckets:** This is a hard constraint in the free tier. Users need to know before they try to add a 9th bucket. **Important:** There is no CLI command to delete buckets.

### balde transaction add <amount> <desc> <account_id> <bucket_id>
Record a financial transaction.

**Clarification needed:**
- Amount: Is this income (positive) or expense (negative)? If user provides "$50" with no sign, ask.
- Account/Bucket: If IDs are ambiguous or names are given, resolve with user. Try to match by name if possible.
- Description: If missing or too vague, ask for a meaningful description

```bash
cd /path/to/budget && balde transaction add <cents> "<description>" <account_id> <bucket_id>
```

Confirmation (CLI output may not include ID):
```markdown
✓ Recorded transaction: "-$35.00" (Coffee) from account 1 to bucket 4

Transaction ID: 1  (Note: get from 'balde view transactions --json' if CLI output is incomplete)
```

**Important:** The CLI transaction add response may show `id=` without a value. Use `balde view transactions --json` to get the actual ID.

### balde init [--password] [--dir]
Initialize a new budget in a directory.

**Clarification needed:**
- If `--dir` not provided, confirm: "Initialize in current directory?"
- **Non-interactive mode:** Use `echo "N" | balde init` to skip encryption prompt
- Encryption: If user wants encryption, get password securely (use `--password` flag)

```bash
# Non-interactive (no encryption)
cd /path/to/budget && echo "N" | balde init

# Interactive (will prompt)
cd /path/to/budget && balde init

# With password
cd /path/to/budget && balde init --password <pass>
```

Output:
```markdown
✓ Budget initialized successfully

**Location:** /path/to/directory
**Encryption:** Disabled (or password-protected if specified)
**Default buckets:** 6 created (financial freedom, fixed costs, pleasures, comfort, knowledge, goals)
```

**Why use non-interactive mode:** In automated contexts, the encryption prompt may hang. Use `echo "N" | balde init` to skip it.

### balde unlock [--password] [--db]
Unlock an encrypted budget database.

**Prerequisite:** Database must be encrypted (`encrypted: true` in `balde.json`)

```bash
cd /path/to/budget && balde unlock [--password <pass>]
```

Or with `--dir` flag (only command that supports it):
```bash
balde unlock --dir /path/to/budget [--password <pass>]
```

Output:
```markdown
✓ Database unlocked successfully

**Session valid:** 30 minutes
**Session file:** /tmp/balde-session-<hash>
```

### balde lock
Lock an encrypted database by invalidating the session.

```bash
cd /path/to/budget && balde lock
```

Output:
```markdown
✓ Database locked
Session invalidated
```

### balde encrypt [--password] [--db]
Convert a plain database to encrypted.

**Clarification needed:**
- Confirm this action: "This will encrypt your database. Backup created automatically. Continue?"
- Get password if not provided

```bash
cd /path/to/budget && balde encrypt [--password <pass>]
```

Output:
```markdown
✓ Database encrypted successfully

**Backup:** balde.db.backup
**Session created:** Valid for 30 minutes
```

**Why confirm:** Encryption is irreversible (without password). Users should be certain.

## Clarification Rules

Ask the user when:

1. **Ambiguous amount:** User says "50" — clarify income vs expense
2. **Multiple matches:** User says "allocate to 'rent'" but 2 buckets match substring — list matches and ask which
3. **Duplicate bucket names:** User provides bucket name that matches multiple buckets:
   ```markdown
   Multiple buckets match "fixed costs":
     ID 2: fixed costs (target: $0, balance: $500.00)
     ID 8: fixed costs (target: $2,000, balance: $0.00)

   Which bucket do you want to allocate to?
   ```
4. **Constraint violation:** User tries to add 9th bucket — warn about max 8 limit and note that buckets cannot be deleted
5. **Missing prerequisites:** No budget in directory, DB locked, or balde not installed
6. **Unclear intent:** User says "balance" — clarify if they want `status`, `rain`, or account/bucket balances
7. **Over-allocation:** Rain is $500, user wants to allocate $600 — confirm intent to overspend with explicit warning

Proceed without asking when:

1. User provides explicit IDs (e.g., `bucket_id: 3`)
2. Amount has explicit sign (e.g., `-500` = expense, `+500` = income)
3. User confirms (e.g., "yes" to a warning)
4. Command is read-only (view, status, rain) — always safe

## Error Handling

Common CLI errors and how to present them:

| Error Message | User-Facing Explanation | Suggested Fix |
|---------------|-------------------------|---------------|
| `budget already initialized` | A budget already exists in this directory | Use existing budget or init in different directory |
| `database is not encrypted` | Trying to unlock/encrypt a plain database | Use `balde encrypt` to add encryption |
| `maximum of 8 buckets exceeded` | Free tier limit reached | Cannot delete buckets in current CLI; upgrade or reinitialize budget |
| `passwords do not match` | Password confirmation failed | Re-enter password carefully |
| `no active session` | Database is locked | Run `balde unlock` first |
| `invalid amount: <input>` | Amount parse failed | Use format: `1234` (cents) or `1234.56` with configured separators |
| `bucket not found: <id>` | Bucket ID doesn't exist | Run `balde view buckets --json` to see valid IDs |

Always show:
1. What went wrong (plain English)
2. Why it failed (briefly)
3. How to fix (concrete next step)

## Communication Contract

When the user asks a budget-related question:

1. **Understand intent:** Parse what they want (view, modify, analyze)
2. **Check prerequisites:** Verify balde is installed, budget exists, DB is unlocked
3. **Navigate to budget directory:** Use `cd /path && balde command` pattern
4. **Run command:** Execute appropriate `balde` command
5. **Parse output:** Prefer `--json` for structured data, handle null fields, fallback to plain text
6. **Format nicely:** Convert cents → currency, create markdown tables, handle edge cases
7. **Present clearly:** Show results first, then context/stats

Example response structure:
```markdown
<Optional: What I'm doing>
I'll check your current bucket allocation status.

<Command output formatted nicely>
## Buckets Overview
[... table ...]

<Optional: Insights or next steps>
You have $500 of rain available. Would you like to allocate to the "fixed costs" bucket?
```

## Known CLI Limitations

The skill works around these current CLI limitations:

- **No bucket deletion:** Once 8 buckets are created, you cannot add more. There is no `balde bucket delete` command.
- **No separate account view:** To see accounts, use `balde status --json` (no `balde view accounts` command).
- **No transaction update/delete:** Transactions are immutable once created.
- **Interactive init prompt:** `balde init` prompts for encryption. Use `echo "N" | balde init` for non-interactive mode.
- **No --dir flag for most commands:** Only `init` and `unlock` support `--dir`. Run other commands from budget directory.

## References

For detailed command signatures, argument types, and edge cases, see:
- `references/command-reference.md` — Complete command documentation with I/O examples