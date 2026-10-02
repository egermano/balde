# Balde Plugins

> **Experimental feature.** Plugins let anyone extend balde — new commands,
> importers, reports — without changing balde itself.

## Documentation

- [Using plugins](using-plugins.md) — install, list, remove, the lockfile,
  troubleshooting
- [Recommended plugins](recommended.md) — curated community list
- Creating plugins, the plugin protocol, and testing plugins — *coming with
  the plugin runtime (Phase C)*

## Quickstart

```sh
export BALDE_EXPERIMENTAL=1
balde plugin install github.com/user/balde-plugin-example
balde plugin list
```

## What plugins are

A plugin is a git repository with a `balde-plugin.json` manifest and an
executable. It runs as a separate process and talks to balde over a
versioned JSON protocol — it never touches your database file, your
password, or anything it did not declare in its manifest permissions.
