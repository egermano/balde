# Recommended Plugins

> **Experimental feature.** Everything here is experimental while the plugin
> system matures.

This is a curated list of balde plugins. None are listed yet — the plugin
system just shipped. This page will grow as the ecosystem does.

## Inclusion criteria

A plugin is listed when it:

1. Installs with `balde plugin install <github-ref>` with no manual steps
2. Has a valid `balde-plugin.json` with correctly declared permissions
   (least privilege — no blanket read/write)
3. Has automated tests, including protocol conformance (`balde plugin check`)
4. Is actively maintained (responds to issues, works with the current
   protocol version)

## The list

| Plugin | Capabilities | Permissions | Description |
|---|---|---|---|
| [Vacation planner](../../plugins/vacation) | `command` | read `accounts`, `buckets`; write `buckets`, `allocate` | Creates a date-encoded vacation bucket, funds the first monthly installment when rain covers it, and reports the remaining schedule. Install with `balde plugin install github.com/egermano/balde --path plugins/vacation`. |

## Submitting a plugin

Open a pull request against this repository adding your plugin to the table
above. Include:

- The exact install command
- A short description of what it does
- Which capabilities and permissions it declares and why each is needed
