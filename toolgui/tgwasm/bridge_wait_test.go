//go:build js && wasm

package tgwasm

import (
	"strings"
	"syscall/js"
	"testing"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// timerApp's pages wait for a JavaScript timer and ignore their context, the
// way a page func awaiting IndexedDB does. Each finished run sends its page
// name to runs.
func timerApp(runs chan<- string) *tgframe.App {
	app := tgframe.NewApp()

	page := func(name string) tgframe.RunFunc {
		return func(p *tgframe.Params) error {
			fired := make(chan struct{})
			f := js.FuncOf(func(js.Value, []js.Value) any {
				close(fired)
				return nil
			})
			defer f.Release()

			js.Global().Call("setTimeout", f, 10)
			<-fired

			runs <- name
			return nil
		}
	}
	app.AddPage("index", "Index", page("index"))
	app.AddPage("other", "Other", page("other"))

	return app
}

// TestCallsDoNotWaitForTheRun checks start and update return while a run is
// still waiting on JavaScript. They are called from JavaScript, so a call that
// waited for the run would keep the timer it needs from ever firing.
func TestCallsDoNotWaitForTheRun(t *testing.T) {
	runs := make(chan string, 8)
	b := newBridge(timerApp(runs))
	defer stopped(b)()

	onPack := js.FuncOf(func(js.Value, []js.Value) any { return nil })
	defer onPack.Release()
	b.jsOnPack(js.Undefined(), []js.Value{onPack.Value})

	start := js.FuncOf(b.jsStart)
	defer start.Release()
	update := js.FuncOf(b.jsUpdate)
	defer update.Release()

	start.Invoke("index")
	update.Invoke(`{"type":""}`)
	start.Invoke("other")

	for _, want := range []string{"index", "index", "other"} {
		select {
		case got := <-runs:
			if got != want {
				t.Fatalf("run of %q, want %q", got, want)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("no run of %q", want)
		}
	}
}

// TestFatalStartComesAfterQueuedRuns checks a start on an unknown page reports
// after the runs queued before it. A run's ready pack clears the error shown,
// so a fatal sent ahead of one would be lost.
func TestFatalStartComesAfterQueuedRuns(t *testing.T) {
	runs := make(chan string, 8)
	b := newBridge(timerApp(runs))
	defer stopped(b)()

	var packs []string
	onPack := js.FuncOf(func(_ js.Value, args []js.Value) any {
		packs = append(packs, args[0].String())
		return nil
	})
	defer onPack.Release()
	b.jsOnPack(js.Undefined(), []js.Value{onPack.Value})

	start := js.FuncOf(b.jsStart)
	defer start.Release()
	update := js.FuncOf(b.jsUpdate)
	defer update.Release()

	start.Invoke("index")
	update.Invoke(`{"type":""}`)
	start.Invoke("nowhere")
	settle(b)

	if len(packs) == 0 || !strings.Contains(packs[len(packs)-1], `"fatal":true`) {
		t.Errorf("last pack = %v, want the fatal result", packs)
	}
}
