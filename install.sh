#!/bin/sh
# balde installer/upgrader — https://github.com/egermano/balde
#
# Install or upgrade with a single command:
#
#   curl -fsSL https://raw.githubusercontent.com/egermano/balde/main/install.sh | bash
#
# Environment overrides:
#   BALDE_VERSION      pin a specific release tag (e.g. v0.1.0-alpha.3)
#   BALDE_INSTALL_DIR  install destination (default: ~/.local/bin, or the
#                      directory of an existing balde installation)
#   BALDE_FORCE=1      reinstall even when the latest version is already installed
#
# The script leaves nothing behind: all downloads happen in a temporary
# directory that is always removed, and the binary is moved into place
# atomically only after its sha256 checksum has been verified.

set -eu

REPO="egermano/balde"
GITHUB="https://github.com/${REPO}"

log()  { printf '%s\n' "$*"; }
info() { printf '==> %s\n' "$*"; }
fail() { printf 'error: %s\n' "$*" >&2; exit 1; }

have() { command -v "$1" >/dev/null 2>&1; }

if ! have curl && ! have wget; then
  fail "curl or wget is required to download balde"
fi

fetch() { # fetch <url> <output-file>
  if have curl; then curl -fsSL -o "$2" "$1"; else wget -qO "$2" "$1"; fi
}

fetch_stdout() { # fetch_stdout <url>
  if have curl; then curl -fsSL "$1"; else wget -qO- "$1"; fi
}

# Resolve the latest release tag from the releases/latest redirect (no GitHub
# API involved). If the project only has prereleases, that URL returns 404 and
# we fall back to the API, which returns the most recent release including
# prereleases.
resolve_latest_tag() {
  _tag=""
  _url=""
  if have curl; then
    _url="$(curl -fsSLo /dev/null -w '%{url_effective}' "${GITHUB}/releases/latest" 2>/dev/null)" || _url=""
  else
    _url="$(wget -q --server-response --spider -O /dev/null "${GITHUB}/releases/latest" 2>&1 |
      sed -n 's/^[[:space:]]*[Ll]ocation: *//p' | tail -n 1)" || _url=""
  fi
  case "$_url" in
    */tag/*) _tag="${_url##*/tag/}" ;;
  esac
  if [ -z "$_tag" ]; then
    _tag="$(fetch_stdout "https://api.github.com/repos/${REPO}/releases?per_page=1" |
      sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p')" || _tag=""
  fi
  [ -n "$_tag" ] || return 1
  printf '%s' "$_tag"
}

# --- detect platform --------------------------------------------------------

os="$(uname -s)"
case "$os" in
  Linux)  os_title="Linux"  ;;
  Darwin) os_title="Darwin" ;;
  *) fail "unsupported operating system '${os}' — install.sh supports Linux and macOS (Windows: use install.ps1)" ;;
esac

arch="$(uname -m)"
case "$arch" in
  x86_64|amd64)  arch="x86_64" ;;
  arm64|aarch64) arch="arm64"  ;;
  *) fail "unsupported architecture '${arch}' on ${os} — supported: x86_64, arm64" ;;
esac

asset="balde_${os_title}_${arch}.tar.gz"

# --- resolve release tag ----------------------------------------------------

if [ -n "${BALDE_VERSION:-}" ]; then
  tag="${BALDE_VERSION}"
else
  info "resolving latest release"
  tag="$(resolve_latest_tag)" || fail "could not resolve the latest balde release"
fi
case "$tag" in
  v*) ;;
  *) tag="v${tag}" ;;
esac

# --- pick install destination ----------------------------------------------

[ -n "${HOME:-}" ] || fail "HOME is not set"

if [ -n "${BALDE_INSTALL_DIR:-}" ]; then
  install_dir="${BALDE_INSTALL_DIR}"
elif command -v balde >/dev/null 2>&1; then
  install_dir="$(dirname -- "$(command -v balde)")"
else
  install_dir="${HOME}/.local/bin"
fi
target="${install_dir}/balde"

old_version=""
if [ -x "${target}" ]; then
  old_version="$("${target}" --version 2>/dev/null | awk 'NR == 1 {print $3; exit}')"
fi

# --- no-op when already up to date ------------------------------------------

if [ "${BALDE_FORCE:-0}" != "1" ] && [ -n "$old_version" ]; then
  if [ "${old_version#v}" = "${tag#v}" ]; then
    info "balde ${tag} is already installed at ${target} — up to date, nothing to do"
    exit 0
  fi
fi

# --- download and verify ----------------------------------------------------

tmp_dir="$(mktemp -d)" || fail "could not create a temporary directory"
trap 'rm -rf "${tmp_dir}"' EXIT
trap 'exit 1' INT TERM

base_url="${GITHUB}/releases/download/${tag}"
info "downloading balde ${tag} (${os_title}/${arch})"
fetch "${base_url}/${asset}" "${tmp_dir}/${asset}" ||
  fail "download failed: ${base_url}/${asset}"
fetch "${base_url}/checksums.txt" "${tmp_dir}/checksums.txt" ||
  fail "download failed: ${base_url}/checksums.txt"

if have sha256sum; then
  shatool="sha256sum"
elif have shasum; then
  shatool="shasum -a 256"
else
  fail "sha256sum or shasum is required to verify the download"
fi

expected="$(awk -v a="${asset}" '$2 == a {print $1; exit}' "${tmp_dir}/checksums.txt")"
[ -n "$expected" ] || fail "checksum for ${asset} not found in checksums.txt"
actual="$(${shatool} "${tmp_dir}/${asset}" | awk '{print $1; exit}')"
if [ "$actual" != "$expected" ]; then
  fail "checksum mismatch for ${asset}
  expected: ${expected}
  actual:   ${actual}"
fi

tar -xzf "${tmp_dir}/${asset}" -C "${tmp_dir}" || fail "failed to extract ${asset}"
[ -f "${tmp_dir}/balde" ] || fail "archive did not contain a balde binary"

# --- install atomically -----------------------------------------------------

if [ ! -d "${install_dir}" ]; then
  mkdir -p "${install_dir}" || fail "could not create install directory: ${install_dir}"
fi
if [ ! -w "${install_dir}" ]; then
  fail "no write permission for ${install_dir}
re-run with sudo, or set BALDE_INSTALL_DIR to a writable directory on your PATH"
fi

info "installing balde ${tag} -> ${target}"
tmp_target="${install_dir}/.balde.tmp.$$"
cp "${tmp_dir}/balde" "${tmp_target}" || { rm -f "${tmp_target}"; fail "could not write ${tmp_target}"; }
chmod 0755 "${tmp_target}"
mv -f "${tmp_target}" "${target}" || { rm -f "${tmp_target}"; fail "could not move balde into place at ${target}"; }

info "balde ${tag} installed at ${target}${old_version:+ (was ${old_version})}"

case ":${PATH}:" in
  *":${install_dir}:"*) ;;
  *)
    log ""
    log "NOTE: ${install_dir} is not on your PATH."
    log "Add it to your shell profile (e.g. ~/.bashrc or ~/.zshrc):"
    log "    export PATH=\"${install_dir}:\$PATH\""
    ;;
esac
