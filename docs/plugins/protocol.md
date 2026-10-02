# Plugin Protocol v1

> **Experimental.** This is the canonical protocol contract for version 1.

The host starts the verified executable with:

```text
<entrypoint> --balde-protocol 1
```

Communication is newline-delimited JSON-RPC 2.0 on stdin/stdout. Keep stdout
for protocol messages only; write diagnostics to stderr. Balde provides a
clean environment (`PATH`, `HOME`, `TMPDIR`, `LANG`) and never supplies a DB
path, password, or session secret.

## Lifecycle

1. Host → plugin: `initialize` with `{protocol: 1, budget: {...}}`.
2. Plugin → host: result containing at least `{name, protocol: 1}`.
3. Host → plugin: `command/execute` with `{command, args}`.
4. Plugin optionally makes host calls while handling the command.
5. Plugin → host: `{text, json?}` result for `command/execute`.

`budget` contains only currency symbol/separators and allocation frequency.

## Host functions

| Method | Manifest permission | Params | Result |
|---|---|---|---|
| `host/listAccounts` | read `accounts` | `{}` | account DTO array |
| `host/listBuckets` | read `buckets` | `{}` | bucket DTO array |
| `host/listTransactions` | read `transactions` | `{}` | transaction DTO array |
| `host/addBucket` | write `buckets` | `{name, target}` | bucket DTO |
| `host/allocate` | write `allocate` | `{bucket_id, amount}` | `{}` |

Amounts are integer cents. Writes run through the same `app` service used by
the CLI, including duplicate and eight-bucket validation.

Errors use a JSON-RPC error object. Important messages: `E_PERM` (permission
not declared), `E_VALIDATION` (domain validation failed), `E_PARAMS` (bad
JSON), and `E_METHOD` (unknown host method).

## Command result

Return both human and agent-friendly forms where possible:

```json
{
  "text": "Created vacation-japan-2027-03. Allocated 240000 cents now.",
  "json": {"bucket_name": "vacation-japan-2027-03", "allocated_now": 240000}
}
```

Balde prints `text` normally and emits `json` for `--json`.
