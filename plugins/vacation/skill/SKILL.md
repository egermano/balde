---
name: balde-vacation
description: Plan or track a vacation, trip, holiday, travel savings goal, or travel budget in Balde. Use this skill whenever the user mentions saving for a trip, wants to budget a vacation, asks how much to set aside per month for travel, or asks for the progress of a vacation fund.
---

# Balde Vacation Planner

Use the installed `vacation` plugin to create a dedicated vacation bucket and
turn a travel budget into affordable monthly allocations. Balde is
agent-operated, so favor structured CLI commands and clearly report what will
happen before spending or allocating money.

## Before planning

1. Work in the user's budget directory (where `balde.db` and `balde.json`
   live).
2. Plugin support is experimental. Set this for every command session:

   ```sh
   export BALDE_EXPERIMENTAL=1
   ```

3. Check current buckets and rain before making a plan:

   ```sh
   balde view buckets --json
   balde rain
   ```

The free Balde budget has six default buckets and supports at most eight.
Creating a vacation bucket requires a free bucket slot. If all slots are used,
explain the limit and ask the user whether to archive an unused bucket rather
than deleting or changing one automatically.

## Create a plan

Collect three inputs if they were not provided:

- destination or vacation name
- vacation month as `YYYY-MM` (must be a future month)
- total budget in **integer cents**

Run:

```sh
balde vacation plan "Japan" --date 2027-03 --budget 1200000
```

This creates a bucket named `vacation-japan-2027-03`, sets its target to the
full budget, and allocates the first installment from current rain. It does
not pre-allocate money that has not arrived: follow the printed schedule for
later months. The final installment can be smaller because cents are rounded
up for earlier months.

Always report the created bucket, amount allocated now, regular monthly
amount, final amount when different, and how many allocations remain.

## Check a plan

```sh
balde vacation schedule "Japan" --date 2027-03
```

This reports the remaining balance and the next allocation command. Use it
before each monthly allocation or whenever the user asks how their vacation
fund is progressing.

## Safety and error handling

- Never invent a date or budget. Ask a concise question when one is missing.
- Treat money values as cents. For example, R$12,000.00 is `1200000` cents.
- If the plugin reports insufficient rain or the 8-bucket limit, explain the
  structured error and ask for a decision; do not force an allocation.
- A bucket name embeds destination and date. If a duplicate plan exists, use
  `vacation schedule` rather than creating another one.

Read [the command reference](references/command-reference.md) when you need
exact input/output details or a response template.
