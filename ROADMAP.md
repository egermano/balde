# Roadmap

Honest, short, prioritized. Each item states the problem it solves and what
"done" means — nothing here is a promise, but nothing here is vaporware
either. Items move to the [changelog](https://github.com/egermano/balde/releases)
when they ship.

## Now — the road to v0.1.0 stable

### 1. CSV import

- **Problem:** entering transactions by hand is the biggest daily friction; bank exports are where the data already lives.
- **Done when:** `balde transaction import <file> --format csv` creates uncategorized transactions with column mapping and duplicate detection (date + amount + description), with tests for empty rows, negative amounts and varied date formats. Planned pipeline is described in [`AGENTS.md`](AGENTS.md#csv-import-pipeline-planned-not-built).

### 2. Encrypted-migration integrity

- **Problem:** `balde encrypt` re-creates accounts and buckets with new auto-increment IDs, so transactions that referenced archived buckets can end up pointing at nothing.
- **Done when:** migration preserves entity IDs (or remaps every reference), with a regression test that encrypts a budget containing an archived bucket and verifies all references resolve.

### 3. Backup & restore

- **Problem:** backing up means hand-copying `balde.db` and `balde.json` and hoping you got both; users discover this the hard way.
- **Done when:** `balde backup` writes a timestamped snapshot of both files and `balde restore <snapshot>` brings a budget back, documented in the README. Interim guidance: [docs/backup.md](docs/backup.md).

## Next

### 4. pt-BR locale (i18n infrastructure)

- **Problem:** the budgeting method is popular in Brazil, but output formatting is US-only.
- **Done when:** locale dictionaries load from JSON, `balde config set locale pt-BR` changes user-facing strings, and adding a new language requires zero code changes.

### 5. v0.1.0 stable

- **Problem:** the alpha label is honest but limits everyday adoption.
- **Done when:** items 1–3 ship, the CLI surface is documented and frozen for the 0.x series, and installers are verified on Linux, macOS and Windows from an actual release.

## Explicitly not on the roadmap

Cloud sync, multi-user, web UI, charting dashboards, floating-point money.
These are design boundaries, not missing features — see the constraints
section in [`AGENTS.md`](AGENTS.md#constraints).
