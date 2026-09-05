#!/usr/bin/env bash
# Create placeholder web assets so `go build` works without a yarn build.
# toolgui-web/web.go embeds app/build, which is gitignored.
set -euo pipefail

BUILD_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/toolgui-web/app/build"

if [ -f "$BUILD_DIR/index.html" ]; then
	echo "web assets already present, skipping stub"
	exit 0
fi

mkdir -p "$BUILD_DIR/static"
cat > "$BUILD_DIR/index.html" <<'HTML'
<!doctype html>
<title>toolgui: assets not built</title>
<p>Placeholder page. Run <code>task asset_lib asset_app</code> to build the real web assets.</p>
HTML
# go:embed skips dot files, so the static dir needs a visible placeholder.
echo "placeholder" > "$BUILD_DIR/static/placeholder.txt"

echo "stubbed web assets in $BUILD_DIR"
