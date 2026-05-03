#!/usr/bin/env bash
#
# install irdocker + bash completion
#
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NAME="irdocker"

SRC_GO="${ROOT}/main.go"
COMP_SRC="${ROOT}/completions/irdocker.bash"

PREFIX="${PREFIX:-/usr/local}"
BINDIR="${BINDIR:-$PREFIX/bin}"
COMPDIR="${COMPDIR:-}"
DRY="${DRY_RUN:-0}"

usage() {
    cat <<'EOF' >&2
Usage: ./install.sh [options]

  PREFIX=/path     default: /usr/local  (user: PREFIX=~/.local ./install.sh)
  BINDIR=...       default: PREFIX/bin
  COMPDIR=...      auto if unset
  DRY_RUN=1        print actions only

EOF
    exit 1
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help) usage ;;
        *) usage ;;
    esac
done

[[ -f "$SRC_GO"   ]] || { echo "install: missing ${SRC_GO}" >&2; exit 1; }
[[ -f "$COMP_SRC" ]] || { echo "install: missing ${COMP_SRC}" >&2; exit 1; }

pick_compdir() {
    if [[ -n "$COMPDIR" ]]; then
        echo "$COMPDIR"
        return
    fi

    for d in \
        "${PREFIX}/share/bash-completion/completions" \
        /usr/share/bash-completion/completions \
        /usr/local/share/bash-completion/completions \
        /etc/bash_completion.d
    do
        if [[ -d "$d" ]]; then
            echo "$d"
            return
        fi
    done

    echo "${PREFIX}/share/bash-completion/completions"
}

COMPDIR="$(pick_compdir)"

run() {
    if [[ "$DRY" == "1" ]]; then
        printf '[dry-run] '
        printf '%q ' "$@"
        echo
    else
        "$@"
    fi
}

echo "irdocker install"
echo "  bin:     ${BINDIR}/${NAME}"
echo "  compl:   ${COMPDIR}/${NAME}"
echo

run mkdir -p "$BINDIR" "$COMPDIR"

# -------------------------
# FIX: build in temp file
# -------------------------
TMP_BIN="$(mktemp)"

run go build -o "$TMP_BIN" "$SRC_GO"

run install -m 0755 "$TMP_BIN" "${BINDIR}/${NAME}"
run rm -f "$TMP_BIN"

run install -m 0644 "$COMP_SRC" "${COMPDIR}/${NAME}"

if [[ "$DRY" != "1" ]]; then
    echo "done."
    echo "Run this to activate completion:"
    echo "  source ${COMPDIR}/${NAME}"
fi