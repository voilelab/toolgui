//go:build js && wasm

package tgwasm

import (
	"syscall/js"

	"github.com/voilelab/toolgui/toolgui/tgjson"
	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// EventPrefix is put before the name [Emit] is given, so an app's events
// cannot be taken for the browser's own.
const EventPrefix = "toolgui:"

// ErrNoEventCallback is returned by [Emit] before the page has registered for
// events, or when no executor runs.
var ErrNoEventCallback = tgutil.NewError("no event callback, the page has not called onEvent")

// emitter is the bridge [Emit] sends through: the one [Executor.Run]
// installed.
var emitter *bridge

// Emit dispatch a DOM event named EventPrefix+name on the window of the page
// the app is shown in, with detail, sent as JSON, as its detail. It reaches
// that page only, not other tabs of the site.
//
// It is how an app tells the page something, e.g. an analytics script loaded
// with toolgui-wasm -head:
//
//	tgwasm.Emit("download", map[string]any{"tool": "novel"})
//
//	addEventListener('toolgui:download', (e) => umami.track('download', e.detail))
func Emit(name string, detail any) error {
	if emitter == nil {
		return ErrNoEventCallback
	}

	return emitter.emit(name, detail)
}

func (b *bridge) emit(name string, detail any) error {
	if b.onEvent.IsUndefined() {
		return ErrNoEventCallback
	}

	bs, err := tgjson.Marshal(detail)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	b.onEvent.Invoke(name, string(bs))
	return nil
}

// jsOnEvent register the callback [Emit] hands each event to, as its name and
// its detail as one JSON string.
func (b *bridge) jsOnEvent(this js.Value, args []js.Value) any {
	if len(args) == 0 {
		return nil
	}

	b.onEvent = args[0]
	return nil
}
