#!/bin/sh

set -eu

root_dir="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT INT TERM

asset="balde_Linux_x86_64.tar.gz"
release_dir="$tmp_dir/releases/download/v1.2.3"
mkdir -p "$release_dir/archive" "$tmp_dir/bin" "$tmp_dir/install"

cat >"$release_dir/archive/balde" <<'EOF'
#!/bin/sh
printf '%s\n' 'balde version 1.2.3'
EOF
chmod 0755 "$release_dir/archive/balde"
tar -czf "$release_dir/$asset" -C "$release_dir/archive" balde
sha256sum "$release_dir/$asset" | sed "s|$release_dir/||" >"$release_dir/checksums.txt"

cat >"$tmp_dir/bin/uname" <<'EOF'
#!/bin/sh
case "$1" in
  -s) printf '%s\n' Linux ;;
  -m) printf '%s\n' x86_64 ;;
esac
EOF
chmod 0755 "$tmp_dir/bin/uname"

cat >"$tmp_dir/bin/curl" <<EOF
#!/bin/sh
url=""
output=""
while [ \$# -gt 0 ]; do
  case "\$1" in
    -o) output="\$2"; shift 2 ;;
    -w) shift 2 ;;
    -*) shift ;;
    *) url="\$1"; shift ;;
  esac
done
path="\${url#https://github.com/egermano/balde}"
cp "$tmp_dir\$path" "\$output"
EOF
chmod 0755 "$tmp_dir/bin/curl"

output="$(
  PATH="$tmp_dir/bin:/usr/bin:/bin" \
    BALDE_VERSION=v1.2.3 \
    BALDE_INSTALL_DIR="$tmp_dir/install" \
    sh "$root_dir/install.sh"
)"

test -x "$tmp_dir/install/balde"
test "$("$tmp_dir/install/balde" --version)" = "balde version 1.2.3"
printf '%s' "$output" | grep -q 'balde v1.2.3 installed'

rm -rf "$release_dir"
output="$(
  PATH="$tmp_dir/install:$tmp_dir/bin:/usr/bin:/bin" \
    BALDE_VERSION=v1.2.3 \
    BALDE_INSTALL_DIR="$tmp_dir/install" \
    sh "$root_dir/install.sh"
)"
printf '%s' "$output" | grep -q 'up to date, nothing to do'

mkdir -p "$release_dir/archive"
cp "$tmp_dir/install/balde" "$release_dir/archive/balde"
tar -czf "$release_dir/$asset" -C "$release_dir/archive" balde
printf '%s\n' 'not-a-checksum  balde_Linux_x86_64.tar.gz' >"$release_dir/checksums.txt"
if PATH="$tmp_dir/bin:/usr/bin:/bin" \
  BALDE_VERSION=v1.2.3 \
  BALDE_INSTALL_DIR="$tmp_dir/install" \
  BALDE_FORCE=1 \
  sh "$root_dir/install.sh" >"$tmp_dir/failure.log" 2>&1; then
  printf '%s\n' 'installer accepted an invalid checksum' >&2
  exit 1
fi
grep -q 'checksum mismatch' "$tmp_dir/failure.log"
test "$("$tmp_dir/install/balde" --version)" = "balde version 1.2.3"

rm -f "$release_dir/$asset"
if PATH="$tmp_dir/bin:/usr/bin:/bin" \
  BALDE_VERSION=v1.2.3 \
  BALDE_INSTALL_DIR="$tmp_dir/install" \
  BALDE_FORCE=1 \
  sh "$root_dir/install.sh" >"$tmp_dir/missing.log" 2>&1; then
  printf '%s\n' 'installer accepted a missing release asset' >&2
  exit 1
fi
grep -q 'download failed' "$tmp_dir/missing.log"
test "$("$tmp_dir/install/balde" --version)" = "balde version 1.2.3"

printf '%s\n' 'installer smoke test passed'
