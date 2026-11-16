# Security Policy

Balde is in alpha. Security reports are welcome and taken seriously.

## Supported versions

| Version | Supported |
| ------- | ---------- |
| latest `v0.1.0-alpha.x` release | ✅ |
| `main` branch | ✅ |
| older alpha releases | ❌ |

Balde moves fast during alpha; there are no backports. Upgrade to the latest
release — `install.sh`/`install.ps1` do this in one command.

## Reporting a vulnerability

**Do not open a public issue for security problems.**

Use GitHub's **private vulnerability reporting**: go to the
[Security tab](https://github.com/egermano/balde/security) and click
**"Report a vulnerability"**.

Include when possible:

- Balde version (`balde --version` or the binary's build info)
- OS/architecture
- Steps to reproduce, with synthetic data only (never real financial data)
- Impact assessment (e.g., budget data exposure, encryption bypass)

## What to expect

Balde is maintained by one person, so there is no guaranteed response time.
The realistic expectation: acknowledgment within a few days, a fix depending
on severity, and credit in the release notes if you wish.

## Scope

In scope:

- The `balde` CLI and its command surface
- Encryption at rest (AES-256-GCM) and session handling (`unlock`/`lock`)
- The plugin system runtime (subprocess mediation, install path handling)
- Install scripts (`install.sh`, `install.ps1`)

Out of scope:

- Compromised machines: if an attacker controls your OS user, no on-disk
  encryption password helps
- Loss of `balde.db`/`balde.json` files — back them up; Balde has no built-in
  backup command yet (see roadmap)
- Forgotten passwords: encrypted budgets cannot be recovered without the
  password, by design
