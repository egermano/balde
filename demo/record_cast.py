#!/usr/bin/env python3
"""Record a real balde session into an asciinema v2 cast file.

Runs the actual CLI in a throwaway budget dir and captures real command
output with timing, emitting demo/balde-quickstart.cast. The typing
animation is cosmetic; every output byte comes from the real binary.
"""

import json
import os
import subprocess
import sys
import tempfile
import time

REPO = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
BALDE = os.path.join(REPO, "bin", "balde")


def main() -> int:
    if not os.path.exists(BALDE):
        subprocess.run(["go", "build", "-o", BALDE, "."], cwd=REPO, check=True)

    cast_path = os.path.join(REPO, "demo", "balde-quickstart.cast")
    os.makedirs(os.path.dirname(cast_path), exist_ok=True)

    events = []
    start = time.time()

    def out(text: str) -> None:
        events.append([round(time.time() - start, 3), "o", text])

    def type_cmd(cmd: str) -> None:
        # cosmetic typing effect for the recorded command line
        out("$ ")
        for i in range(1, len(cmd) + 1):
            out(cmd[i - 1])
            time.sleep(0.025)
        out("\r\n")

    def run(cmd: str, args: list[str], stdin: str = "") -> None:
        type_cmd(cmd)
        proc = subprocess.run(
            [BALDE, *args],
            capture_output=True,
            text=True,
            timeout=30,
        )
        for stream in (proc.stdout, proc.stderr):
            for line in stream.splitlines():
                out(line + "\r\n")
                time.sleep(0.08)
        time.sleep(0.35)

    with tempfile.TemporaryDirectory() as demo:
        os.chdir(demo)
        banner = [
            "balde quickstart — bucket budgeting demo (synthetic data)",
            "",
        ]
        for line in banner:
            out(line + "\r\n")

        run("echo N | balde init", ["init"], stdin="N\n")
        run("balde account add checking checking 100000",
            ["account", "add", "checking", "checking", "100000"])
        run("balde bucket add housing 50000", ["bucket", "add", "housing", "50000"])
        run("balde transaction add 250000 salary 1 1",
            ["transaction", "add", "250000", "salary", "1", "1"])
        run("balde transaction add -50000 rent 1 7",
            ["transaction", "add", "-50000", "rent", "1", "7"])
        run("balde allocate 60000 7 --force", ["allocate", "60000", "7", "--force"])
        run("balde rain", ["rain"])
        run("balde view buckets", ["view", "buckets"])
        run("balde status --json", ["status", "--json"])

    header = {
        "version": 2,
        "width": 80,
        "height": 24,
        "timestamp": int(start),
        "env": {"SHELL": "/bin/sh", "TERM": "xterm-256color"},
        "title": "balde — bucket budgeting CLI quickstart",
    }
    with open(cast_path, "w", encoding="utf-8") as f:
        f.write(json.dumps(header) + "\n")
        for event in events:
            f.write(json.dumps(event) + "\n")

    print(f"wrote {cast_path} ({len(events)} events)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
