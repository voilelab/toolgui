package tgwails

import _ "embed"

// JS is the bundled frontend. Wails v1 injects it into the webview, so it has
// to be a single file. Build it with `task asset_wails`.
//
//go:embed assets/app.js
var JS string

// CSS is the bundled stylesheet, single file for the same reason.
//
//go:embed assets/app.css
var CSS string
