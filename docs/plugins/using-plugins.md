# Using Plugins

> **Experimental feature.** Plugin support may change or break without
> notice while it is experimental. Nothing you install can touch your budget
> until you explicitly run it, and every write goes through balde's own
> validation.

## Opting in

Plugin commands refuse to run until you opt in, either per command:

```sh
balde plugin list --experimental
```

or for your whole session:

```sh
export BALDE_EXPERIMENTAL=1
balde plugin list
```

Each gated run prints a one-line experimental notice on stderr.

## Installing a plugin

```sh
balde plugin install github.com/user/balde-plugin-vacation
```

Pin a GitHub shorthand to a tag, branch, or commit with `@<revision>`. Balde
records the resolved commit in the lockfile:

```sh
balde plugin install github.com/user/balde-plugin-vacation@v1.2.0
```

Repositories can host multiple plugins below their root. Use `--path` to
select one:

```sh
balde plugin install github.com/egermano/balde --path plugins/vacation
```

`install` accepts any git repository reference (path or URL). It will:

1. Clone the repository into `<budget>/.balde/plugins/src/<name>/`, pinned to
   the exact commit it fetched (recorded in the lockfile).
2. Read and validate the plugin's `balde-plugin.json` manifest — name,
   protocol version, capabilities and declared permissions.
3. Run the manifest's `entrypoint.build` command, if any. This is the only
   code that runs at install time.
4. Verify the `entrypoint.run` artifact exists and record its SHA-256 hash in
   `.balde/plugins/lock.json`.

Installing is the trust grant: you explicitly chose to run this code. After
installing, balde enforces what the plugin declared — nothing more.

## Listing plugins

```sh
balde plugin list
```

Shows each installed plugin's name, version and source.

## Removing a plugin

```sh
balde plugin remove vacation
```

Deletes the plugin's source directory and its lockfile entry.

## The lockfile

`.balde/plugins/lock.json` records, for every installed plugin:

- `name`, `version`, `source` — where it came from
- `sha` — the exact commit the source is pinned to
- `artifact_hash` — SHA-256 of the executable, verified before every run
- `capabilities` and `permissions` — what the plugin declared at install
- `run` — the relative executable path used by the runtime
- `skill` — the optional agent skill copied into `.agents/skills/`

Plugins are **project-local**: they live next to `balde.db`, so each budget
chooses its own plugins. Commit `.balde/plugins/lock.json` if you keep your
budget in git.

## Running plugin commands

Command capabilities appear as normal Balde commands after installation:

```sh
balde vacation plan "Japan" --date 2027-03 --budget 1200000 --json
```

Plugin arguments are passed through unchanged. `--json` is recognized by
Balde and emits the plugin's structured result. Before every run, Balde
recalculates the executable SHA-256 and refuses a tampered artifact.

## Troubleshooting

| Error | Meaning |
|---|---|
| `plugin commands are experimental: opt in with --experimental or BALDE_EXPERIMENTAL=1` | You have not opted in yet |
| `install: clone: ...` | The source could not be cloned — check the path/URL and that `git` is installed |
| `install: read manifest` / `install: manifest: ...` | The repository has no `balde-plugin.json`, or it failed validation |
| `install: plugin X already installed` | Remove it first, or pick another name |
| `remove: plugin X is not installed` | No lockfile entry with that name — check `balde plugin list` |
