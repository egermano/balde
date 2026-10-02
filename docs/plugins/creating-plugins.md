# Creating Plugins

> **Experimental feature.** Protocol version 1 can change before plugins are
> declared stable.

A plugin is a git repository (or a subdirectory in one) with a
`balde-plugin.json` manifest and executable. Any language can participate as
long as its executable reads newline-delimited JSON-RPC 2.0 from stdin and
writes it to stdout.

## Minimal manifest

```json
{
  "name": "vacation",
  "version": "0.1.0",
  "protocol": 1,
  "capabilities": [{
    "type": "command",
    "name": "vacation",
    "description": "Plan vacation savings"
  }],
  "permissions": {
    "read": ["buckets"],
    "write": ["buckets", "allocate"]
  },
  "entrypoint": {
    "build": "go build -o bin/vacation .",
    "run": "bin/vacation"
  },
  "skill": "skill"
}
```

| Field | Required | Meaning |
|---|---|---|
| `name` | yes | Lowercase letters, digits and dashes. It names the install directory. |
| `version` | yes | Plugin release version. |
| `protocol` | yes | `1` today. |
| `capabilities` | yes | `command`, `importer`, `reporter`, or `connector`. Runtime v1 exposes `command`. |
| `permissions` | yes | Least-privilege read/write scopes; undeclared calls receive `E_PERM`. |
| `entrypoint.run` | yes | Relative executable path inside the plugin. Absolute and `..` paths are rejected. |
| `entrypoint.build` | no | Shell command run once during explicit installation. |
| `skill` | no | Skill directory (or `SKILL.md`) copied to the project agent-skill directory. |

Valid read scopes are `accounts`, `buckets`, and `transactions`. Valid write
scopes are `transactions`, `buckets`, and `allocate`.

## Agent skill

Set `"skill": "skill"` to ship a skill directory with `skill/SKILL.md`.
Its YAML frontmatter must use `name: balde-<plugin-name>` and have a non-empty
`description`. Installation copies it to:

```
.agents/skills/balde-<plugin-name>/
```

The destination contains a Balde ownership marker. A plugin cannot overwrite
a user-authored skill with the same name, and removal only removes skill
directories carrying that marker. Keep agent instructions, command examples,
and optional `references/` together in this directory.

The bundled [`plugins/vacation`](../../plugins/vacation) plugin is the
reference implementation: a standalone Go module, manifest, protocol client,
tests, and agent skill.

Read [the protocol reference](protocol.md) before implementing I/O.
