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

All commands must run from a directory containing `balde.json` and `balde.db`. If these files are missing in the current working directory, tell the user:

```
No budget found in this directory. Navigate to a directory with balde.json and balde.db,
or run 'balde init [--dir <path>]' to create a new budget.
```

## Currency Formatting

The CLI stores all amounts as **integer cents** (e.g., `50000` = $500.00). To format properly:

1. Read `balde.json` to get currency settings:
   - `currency_symbol` (e.g., `$`, `R$`, `€`)
   - `decimal_separator` (e.g., `.` or `,`)
   - `thousands_separator` (e.g., `,` or `.`)

2. Convert cents to display value: `value = cents / 100.0`

3. Apply separators:
   - Add thousands separator every 3 digits
   - Insert decimal separator for the 2 decimal places

4. Prepend currency symbol

Example: `cents=123456` → `$1,234.56` (US) or `R$ 1.234,56` (Brazil)

**Why format this way:** The CLI outputs raw integers which are hard for users to parse. By reading the config and formatting correctly, we present data in a way that matches the user's locale and expectations.

## Command Workflows

### balde status
Get a complete snapshot of the budget state.

```bash
balde status --json
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
| ID | Name             | Target    | Balance   | Fill % |
|----|------------------|-----------|-----------|--------|
| 1  | financial freedom| $1,000.00 | $500.00   | 50%    |

### Recent Transactions (last 10)
| ID | Date       | Description  | Amount    |
|----|------------|--------------|-----------|
| 1  | 2026-05-29 | Rent         | -$1,500.00 |
```

**Why show all data:** `status` is the "dashboard" command — users want the full picture in one glance.

### balde view buckets
List all buckets with their allocation status.

```bash
balde view buckets --json
```

Present as markdown table with fill percentage:

```markdown
## Buckets Overview

| ID | Name             | Target    | Balance   | Fill % | Status      |
|----|------------------|-----------|-----------|--------|-------------|
| 1  | financial freedom| $1,000.00 | $1,000.00 | 100%   | Full ✓      |
| 2  | fixed costs      | $2,000.00 | $1,500.00 | 75%    | On track    |
| 3  | pleasures        | $500.00   | $0.00     | 0%     | Empty       |
```

Status logic:
- 100% = "Full ✓"
- 80-99% = "Near full"
- 20-79% = "On track"
- 0-19% = "Low"
- 0% = "Empty"

### balde view transactions
List all transactions chronologically.

```bash
balde view transactions --json
```

Present with formatted amounts (red for expenses, green for income):

```markdown
## All Transactions

| ID | Date       | Description    | Amount    | Account | Bucket     |
|----|------------|----------------|-----------|---------|------------|
| 1  | 2026-05-29 | Monthly salary | +$5,000.00| checking| (none)     |
| 2  | 2026-05-28 | Rent           | -$1,500.00| checking| fixed costs|
```

Note: Expenses are negative — format with a minus sign and red color indicator if markdown supports it.

### balde rain
Show unallocated money available for distribution.

```bash
balde rain
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

**Clarification before running:**
- Parse the amount — if ambiguous whether income/expense, ask: "Is this an allocation (positive) or withdrawal (negative)?"
- Check rain first — if rain < amount, warn: "Rain is only $X, you're trying to allocate $Y. Confirm overspend?"
- Validate bucket_id — if invalid or ambiguous, ask user to specify

```bash
balde allocate <cents> <bucket_id>
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
balde account add "<name>" "<type>" <cents>
```

Confirmation:

```markdown
✓ Created account: "My Checking" (checking) with balance $5,000.00
Account ID: 1
```

### balde bucket add <name> <target>
Add a new bucket (envelope) for budgeting.

**Clarification needed:**
- Check bucket count — if already 8, warn: "Maximum 8 buckets reached. Free version supports 8; delete an existing bucket first."
- If target amount ambiguous, ask user to confirm

```bash
balde bucket add "<name>" <cents>
```

Confirmation:

```markdown
✓ Created bucket: "Emergency Fund" with target $5,000.00
Bucket ID: 3
```

**Why warn about 8 buckets:** This is a hard constraint in the free tier. Users need to know before they try to add a 9th bucket.

### balde transaction add <amount> <desc> <account_id> <bucket_id>
Record a financial transaction.

**Clarification needed:**
- Amount: Is this income (positive) or expense (negative)? If user provides "$50" with no sign, ask.
- Account/Bucket: If IDs are ambiguous or names are given, resolve with user. Try to match by name if possible.
- Description: If missing or too vague, ask for a meaningful description

```bash
balde transaction add <cents> "<description>" <account_id> <bucket_id>
```

Confirmation:

```markdown
✓ Recorded transaction: "-$50.00" (Coffee) from account 1 to bucket 4

Transaction ID: 15
```

### balde init [--password] [--dir]
Initialize a new budget in a directory.

**Clarification needed:**
- If `--dir` not provided, confirm: "Initialize in current directory?"
- Encryption: If no password provided, ask: "Do you want to enable encryption? (requires a password)"
- If yes to encryption, get password securely (use `--password` flag or prompt)

```bash
balde init [--password <pass>] [--dir <path>]
```

Output:

```markdown
✓ Budget initialized successfully

**Location:** /path/to/directory
**Encryption:** Enabled (password-protected)
**Default buckets:** 6 created (financial freedom, fixed costs, pleasures, comfort, knowledge, goals)
```

**Why ask about encryption:** Encryption is optional but has UX implications (unlock required each session). Users should opt-in consciously.

### balde unlock [--password] [--db]
Unlock an encrypted budget database.

**Prerequisite:** Database must be encrypted (`encrypted: true` in `balde.json`)

```bash
balde unlock [--password <pass>]
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
balde lock
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
balde encrypt [--password <pass>]
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
3. **Constraint violation:** User tries to add 9th bucket — warn about max 8 limit
4. **Missing prerequisites:** No budget in directory, DB locked, or balde not installed
5. **Unclear intent:** User says "balance" — clarify if they want `status`, `rain`, or account/bucket balances
6. **Over-allocation:** Rain is $500, user wants to allocate $600 — confirm intent to overspend

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
| `maximum of 8 buckets exceeded` | Free tier limit reached | Delete a bucket first or upgrade |
| `passwords do not match` | Password confirmation failed | Re-enter password carefully |
| `no active session` | Database is locked | Run `balde unlock` first |
| `invalid amount: <input>` | Amount parse failed | Use format: `1234` (cents) or `1234.56` with configured separators |

Always show:
1. What went wrong (plain English)
2. Why it failed (briefly)
3. How to fix (concrete next step)

## Communication Contract

When the user asks a budget-related question:

1. **Understand intent:** Parse what they want (view, modify, analyze)
2. **Check prerequisites:** Verify balde is installed, budget exists, DB is unlocked
3. **Run command:** Execute appropriate `balde` command
4. **Parse output:** Prefer `--json` for structured data, fallback to plain text
5. **Format nicely:** Convert cents → currency, create markdown tables
6. **Present clearly:** Show results first, then context/stats

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

## References

For detailed command signatures, argument types, and edge cases, see:
- `references/command-reference.md` — Complete command documentation with I/O examples