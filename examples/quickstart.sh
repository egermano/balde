#!/bin/sh
# Balde quickstart — a complete budgeting flow with synthetic data.
#
# Run from anywhere; it builds the CLI if needed, creates a throwaway
# budget in a temp directory and walks through the whole method:
# init → account → buckets → income → expense → allocate → rain → status.
#
# No real financial data is used or required.

set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
BALDE="$repo_root/bin/balde"

if [ ! -x "$BALDE" ]; then
    echo "==> Building balde (first run)"
    (cd "$repo_root" && go build -o bin/balde .)
fi

demo_dir=$(mktemp -d)
trap 'rm -rf "$demo_dir"' EXIT
cd "$demo_dir"

echo "==> balde init (no encryption for the demo)"
echo "N" | "$BALDE" init

echo
echo "==> account: checking, starting balance 100000 cents (= \$1000.00)"
"$BALDE" account add checking checking 100000

echo
echo "==> two custom buckets on top of the 6 defaults (amounts in cents)"
"$BALDE" bucket add housing 50000
"$BALDE" bucket add travel 12000

echo
echo "==> income: salary 250000 cents to account 1, bucket 1 (financial freedom)"
"$BALDE" transaction add 250000 "salary" 1 1

echo
echo "==> expense: rent -50000 cents from account 1, bucket 7 (housing)"
"$BALDE" transaction add -50000 "rent" 1 7

echo
echo "==> allocate 60000 cents of rain to bucket 7 (housing)"
"$BALDE" allocate 60000 7 --force

echo
echo "==> balde rain (money still unallocated)"
"$BALDE" rain

echo
echo "==> balde view buckets"
"$BALDE" view buckets

echo
echo "==> balde status --json (machine-readable snapshot)"
"$BALDE" status --json

echo
echo "Done. The demo budget lived in: $demo_dir (now removed)."
echo "Start a real budget anywhere with: balde init"
