# Balde Plugin System — Plan

> Status: **planned** — implementation starting on `feature/plugin-system`.
> Feature is **experimental** from day one (see Experimental gating).

Plugins let users create their own features (importers, reports, commands,
connectors) without waiting on the core maintainer. Plugins are installed
from GitHub repo references and run as external processes.

## Locked decisions

| Decision | Choice |
|---|---|
| Runtime | Subprocess + versioned JSON-RPC 2.0 over stdio (exec) |
| First capability | Custom commands (`balde <plugin-cmd>`) |
| Install scope | Project-local (`.balde/plugins/` next to `balde.db`) |
| Plugin language | Any executable; Go SDK later as convenience |
| Experimental gate | Opt-in: `--experimental` flag or `BALDE_EXPERIMENTAL=1` |
| Dev tooling | `plugin dev` + `plugin check` + example plugin in first increment |
| Docs | `docs/plugins/` markdown set, versioned with code |

## Phase A — Foundation: `app` service + stable wire schema

Problem: `core` structs have no JSON tags (wire shape is accidental Go field
names) and every CLI command copy-pastes the `--json` encoder. The plugin
protocol needs a stable, explicitly-tagged schema and one validation path.

- New `app/` package: `App` wrapping `core.Budget` + `core.Store` with typed
  operations mirroring the CLI surface (`Status`, `ListAccounts`,
  `AddTransaction`, `Allocate`, `Rain`, …).
- DTOs with explicit snake_case JSON tags + structured
  `ErrorResponse{code, message}`. Core types never cross the wire.
- Refactor existing CLI commands onto `app`; one shared JSON emitter.
- TDD: golden tests pin the new DTO schema; behavior parity for every
  refactored command.

## Phase B — Plugin skeleton: manifest, install, lockfile, gate

- `balde-plugin.json` manifest:

```jsonc
{
  "name": "vacation",
  "version": "0.1.0",
  "protocol": 1,
  "capabilities": [
    { "type": "command", "name": "vacation", "description": "Plan vacation savings" }
  ],
  "permissions": {
    "read":  ["accounts", "buckets", "transactions"],
    "write": ["transactions", "allocate"]
  },
  "entrypoint": { "build": "go build -o bin/vacation .", "run": "bin/vacation" }
}
```

- `balde plugin install github.com/user/repo[@ref]`:
  - `git clone` (git is a documented prerequisite), pinned to resolved commit
    SHA into `.balde/plugins/src/<name>/`
  - run `entrypoint.build` if present → artifact in `.balde/plugins/bin/`
  - lockfile `.balde/plugins/lock.json` records:
    `{name, source, sha, version, artifactHash, capabilities, permissions}`
- `balde plugin list` (audit view incl. permissions), `plugin remove`,
  `plugin update`.
- **Experimental gate**: all `plugin` commands and plugin-registered commands
  refuse to run without `--experimental` or `BALDE_EXPERIMENTAL=1`. Notice on
  every run (stderr, `--quiet`-aware). While experimental, `protocol: 1` may
  have breaking changes; semver freeze starts when the feature goes stable.

## Phase C — Runtime: protocol + custom commands + dev tooling

Protocol (newline-delimited JSON-RPC 2.0, both directions, id-matched):

1. Host spawns `<entrypoint> --balde-protocol 1` with a **scrubbed env**
   (no DB path, no `BALDE_PASSWORD`).
2. Host → plugin: `initialize {protocol, budgetMeta}` (currency symbols,
   frequency — nothing sensitive).
3. Plugin → host: manifest echo + supported methods.
4. Host → plugin: `command/execute {args, flags}`.
5. During execution plugin → host calls (permission-checked):
   - read: `host/listAccounts`, `host/listBuckets`, `host/listTransactions`,
     `host/status`
   - write: `host/addTransaction`, `host/allocate`, `host/addBucket` —
     identical validation to the CLI (8-bucket limit, archived buckets,
     negative-rain rules); undeclared permission → `E_PERM`
6. Plugin returns `{text, json}` → host renders (respects `--json`,
   `--quiet`).

Command registration: `NewRootCmd()` scans the project lockfile and registers
each `command` capability as a cobra subcommand from manifest metadata.
Plugin process spawns lazily on invocation only; artifact hash verified
against lockfile before each spawn.

Dev tooling (first increment):

- `balde plugin dev --dir ./my-plugin` — run from source, no install.
- `balde plugin check [--dir]` — conformance validator: manifest schema,
  entrypoint runs, handshake roundtrip, known capabilities only. For authors'
  local + CI use.
- `examples/plugins/hello/` — canonical example; doc walkthrough target +
  conformance fixture for our own tests + user template.

## Documentation — `docs/plugins/`

| File | Contents |
|---|---|
| `README.md` | Index, experimental warning, quickstart (use / create / test) |
| `using-plugins.md` | Install/list/remove/update, lockfile, permissions audit, troubleshooting |
| `creating-plugins.md` | Manifest reference, capabilities, permissions, hello-plugin walkthrough |
| `protocol.md` | Canonical wire contract: handshake, methods, host functions, error codes |
| `testing-plugins.md` | Dev mode, conformance check, fixtures, CI patterns |
| `recommended.md` | Curated list + inclusion criteria (GitHub-installable, valid manifest, declared permissions, tested, maintained) |

## Later phases

- **D** — importer capability, built with the `import/` package
  (`balde transaction import --plugin x`, dedup pipeline).
- **E** — reporter/exporter capability with `report/`.
- **F** — connectors (openbanking) declared as `connector` capability but run
  agent-side as MCP servers (harness holds credentials, never balde);
  `balde mcp` server mode; Go SDK (`plugin/sdk/`); plugin registry; signed
  plugins; Wasm swap-in (manifest/protocol is runtime-agnostic).

## Architecture rule

```
plugin/ ──→ app ──→ core ←── store
```

`core` stays pure; `plugin` never imports `cli`; `store` injected into `app`.

## Security model

- Explicit `plugin install` = the trust grant; everything after is enforcement.
- SHA-pinned lockfile + artifact-hash verification on every spawn.
- Clean env per spawn; no DB path, no password, no session leakage. Encrypted
  DBs work unchanged (host holds the session).
- All writes mediated through host functions — single validation path for
  humans, agents, and plugins.

## Testing strategy

- Phase A: golden DTO schema tests, CLI parity per refactored command.
- Phase B: manifest validation matrix, install against local fixture repo,
  lockfile tamper/hash-mismatch detection, gate integration tests.
- Phase C: fake plugin binaries as fixtures, permission matrix, mediated-write
  validation parity (over-allocate, archived bucket…), timeouts/malformed
  output/crash handling, e2e `balde <cmd>` + `--json` validity.
- CI runs `plugin check` against `examples/plugins/hello` so docs never drift
  from behavior.
