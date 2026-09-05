#!/usr/bin/env bash
# Create placeholder web assets so `go build` works without a yarn build.
# toolgui-web/web.go and toolgui-wails/assets.go embed build output that is
# gitignored.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="$ROOT/toolgui-web/app/build"
WAILS_DIR="$ROOT/toolgui-wails/frontend/dist"

if [ -f "$BUILD_DIR/index.html" ]; then
	echo "web assets already present, skipping stub"
else
	mkdir -p "$BUILD_DIR/static"
	cat > "$BUILD_DIR/index.html" <<'HTML'
<!doctype html>
<title>toolgui: assets not built</title>
<p>Placeholder page. Run <code>task asset_lib asset_app</code> to build the real web assets.</p>
HTML
	# go:embed skips dot files, so the static dir needs a visible placeholder.
	echo "placeholder" > "$BUILD_DIR/static/placeholder.txt"

	echo "stubbed web assets in $BUILD_DIR"
fi

if [ -f "$WAILS_DIR/index.html" ]; then
	echo "wails assets already present, skipping stub"
else
	mkdir -p "$WAILS_DIR"
	cat > "$WAILS_DIR/index.html" <<'HTML'
<!doctype html>
<title>toolgui: assets not built</title>
<p>Placeholder page. Run <code>task asset_wails</code> to build the real desktop assets.</p>
HTML

	echo "stubbed wails assets in $WAILS_DIR"
fi
