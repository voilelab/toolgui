package tgwails

import "embed"

// Assets is the built frontend. Wails serves it from its own origin, so it
// stays an ordinary directory of files rather than one inlined bundle.
// Build it with `task asset_wails`.
//
//go:embed all:frontend/dist
var Assets embed.FS
