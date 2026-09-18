//go:build js && wasm

package tgwasm

import "syscall/js"

// EmbedName is the global the page sets the display mode on, before the wasm
// program starts.
const EmbedName = "toolguiEmbed"

// Embedded reports whether the app is being shown inside a frame, which is
// what the frontend's `?embed` means: no nav, no chrome, just the page.
//
// It is for a page that is laid out differently in a frame -- narrower, or
// without what the page around it already carries. Everything else about
// embedding is the frontend's, and an app need not ask.
//
// A host that says nothing is not embedding, so this is false wherever the
// flag was never set.
func Embedded() bool {
	v := js.Global().Get(EmbedName)
	if v.Type() != js.TypeBoolean {
		return false
	}

	return v.Bool()
}
