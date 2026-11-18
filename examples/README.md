# Examples

## Quickstart

[`quickstart.sh`](quickstart.sh) walks through the complete bucket-budgeting
flow with **synthetic data only** — safe to run anywhere:

```sh
git clone https://github.com/egermano/balde.git
cd balde
./examples/quickstart.sh
```

It builds the CLI if needed, creates a throwaway budget in a temp directory
(removed at the end) and exercises:

1. `balde init` — budget with the 6 default buckets
2. `balde account add` — a checking account with a starting balance
3. `balde bucket add` — custom buckets (housing, travel)
4. `balde transaction add` — income and an expense (integer cents)
5. `balde allocate` — moving rain into a bucket
6. `balde rain` / `balde view buckets` / `balde status --json` — inspecting state

### Expected output (abridged)

```
==> account: checking, starting balance 100000 cents (= $1000.00)
Account created: checking (checking) balance=100000 id=1

==> two custom buckets on top of the 6 defaults (amounts in cents)
Bucket created: housing target=50000 id=7
Bucket created: travel target=12000 id=8

==> income: salary 250000 cents to account 1, bucket 1 (financial freedom)
Transaction created: amount=250000 desc=salary id=1

==> expense: rent -50000 cents from account 1, bucket 7 (housing)
Transaction created: amount=-50000 desc=rent id=2

==> allocate 60000 cents of rain to bucket 7 (housing)
Allocated 60000 cents to bucket 7

==> balde rain (money still unallocated)
Rain (unallocated): 240000 cents

==> balde view buckets
4	comfort	target=0	balance=0	fill=Not set
1	financial freedom	target=0	balance=250000	fill=Not set
2	fixed costs	target=0	balance=0	fill=Not set
6	goals	target=0	balance=0	fill=Not set
7	housing	target=50000	balance=10000	fill=20.0%
5	knowledge	target=0	balance=0	fill=Not set
3	pleasures	target=0	balance=0	fill=Not set
8	travel	target=12000	balance=0	fill=0.0%
```

The rain math: accounts hold `300000` cents (100000 starting + 250000
salary − 50000 rent) and buckets hold `260000` (250000 salary in financial
freedom + 10000 net in housing: −50000 rent + 60000 allocated), so
`rain = 300000 − 260000 = 40000`. Every number above is printed by the CLI
itself — run the script and follow along.

> **Note on data:** quickstart uses synthetic data exclusively. When an AI
> agent operates a real budget for you, its tools decide what gets read and
> sent — see the
> [data privacy note](../README.md#data-privacy-local-first--llms) in the
> README.
