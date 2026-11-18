# Agent integration demo

This documents a **real session**: an AI agent (OpenCode, GLM model) with the
balde skill installed, operating a budget end to end. No compatibility is
claimed that was not executed. All data is synthetic.

## Setup

The agent environment had the balde skill installed. To reproduce in your own
agent (OpenCode, Claude Code, Codex CLI, or any tool that supports the skills
format):

```sh
npx skills add egermano/balde --skill balde
```

The skill (`.agents/skills/balde/SKILL.md` in this repo) teaches the agent:

- how to run balde and where budgets live (`balde.json` + `balde.db`)
- to prefer `--json` output and handle `null` arrays
- to format integer cents into the budget's configured currency using integer
  math only
- to check rain before allocating, and when to ask the user for confirmation

## The task

> "Check how much rain I have, allocate $400.00 to my goals bucket, then show
> me the state of the budget."

## What the agent did (actual commands and output)

The agent followed the skill workflow: work from the budget directory, read
the currency config, check rain before any allocation, use `--json` for
state.

```console
$ cat balde.json
{
  "currency_symbol": "$",
  "decimal_separator": ".",
  "thousands_separator": ",",
  "frequency": "monthly"
}

$ balde rain
Rain (unallocated): 300000 cents

# skill rule: verify rain covers the allocation before running it
# 300000 cents ($3,000.00) >= 40000 cents ($400.00) -> safe to proceed

$ balde allocate 40000 6 --force
Allocated 40000 cents to bucket 6

$ balde status --json
{ "accounts": [ { "id": "1", "name": "checking", "type": "checking", "balance": 500000 } ],
  ... "rain": 260000 }
```

## What the agent showed the user

Per the skill's communication contract, the agent formats cents into the
configured currency (`$`, `1,234.56`) and presents a dashboard:

✓ Allocated **$400.00** to bucket **goals** (ID: 6)

**New rain balance: $2,600.00**

## Why this matters

- The CLI works with zero AI involvement — every command above is a plain
  shell command
- The agent's value is orchestration: resolving "goals" to bucket ID 6,
  checking rain first, converting cents to currency, presenting a readable
  summary, and refusing unsafe operations (over-allocation) unless confirmed
- Commands and JSON output are stable and deterministic, which is what makes
  this automation reliable

## Privacy

This session used synthetic data. With a real budget, the agent's tools and
provider configuration determine what data is read and sent — see the
[data privacy note](../README.md#data-privacy-local-first--llms) in the README.
