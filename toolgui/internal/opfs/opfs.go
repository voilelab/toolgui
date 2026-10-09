//go:build js && wasm

// Package opfs is what tgframe's upload store and tgwasm's Store share for
// reaching the browser's origin private file system from Go.
//
// Only reads and writes on a sync access handle are synchronous. Everything
// else answers with a promise, and [Await] must not be called from a [js.Func]
// callback: the event loop is stopped for the length of one, so the promise
// would never settle.
package opfs

import (
	"fmt"
	"log/slog"
	"syscall/js"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// Create and Recursive are the option objects the directory calls take. They
// are never written to, so one of each is enough.
var (
	Create    = map[string]any{"create": true}
	Recursive = map[string]any{"recursive": true}

	Uint8Array = js.Global().Get("Uint8Array")
)

// At is the {at: offset} a read or a write takes. The offset crosses as a
// float64 because [js.ValueOf] takes no int64, and no browser will hold a file
// anywhere near where that loses a byte.
func At(off int64) map[string]any {
	return map[string]any{"at": float64(off)}
}

// Call calls a method and returns what JavaScript threw as an error instead of
// panicking with it. A write past the origin's quota arrives this way, and has
// to reach the caller as an error.
func Call(v js.Value, method string, args ...any) (res js.Value, err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}

		e, ok := r.(js.Error)
		if !ok {
			panic(r)
		}

		res, err = js.Undefined(), fmt.Errorf("%s: %w", method, Err(e.Value))
	}()

	return v.Call(method, args...), nil
}

// Await blocks until p settles. It must not be called from a [js.Func]
// callback.
func Await(p js.Value) (js.Value, error) {
	type settled struct {
		value js.Value
		err   error
	}

	ch := make(chan settled, 1)

	var onValue, onReason js.Func

	release := func() {
		onValue.Release()
		onReason.Release()
	}

	onValue = js.FuncOf(func(_ js.Value, args []js.Value) any {
		ch <- settled{value: first(args)}
		release()
		return nil
	})

	onReason = js.FuncOf(func(_ js.Value, args []js.Value) any {
		ch <- settled{err: Err(first(args))}
		release()
		return nil
	})

	if _, err := Call(p, "then", onValue, onReason); err != nil {
		release()
		return js.Undefined(), err
	}

	s := <-ch
	return s.value, s.err
}

// AwaitCall calls a method that answers with a promise and waits for it.
func AwaitCall(v js.Value, method string, args ...any) (js.Value, error) {
	p, err := Call(v, method, args...)
	if err != nil {
		return js.Undefined(), err
	}

	return Await(p)
}

// Detach lets p run to completion without waiting for it, logging a
// rejection.
func Detach(p js.Value, what string) {
	var onValue, onReason js.Func

	release := func() {
		onValue.Release()
		onReason.Release()
	}

	onValue = js.FuncOf(func(_ js.Value, _ []js.Value) any {
		release()
		return nil
	})

	onReason = js.FuncOf(func(_ js.Value, args []js.Value) any {
		slog.Error(what, "error", Err(first(args)))
		release()
		return nil
	})

	if _, err := Call(p, "then", onValue, onReason); err != nil {
		release()
		slog.Error(what, "error", err)
	}
}

func first(args []js.Value) js.Value {
	if len(args) == 0 {
		return js.Undefined()
	}

	return args[0]
}

// Err turns a JavaScript error value into a Go error, keeping the name the
// caller needs to tell one failure from another -- QuotaExceededError for a
// full origin, NoModificationAllowedError for a file somebody else holds.
func Err(v js.Value) error {
	if v.Type() != js.TypeObject {
		return tgutil.Errorf("%s", str(v))
	}

	name, message := str(v.Get("name")), str(v.Get("message"))

	switch {
	case name != "" && message != "":
		return tgutil.Errorf("%s: %s", name, message)
	case name != "":
		return tgutil.Errorf("%s", name)
	case message != "":
		return tgutil.Errorf("%s", message)
	default:
		return tgutil.NewError("rejected with no reason")
	}
}

func str(v js.Value) string {
	if v.Type() != js.TypeString {
		return ""
	}

	return v.String()
}

// Origin opens the origin's root directory, or says why there is none. It
// blocks, so it is for a goroutine of the caller's own.
func Origin() (js.Value, error) {
	// A sync access handle is the only way to read and write without awaiting
	// anything, and a dedicated Web Worker is the only place it exists. The Go
	// program has to run in one; see toolgui/tgwasm/README.md.
	if handle := js.Global().Get("FileSystemFileHandle"); !handle.Truthy() ||
		handle.Get("prototype").Get("createSyncAccessHandle").Type() != js.TypeFunction {
		return js.Undefined(), tgutil.NewError(
			"no synchronous file access here: the Go program has to run in a" +
				" dedicated Web Worker")
	}

	// The origin private file system belongs to a secure context, so a build
	// served over plain http from anything but localhost has none. That is a
	// hosting requirement rather than something to degrade around: falling
	// back to the heap would quietly lose what the caller meant to keep.
	if secure := js.Global().Get("isSecureContext"); secure.Type() == js.TypeBoolean &&
		!secure.Bool() {
		return js.Undefined(), tgutil.NewError(
			"no origin private file system: this is not a secure context, so" +
				" the site has to be served over https, or from localhost")
	}

	storage := js.Global().Get("navigator").Get("storage")
	if !storage.Truthy() || storage.Get("getDirectory").Type() != js.TypeFunction {
		return js.Undefined(), tgutil.NewError("no origin private file system")
	}

	origin, err := AwaitCall(storage, "getDirectory")
	if err != nil {
		return js.Undefined(), tgutil.Errorf("%w", err)
	}

	return origin, nil
}
