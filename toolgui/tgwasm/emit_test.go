//go:build js && wasm

package tgwasm

import (
	"errors"
	"syscall/js"
	"testing"
)

// TestEmitHandsTheEventToThePage checks Emit reaches the callback the page
// registered, with the detail as JSON.
func TestEmitHandsTheEventToThePage(t *testing.T) {
	b := newBridge(testApp())

	err := b.emit("download", nil)
	if !errors.Is(err, ErrNoEventCallback) {
		t.Fatalf("expect ErrNoEventCallback before onEvent, got %v", err)
	}

	var name, detail string
	callback := js.FuncOf(func(this js.Value, args []js.Value) any {
		name, detail = args[0].String(), args[1].String()
		return nil
	})
	defer callback.Release()
	b.jsOnEvent(js.Undefined(), []js.Value{callback.Value})

	err = b.emit("download", map[string]any{"tool": "novel"})
	if err != nil {
		t.Fatal(err)
	}

	if name != "download" || detail != `{"tool":"novel"}` {
		t.Errorf("expect download {\"tool\":\"novel\"}, got %s %s", name, detail)
	}
}
