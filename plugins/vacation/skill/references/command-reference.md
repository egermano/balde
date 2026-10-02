# Vacation Plugin Command Reference

> The plugin is experimental. Export `BALDE_EXPERIMENTAL=1` before invoking
> its commands.

## `balde vacation plan <destination> --date YYYY-MM --budget <cents>`

Creates `vacation-<destination-slug>-<date>` with target `<cents>`, then
allocates the first monthly installment from rain.

| Input | Rule |
|---|---|
| `destination` | Required; becomes part of the bucket name |
| `--date` | Required future month in `YYYY-MM` format |
| `--budget` | Required positive integer cents |

The regular installment is `ceil(budget / calendar_months_until_date)`. The
last installment is reduced when necessary so total allocations equal the
target exactly.

Example:

```sh
balde vacation plan "Japan" --date 2027-03 --budget 1200000
```

## `balde vacation schedule <destination> --date YYYY-MM`

Looks up the date-encoded bucket and reports remaining cents plus the next
`balde allocate` command.

```sh
balde vacation schedule "Japan" --date 2027-03
```

## Agent response template

After `plan`, respond with:

```markdown
Created vacation bucket: `<bucket name>`
- Target: `<budget>` cents
- Allocated today: `<amount>` cents
- Monthly allocation: `<amount>` cents
- Remaining allocations: `<count>`
```

If the final installment differs, include it explicitly.
