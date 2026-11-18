# Contributing to Balde

Thanks for considering a contribution! Balde is a small, focused CLI — here is everything you need to get a change merged.

## Talk first (for anything non-trivial)

Open an issue before writing code for bug fixes you can't reproduce from a test, or any new feature. A short conversation saves everyone time when the design doesn't fit. Small fixes (typos, docs, obvious bugs with a clear failing test) can go straight to a PR.

## Development environment

- Go 1.26+
- Make (optional, for convenience targets)

```sh
git clone https://github.com/egermano/balde.git
cd balde
go mod download
make test          # full suite
make lint          # go vet + gofmt
```

There are no other dependencies — SQLite is pure Go (`modernc.org/sqlite`, no CGO).

## How to work on Balde

### Tests come first (TDD)

Every behavior change follows red → green → refactor:

1. **Red** — write a failing test that captures the behavior you want
2. **Green** — the minimum code that makes it pass
3. **Refactor** — clean up while staying green

CI runs the full suite on every PR; a PR without tests for new behavior will be asked for them.

### Tests must be environment-independent

A test that only passes on your machine is a bug. In particular:

- never depend on the current date (compute relative dates; see `TestVacationPluginInstallAndPlanEndToEnd` for the pattern)
- never depend on network, `git` state, or absolute paths
- use synthetic data only — **never real financial data**, not even "harmless" fragments

### Architecture: respect the dependency rule

```
cli ──→ app ──→ core ←── store
```

- `core` is a pure library: no I/O, no SQLite, no user-facing strings
- `store` implements the `Store` interface defined in `core`
- `app` wraps core with the JSON DTOs; only `cli` and the plugin runtime touch user I/O

If your change needs `core` to import something, stop and open an issue first. [`AGENTS.md`](AGENTS.md) has the full map.

### Money is integer cents

Never floats, anywhere. Amounts are `int64` cents from CLI parse to SQLite.

## Pull requests

1. Fork, then create a branch (`feat/…`, `fix/…`, `docs/…`)
2. Make the change with tests (see above)
3. `make lint && make test-short` must pass locally
4. Open a PR against `main` using the PR template
5. Conventional commits (`fix: …`, `feat: …`, `docs: …`) — history stays readable

All changes to `main` go through a PR, including the maintainer's. Keep PRs small and single-purpose; stack them if a bigger change needs slicing.

## Reporting bugs

Use the bug issue template. Include the balde version (`balde --version`), OS/arch, and a reproduction with **synthetic data**. Security issues go to [SECURITY.md](SECURITY.md) — never a public issue.

## Licensing

By contributing you agree that your contributions are licensed under the [MIT License](LICENSE).
