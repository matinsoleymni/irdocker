DRY="${DRY_RUN:-0}"

run() {
  if [[ "$DRY" == "1" ]]; then
    printf '[dry-run] '; printf '%q ' "$@"; echo
  else
    "$@"
  fi
}
# Pick completion dir
pick_compdir() {
  if [[ -n "$COMPDIR" ]]; then
    echo "$COMPDIR"
    return
  fi
  for d in \
    "$PREFIX/share/bash-completion/completions" \
    /usr/share/bash-completion/completions \
    /usr/local/share/bash-completion/completions \
    /etc/bash_completion.d
  do
    if [[ -d "$d" ]]; then
      echo "$d"
      return
    fi
  done
  echo "$PREFIX/share/bash-completion/completions"
}

COMPDIR="$(pick_compdir)"
# Check for required files
[[ -f "$SRC_GO" ]] || { echo "install: missing ${SRC_GO}" >&2; exit 1; }
[[ -f "$COMP_SRC" ]] || { echo "install: missing ${COMP_SRC}" >&2; exit 1; }

# Parse arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help) usage ;;
    *) usage ;;
  esac
done
PREFIX="${PREFIX:-/usr/local}"
BINDIR="${BINDIR:-$PREFIX/bin}"
COMPDIR="${COMPDIR:-}"

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

#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NAME="irdocker"
SRC_GO="${ROOT}/main.go"
COMP_SRC="${ROOT}/completions/irdocker.bash"

echo "irdocker install"
echo "  bin:     ${BINDIR}/${NAME}"
echo "  compl:   ${COMPDIR}/${NAME}"
echo

run mkdir -p "$BINDIR" "$COMPDIR"
run go build -o "${BINDIR}/${NAME}" "$SRC_GO"
run install -m 0755 "${BINDIR}/${NAME}" "${BINDIR}/${NAME}"
run install -m 0644 "$COMP_SRC" "${COMPDIR}/${NAME}"

if [[ "$DRY" != "1" ]]; then
    cat <<EOF

  done. ensure ${BINDIR} is in PATH.
  completion: new shell, or: source ${COMPDIR}/${NAME}

EOF
fi
