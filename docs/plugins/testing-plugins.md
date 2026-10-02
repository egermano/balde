# Testing Plugins

Test three boundaries:

1. **Pure planning logic** — dates, rounding, names, and validation without a
   process or database.
2. **Protocol behavior** — handshake, command responses, and host-call error
   handling against a fake host.
3. **Installed behavior** — install a local fixture repository, run the
   registered command against a temporary Balde budget, and assert the
   lockfile and agent skill outcome.

The bundled vacation plugin follows this split:

```sh
cd plugins/vacation
go test .
go build -o bin/vacation .
```

At the repository level, run the host integration suite:

```sh
go test ./plugin/ ./cli/
```

## Agent-skill evaluation

Plugin skills should have realistic prompts under `skill/evals/evals.json`.
The vacation plugin includes planning, casual phrasing, and progress-check
prompts. When evolving a skill, use the skill-creator workflow: compare
with-skill and baseline runs, review outputs with the evaluator, and adjust
instructions based on observed failures rather than adding speculative rules.
