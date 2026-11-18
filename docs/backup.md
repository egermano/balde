# Backup, restore & destructive operations

Balde has no built-in backup command yet (it is on the
[roadmap](../ROADMAP.md)); this document explains what you can do today,
safely, and which operations destroy data.

## What a budget is made of

A budget directory contains exactly two files that matter:

| File | Contents | If lost |
| --- | --- | --- |
| `balde.db` | All accounts, buckets, transactions (SQLite; encrypted at rest if you ran `balde encrypt`) | Everything is gone |
| `balde.json` | Configuration: currency symbol, separators, frequency, `encrypted` flag | CLI falls back to defaults; encrypted budgets won't open without it |

Back up **both** files. A database without its config is a problem; a config
without its database is just settings.

## Manual backup

```sh
# from the budget directory, while no other balde command is running
tar czf ~/balde-backup-$(date +%Y%m%d).tar.gz balde.db balde.json
```

Run backups when no command is writing (a copy taken mid-write can capture a
torn SQLite state). Verify a backup by restoring it into a fresh directory
and running `balde status`.

To restore: copy both files into a (new) directory and run `balde status`
there. For encrypted budgets the restored `balde.db` stays encrypted — you
will need the password and `balde unlock`.

Note: `balde encrypt` automatically leaves the pre-encryption plain database
at `balde.db.backup`. That file is **not** encrypted — delete it once you
have confirmed the encrypted budget works and you have a real backup.

## Destructive operations

| Command | What it actually does | Reversible? |
| --- | --- | --- |
| `balde transaction delete <id>` | Hard-deletes the transaction **and rolls back its effect** on account/bucket balances | No |
| `balde bucket delete <id>` | **Archives** the bucket (asks for confirmation if it holds money or has transactions) | Yes in data terms — archived rows remain in the database, but there is no unarchive command yet |
| `balde encrypt` | Re-encrypts the database in place (plain backup left behind) | Only via the backup file |
| `balde init` in an existing budget directory | Refuses (`budget already initialized`) — safe | n/a |

Spending from a bucket (`transaction add` with a negative amount) never
destroys history — it is a normal, reversible-by-record transaction.

## What cannot be recovered

- **A forgotten encryption password.** Encrypted budgets are AES-256-GCM;
  without the password the data is intentionally unrecoverable. There is no
  backdoor.
- **A deleted transaction** (after the command succeeds).
- **A budget file that was never backed up** and is corrupted or deleted.

If you rely on Balde daily, put the two files under whatever backup rotation
you already use for important documents — that is the whole story until
`balde backup` ships.
