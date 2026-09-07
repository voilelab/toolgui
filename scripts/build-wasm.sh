#!/usr/bin/env bash
# Assemble the static site for a ToolGUI wasm app: the frontend build, the Go
# program compiled to wasm, and the wasm_exec.js of the toolchain that
# compiled it.
#
#   ./scripts/build-wasm.sh <go package> <out dir>
#
# Build the frontend first with `task asset_wasm`.
set -euo pipefail

if [ $# -ne 2 ]; then
	echo "usage: $0 <go package> <out dir>" >&2
	exit 2
fi

PKG="$1"
OUT="$2"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FRONTEND="$ROOT/toolgui-web/wasm/build"

if [ ! -f "$FRONTEND/index.html" ]; then
	echo "no frontend in $FRONTEND, run 'task asset_wasm' first" >&2
	exit 1
fi

rm -rf "$OUT"
mkdir -p "$OUT"

cp -r "$FRONTEND/." "$OUT/"

GOOS=js GOARCH=wasm go build -o "$OUT/app.wasm" "$PKG"

# Not vendored: it is the runtime shim of the toolchain that just built the
# binary, and the two have to match.
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$OUT/"

echo "built $OUT"
