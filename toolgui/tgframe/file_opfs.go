//go:build js && wasm

package tgframe

import (
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall/js"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// The browser's origin private file system is where a wasm build keeps its
// uploads. It is what the tab has instead of a filesystem, and using it is
// what keeps an upload out of the heap: a file the user picked is written
// straight through and read back a chunk at a time, so it never has to fit
// beside the program that reads it.
//
// Only reads and writes on an open handle are synchronous. Getting to a
// directory or a file handle is a promise, and a promise cannot be awaited
// where the [fileBodies] and [fileBody] methods run: they are called under
// the store's lock, and an upload arrives on the JavaScript callback stack,
// which holds the event loop the promise needs to settle. Everything async
// therefore happens on goroutines of this package's own, ahead of demand.
const (
	// opfsRootName is the one directory this package owns. The origin's root
	// is shared with whatever else the page stores there, so nothing outside
	// this is ever listed or removed.
	opfsRootName = "toolgui-state"

	// opfsLockName marks a state's directory as in use. Its handle is held
	// open for as long as the state lives, and the browser gives a file's
	// sync access handle to one holder at a time, which is how the sweep at
	// startup tells a directory a crashed tab left behind from one another
	// tab still has open.
	opfsLockName = ".lock"

	// opfsStatePrefix and the millisecond stamp after it name a state's
	// directory. The stamp is what lets the sweep leave a directory alone that
	// was made moments ago and may still be setting itself up.
	opfsStatePrefix = "state-"

	// opfsPoolSize is how many files are kept created and open ahead of
	// demand, so an upload has one to go into without waiting.
	opfsPoolSize = 8

	// opfsChunkSize is how much a read or a write moves per call. Bytes cross
	// through a typed array either way, so the buffer is this size and made
	// once rather than per call.
	opfsChunkSize = 64 << 10

	// opfsSweepGrace is how long a state directory is taken to be one still
	// being set up, and left alone whatever is or is not in it.
	opfsSweepGrace = time.Minute
)

// opfsCreate and opfsRecursive are the option objects the directory calls
// take. They are never written to, so one of each is enough.
var (
	opfsCreate    = map[string]any{"create": true}
	opfsRecursive = map[string]any{"recursive": true}

	opfsUint8Array = js.Global().Get("Uint8Array")
)

// opfsAt is the {at: offset} a read or a write takes. The offset crosses as a
// float64 because [js.ValueOf] takes no int64, and no browser will hold a
// file anywhere near where that loses a byte.
func opfsAt(off int64) map[string]any {
	return map[string]any{"at": float64(off)}
}

// opfsCall calls a method and returns what JavaScript threw as an error
// instead of panicking with it. A write past the origin's quota arrives this
// way, and an upload the browser has no room for is a thing that happens
// rather than a bug: it has to reach the caller as an error, so the run
// reports it like any other failure.
func opfsCall(v js.Value, method string, args ...any) (res js.Value, err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}

		e, ok := r.(js.Error)
		if !ok {
			panic(r)
		}

		res, err = js.Undefined(), fmt.Errorf("%s: %w", method, opfsErr(e.Value))
	}()

	return v.Call(method, args...), nil
}

// opfsAwait blocks until p settles. It must not be called from a [js.Func]
// callback: the event loop is stopped for the length of one, so the promise
// would never settle and the wait would never end.
func opfsAwait(p js.Value) (js.Value, error) {
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
		ch <- settled{value: opfsFirst(args)}
		release()
		return nil
	})

	onReason = js.FuncOf(func(_ js.Value, args []js.Value) any {
		ch <- settled{err: opfsErr(opfsFirst(args))}
		release()
		return nil
	})

	if _, err := opfsCall(p, "then", onValue, onReason); err != nil {
		release()
		return js.Undefined(), err
	}

	s := <-ch
	return s.value, s.err
}

// opfsAwaitCall calls a method that answers with a promise and waits for it.
func opfsAwaitCall(v js.Value, method string, args ...any) (js.Value, error) {
	p, err := opfsCall(v, method, args...)
	if err != nil {
		return js.Undefined(), err
	}

	return opfsAwait(p)
}

// opfsDetach lets p run to completion without waiting for it, logging a
// rejection. Removing a file goes this way: it happens under the store's
// lock, and on the JavaScript callback stack when an upload replaces one.
func opfsDetach(p js.Value, what string) {
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
		slog.Error(what, "error", opfsErr(opfsFirst(args)))
		release()
		return nil
	})

	if _, err := opfsCall(p, "then", onValue, onReason); err != nil {
		release()
		slog.Error(what, "error", err)
	}
}

func opfsFirst(args []js.Value) js.Value {
	if len(args) == 0 {
		return js.Undefined()
	}

	return args[0]
}

// opfsErr turns a JavaScript error value into a Go error, keeping the name the
// caller needs to tell one failure from another -- QuotaExceededError for a
// full origin, NoModificationAllowedError for a file somebody else holds.
func opfsErr(v js.Value) error {
	if v.Type() != js.TypeObject {
		return tgutil.Errorf("%s", opfsString(v))
	}

	name, message := opfsString(v.Get("name")), opfsString(v.Get("message"))

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

func opfsString(v js.Value) string {
	if v.Type() != js.TypeString {
		return ""
	}

	return v.String()
}

// opfsRoot is the directory the states' directories live in, opened once. The
// chain of promises that gets there runs on a goroutine, and everything that
// needs the handle waits for it.
type opfsRoot struct {
	once   sync.Once
	ready  chan struct{}
	handle js.Value
	err    error
}

var opfsStateRoot = &opfsRoot{ready: make(chan struct{})}

// get returns the handle, opening it on first use. It blocks, so it is for
// this package's own goroutines and not for a [js.Func] callback.
func (r *opfsRoot) get() (js.Value, error) {
	r.once.Do(func() { go r.open() })

	<-r.ready
	return r.handle, r.err
}

func (r *opfsRoot) open() {
	r.handle, r.err = opfsOpenRoot()

	close(r.ready)

	if r.err != nil {
		return
	}

	opfsSweep(r.handle)
}

func opfsOpenRoot() (js.Value, error) {
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
	// back to the heap would quietly put every upload back where this build
	// stopped keeping them, and the page would have no way to know.
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

	origin, err := opfsAwaitCall(storage, "getDirectory")
	if err != nil {
		return js.Undefined(), tgutil.Errorf("%w", err)
	}

	root, err := opfsAwaitCall(origin, "getDirectoryHandle", opfsRootName, opfsCreate)
	if err != nil {
		return js.Undefined(), tgutil.Errorf("%w", err)
	}

	return root, nil
}

// opfsStateName names a state's directory: the prefix, when it was made, and
// enough randomness that two tabs starting in the same millisecond do not
// collide.
func opfsStateName() string {
	return opfsStatePrefix + strconv.FormatInt(time.Now().UnixMilli(), 10) +
		"-" + rand.Text()
}

// opfsStateAge reports how long ago a state directory of ours was made. The
// second result is false for a name this package did not write, which is left
// alone: the root is only ours by convention.
func opfsStateAge(name string) (time.Duration, bool) {
	rest, ok := strings.CutPrefix(name, opfsStatePrefix)
	if !ok {
		return 0, false
	}

	stamp, _, ok := strings.Cut(rest, "-")
	if !ok {
		return 0, false
	}

	ms, err := strconv.ParseInt(stamp, 10, 64)
	if err != nil {
		return 0, false
	}

	return time.Since(time.UnixMilli(ms)), true
}

// opfsSweep removes the state directories left behind by tabs that went away
// without destroying their state. Storage is per origin, so another tab of the
// same app may be using some of what is there -- the point of the lock file is
// that this can tell which, and remove the abandoned ones rather than the lot.
func opfsSweep(root js.Value) {
	names, err := opfsNames(root)
	if err != nil {
		slog.Error("list the state directories", "error", err)
		return
	}

	for _, name := range names {
		if !opfsOrphaned(root, name) {
			continue
		}

		if _, err := opfsAwaitCall(root, "removeEntry", name, opfsRecursive); err != nil {
			slog.Warn("remove an orphaned state directory",
				"dir", name, "error", err)
			continue
		}

		slog.Info("removed an orphaned state directory", "dir", name)
	}
}

// opfsOrphaned reports whether a state directory belongs to nobody. It answers
// false for anything it cannot be sure of: a false negative leaves a directory
// to the next startup, a false positive takes a live tab's uploads away.
func opfsOrphaned(root js.Value, name string) bool {
	age, ours := opfsStateAge(name)
	if !ours {
		return false
	}

	// Age decides whether a directory can be judged at all, before anything in
	// it is looked at. Setting one up is a chain of promises, and for the turns
	// of the event loop between its lock file being made and that file's handle
	// being taken it is indistinguishable from one nobody owns -- so a young
	// directory is left alone whatever state it is in. A sweep runs while this
	// program's own first state is still setting itself up, and while another
	// tab's may be.
	if age <= opfsSweepGrace {
		return false
	}

	dir, err := opfsAwaitCall(root, "getDirectoryHandle", name)
	if err != nil {
		// Not a directory, or gone between the listing and here.
		return false
	}

	lockFile, err := opfsAwaitCall(dir, "getFileHandle", opfsLockName)
	if err != nil {
		// Old enough to judge, and it never got as far as a lock file.
		return true
	}

	lock, err := opfsAwaitCall(lockFile, "createSyncAccessHandle")
	if err != nil {
		// Refused, so somebody holds it: a live state, here or in another tab.
		return false
	}

	if _, err := opfsCall(lock, "close"); err != nil {
		slog.Warn("close a swept lock", "dir", name, "error", err)
	}

	return true
}

// opfsNames lists what a directory holds. The iterator is async, so this is
// one promise per entry.
func opfsNames(dir js.Value) ([]string, error) {
	it, err := opfsCall(dir, "keys")
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	var names []string

	for {
		res, err := opfsAwaitCall(it, "next")
		if err != nil {
			return nil, tgutil.Errorf("%w", err)
		}

		if res.Get("done").Bool() {
			return names, nil
		}

		names = append(names, res.Get("value").String())
	}
}

// opfsBodies keeps one state's uploads in a directory of its own. The
// directory, its lock and the files in it are all made on run's goroutine,
// because none of that can be awaited where newBody is called from.
type opfsBodies struct {
	// ready is closed once the directory and its lock are in place, or the
	// attempt to make them has failed. dir, lock and err are written before it
	// closes and only read after; dir is set as soon as the directory exists,
	// so that a setup which fails after that still has it to remove, and lock
	// stays undefined when setup stopped before taking one.
	ready chan struct{}
	name  string
	dir   js.Value
	lock  js.Value
	err   error

	// pool holds files already created and open, waiting to be handed out.
	pool chan *opfsBody

	// done is closed by destroy to stop run, and stopped by run on its way
	// out, so the directory is only removed once nothing is still opening
	// handles inside it.
	done    chan struct{}
	stopped chan struct{}

	// filled says the pool has been full at least once, which is to say run
	// has caught up with demand and the state is no longer starting up. Until
	// then an empty pool is one still being filled, and a caller may wait: it
	// is a page function on its own goroutine, or a test. After it, an empty
	// pool is one demand drained, and the caller may be an upload on the
	// JavaScript callback stack -- where waiting for the refill would stop the
	// event loop the refill needs to make progress.
	filled atomic.Bool

	mu        sync.Mutex
	live      map[*opfsBody]struct{}
	destroyed bool
	seq       int
}

func newFileBodies() fileBodies {
	b := &opfsBodies{
		ready:   make(chan struct{}),
		pool:    make(chan *opfsBody, opfsPoolSize),
		done:    make(chan struct{}),
		stopped: make(chan struct{}),
		live:    map[*opfsBody]struct{}{},
	}

	go b.run()

	return b
}

// run opens the state's directory and then keeps the pool filled. It is the
// only goroutine that awaits anything on this state's behalf, which is what
// leaves the rest of the type synchronous.
func (b *opfsBodies) run() {
	defer close(b.stopped)

	b.err = b.setup()

	close(b.ready)

	if b.err != nil {
		slog.Error("open a state's file directory", "error", b.err)
		return
	}

	for {
		body, err := b.create()
		if err != nil {
			if !b.stopping() {
				// Out of quota, or the directory went away. What is already
				// pooled still works, and newBody reports an empty pool.
				slog.Error("prepare a file", "dir", b.name, "error", err)
			}

			return
		}

		if !b.track(body) {
			body.close()
			return
		}

		select {
		case b.pool <- body:
		default:
			// Nowhere to put it, so the pool is full and the state has all the
			// files ahead of demand it is going to get. An empty pool from
			// here on is one that was drained rather than one still filling.
			b.filled.Store(true)

			select {
			case b.pool <- body:
			case <-b.done:
				// destroy closed every handle it was tracking, this one
				// included.
				return
			}
		}
	}
}

// setup makes the state's directory and claims it.
func (b *opfsBodies) setup() error {
	root, err := opfsStateRoot.get()
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	b.name = opfsStateName()

	dir, err := opfsAwaitCall(root, "getDirectoryHandle", b.name, opfsCreate)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	// Kept before the rest of setup can fail, so that destroy has something to
	// remove either way. Nothing reads it while err is set.
	b.dir = dir

	lockFile, err := opfsAwaitCall(dir, "getFileHandle", opfsLockName, opfsCreate)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	lock, err := opfsAwaitCall(lockFile, "createSyncAccessHandle")
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	b.lock = lock

	return nil
}

// create makes one more file in the state's directory and opens it. The handle
// stays open for the file's lifetime: holding it is what lets a body read and
// write without awaiting anything.
func (b *opfsBodies) create() (*opfsBody, error) {
	b.mu.Lock()
	b.seq++
	// Named after the counter: naming it after the upload would mean trusting
	// a name the browser chose.
	name := strconv.Itoa(b.seq)
	b.mu.Unlock()

	file, err := opfsAwaitCall(b.dir, "getFileHandle", name, opfsCreate)
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	handle, err := opfsAwaitCall(file, "createSyncAccessHandle")
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	return &opfsBody{bodies: b, name: name, handle: handle}, nil
}

func (b *opfsBodies) newBody() (fileBody, error) {
	body, err := b.take()
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	if !b.track(body) {
		// destroy got there first, and has closed this one already.
		body.close()
		return nil, tgutil.NewError("the state's files are gone")
	}

	return body, nil
}

// take hands out a file run has already created and opened. It answers at once
// whenever one is pooled, which is what stands in for the promise it cannot
// await.
//
// A pool that has already handed a file out and is empty again is one demand
// drained, and refilling it needs turns of the event loop. Waiting for that is
// what a caller on the JavaScript callback stack must never do, and this has
// no way to tell where it was called from, so nobody waits: taking more files
// in one go than the pool holds is an error rather than a tab that stops dead.
//
// The wait below is only reachable before the first file is through, in the
// turns of the event loop right after a state is made. An upload cannot land
// there -- getting there means beating the render of the component that
// accepts one -- and a page function runs on a goroutine of its own, where
// waiting costs nothing.
func (b *opfsBodies) take() (*opfsBody, error) {
	select {
	case body := <-b.pool:
		return body, nil
	default:
	}

	if b.filled.Load() {
		return nil, tgutil.Errorf(
			"no file open and ready: more than %d were taken before the"+
				" browser could open another", opfsPoolSize)
	}

	select {
	case body := <-b.pool:
		return body, nil
	case <-b.done:
		return nil, tgutil.NewError("the state's files are gone")
	case <-b.stopped:
		// Nothing is making files any more, so the pool is all there is.
		select {
		case body := <-b.pool:
			return body, nil
		default:
		}

		if b.err != nil {
			return nil, tgutil.Errorf("%w", b.err)
		}

		return nil, tgutil.NewError("no file storage")
	}
}

// track takes a body into the set whose handles destroy has to close. It
// answers false once the state is destroyed, and the caller closes the body
// itself: nothing is left holding a handle inside a directory on its way out.
func (b *opfsBodies) track(body *opfsBody) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.destroyed {
		return false
	}

	b.live[body] = struct{}{}
	return true
}

func (b *opfsBodies) forget(body *opfsBody) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.live, body)
}

func (b *opfsBodies) stopping() bool {
	select {
	case <-b.done:
		return true
	default:
		return false
	}
}

// destroy closes every handle the state opened and removes its directory.
// Closing comes first, and waiting for run to stop comes after: the browser
// refuses to remove a file anything still holds a sync access handle to.
func (b *opfsBodies) destroy() {
	b.mu.Lock()
	if b.destroyed {
		b.mu.Unlock()
		return
	}

	b.destroyed = true
	live := b.live
	b.live = map[*opfsBody]struct{}{}
	b.mu.Unlock()

	close(b.done)

	for body := range live {
		body.close()
	}

	// Draining unblocks run, which may be waiting to hand a file over.
	b.drainPool()

	go b.removeDir()
}

func (b *opfsBodies) drainPool() {
	for {
		select {
		case <-b.pool:
		default:
			return
		}
	}
}

// removeDir drops the state's directory once run is out of it. A setup that
// made the directory and then failed leaves one too, and it goes the same way:
// the sweep runs once at startup and not again, so a tab that hit this on every
// page switch would pile them up for the rest of its life.
func (b *opfsBodies) removeDir() {
	<-b.stopped

	if !b.dir.Truthy() {
		// setup never got as far as making one.
		return
	}

	// There may be no lock, if that is where setup stopped.
	if b.lock.Truthy() {
		if _, err := opfsCall(b.lock, "close"); err != nil {
			slog.Error("release a state directory lock",
				"dir", b.name, "error", err)
		}
	}

	root, err := opfsStateRoot.get()
	if err != nil {
		return
	}

	if _, err := opfsAwaitCall(root, "removeEntry", b.name, opfsRecursive); err != nil {
		slog.Error("remove a state's file directory", "dir", b.name, "error", err)
	}
}

// opfsBody is one file and the handle it is read and written through. The
// handle is opened once and kept, because reads and writes on it are the only
// part of this filesystem that does not go through a promise.
//
// Every reader shares that one handle: the browser hands a file's sync access
// handle out exclusively, so a second one is refused. The offset therefore
// lives in the reader and the lock is taken for the length of a single read.
type opfsBody struct {
	bodies *opfsBodies
	name   string

	lock   sync.Mutex
	handle js.Value
	buf    js.Value

	// readers is how many readers are open on the file, and removed says the
	// store has dropped it. The file outlives the drop while a reader is on
	// it, the way an unlinked file does for a descriptor already held: the
	// disk build gets that from the kernel, and this build has to count.
	readers int
	removed bool
}

func (b *opfsBody) open() (FileReader, error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if !b.handle.Truthy() {
		return nil, tgutil.NewError("the file is closed")
	}

	size, err := opfsCall(b.handle, "getSize")
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	b.readers++

	// The reader is capped at what is there now. An append can only add past
	// that, so what it reads doesn't change under it.
	return &opfsReader{body: b, size: int64(size.Float())}, nil
}

func (b *opfsBody) write(r io.Reader, atEnd bool) (int64, error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if !b.handle.Truthy() {
		return 0, tgutil.NewError("the file is closed")
	}

	at, err := b.writeStart(atEnd)
	if err != nil {
		return 0, tgutil.Errorf("%w", err)
	}

	buf := make([]byte, opfsChunkSize)

	var total int64

	for {
		n, readErr := r.Read(buf)
		if n > 0 {
			if err := b.writeChunk(buf[:n], at+total); err != nil {
				return total, tgutil.Errorf("%w", err)
			}

			total += int64(n)
		}

		if readErr == io.EOF {
			return total, nil
		}

		if readErr != nil {
			return total, tgutil.Errorf("%w", readErr)
		}
	}
}

// writeStart returns the offset a write begins at, emptying the file first
// when it is replacing what is there rather than appending to it.
func (b *opfsBody) writeStart(atEnd bool) (int64, error) {
	if !atEnd {
		if _, err := opfsCall(b.handle, "truncate", 0); err != nil {
			return 0, tgutil.Errorf("%w", err)
		}

		return 0, nil
	}

	size, err := opfsCall(b.handle, "getSize")
	if err != nil {
		return 0, tgutil.Errorf("%w", err)
	}

	return int64(size.Float()), nil
}

// writeChunk copies bs through the body's typed array and writes it at off.
func (b *opfsBody) writeChunk(bs []byte, off int64) error {
	if !b.buf.Truthy() {
		b.buf = opfsUint8Array.New(opfsChunkSize)
	}

	js.CopyBytesToJS(b.buf, bs)

	view := b.buf
	if len(bs) < opfsChunkSize {
		v, err := opfsCall(b.buf, "subarray", 0, len(bs))
		if err != nil {
			return tgutil.Errorf("%w", err)
		}

		view = v
	}

	// An upload the origin has no room for throws QuotaExceededError here.
	// opfsCall hands it back as an error rather than a panic, and it travels
	// out through [State.WriteFile] like any other failure of the run.
	wrote, err := opfsCall(b.handle, "write", view, opfsAt(off))
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	if n := int(wrote.Float()); n != len(bs) {
		return tgutil.Errorf("wrote %d of %d bytes", n, len(bs))
	}

	return nil
}

// remove drops the file. A reader already open goes on reading it: the handle
// is shared, so closing it here would break one mid-read, and a page that held
// a reader across the upload that replaced its file would get an error where
// the other builds hand it the bytes it opened. The file goes when the last
// reader closes instead.
//
// [fileBodies.destroy] does not wait like this. A state that is gone takes its
// readers with it, and the directory cannot be removed while a handle inside
// it is open.
func (b *opfsBody) remove() {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.removed = true

	if b.readers > 0 {
		return
	}

	b.drop()
}

// release gives back a reader, and drops the file if it was the last one on a
// file the store has already removed.
func (b *opfsBody) release() {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.readers--

	if b.readers > 0 || !b.removed {
		return
	}

	b.drop()
}

// drop closes the handle and removes the file. It must be called with the lock
// held. The handle goes first: the browser refuses to remove a file something
// still holds one for.
func (b *opfsBody) drop() {
	if !b.handle.Truthy() {
		// Already shut, which means destroy took the whole directory with it.
		return
	}

	b.shut()
	b.bodies.forget(b)

	p, err := opfsCall(b.bodies.dir, "removeEntry", b.name)
	if err != nil {
		slog.Error("remove a file", "name", b.name, "error", err)
		return
	}

	// Nothing waits for this: remove runs under the store's lock, and on the
	// JavaScript callback stack when an upload replaces a file.
	opfsDetach(p, "remove a file")
}

// close shuts the handle and leaves the file where it is. It is what destroy
// uses: every handle has to be closed before the directory can go.
func (b *opfsBody) close() {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.shut()
}

// shut closes the handle. It must be called with the lock held.
func (b *opfsBody) shut() {
	if !b.handle.Truthy() {
		return
	}

	if _, err := opfsCall(b.handle, "close"); err != nil {
		slog.Error("close a file handle", "name", b.name, "error", err)
	}

	b.handle = js.Undefined()
	b.buf = js.Undefined()
}

// opfsReader reads one body. The handle underneath is shared, so the position
// is here and each read takes the body's lock for as long as it needs it.
type opfsReader struct {
	body *opfsBody
	size int64

	off    int64
	buf    js.Value
	closed bool
}

func (r *opfsReader) Read(p []byte) (int, error) {
	n, err := r.readAt(p, r.off)
	r.off += int64(n)

	if n > 0 && err == io.EOF {
		// Read may come up short without that being the end.
		return n, nil
	}

	return n, err
}

func (r *opfsReader) ReadAt(p []byte, off int64) (int, error) {
	return r.readAt(p, off)
}

// readAt fills p from off, a chunk at a time, through the reader's own typed
// array. Nothing is read that the file did not already hold when the reader
// opened, so an append cannot move the end under it.
func (r *opfsReader) readAt(p []byte, off int64) (int, error) {
	if r.closed {
		return 0, tgutil.NewError("the reader is closed")
	}

	if off < 0 {
		return 0, tgutil.NewError("negative offset")
	}

	want := len(p)
	if want == 0 {
		return 0, nil
	}

	if off >= r.size {
		return 0, io.EOF
	}

	if rest := r.size - off; int64(want) > rest {
		p = p[:rest]
	}

	view, span := r.view()

	read := 0
	for read < len(p) {
		n, err := r.readChunk(p[read:], off+int64(read), view, span)
		if err != nil {
			return read, tgutil.Errorf("%w", err)
		}

		if n == 0 {
			break
		}

		read += n
	}

	if read < want {
		return read, io.EOF
	}

	return read, nil
}

// readChunk reads at most one chunk into the front of p.
func (r *opfsReader) readChunk(p []byte, off int64, view js.Value, span int) (int, error) {
	if len(p) < span {
		v, err := opfsCall(view, "subarray", 0, len(p))
		if err != nil {
			return 0, tgutil.Errorf("%w", err)
		}

		view = v
	}

	r.body.lock.Lock()
	defer r.body.lock.Unlock()

	if !r.body.handle.Truthy() {
		return 0, tgutil.NewError("the file is closed")
	}

	// read takes the offset, so the whole file is never pulled in to get at a
	// piece of it -- which is what archive/zip and the image decoders want.
	got, err := opfsCall(r.body.handle, "read", view, opfsAt(off))
	if err != nil {
		return 0, tgutil.Errorf("%w", err)
	}

	n := got.Int()
	if n > 0 {
		js.CopyBytesToGo(p[:n], view)
	}

	return n, nil
}

// view returns the reader's chunk buffer and its length, making it on first
// use. It is made once: a typed array per read would cost more than the copy
// through it.
func (r *opfsReader) view() (js.Value, int) {
	if !r.buf.Truthy() {
		span := r.size
		if span > opfsChunkSize {
			span = opfsChunkSize
		}

		if span < 1 {
			span = 1
		}

		r.buf = opfsUint8Array.New(int(span))
	}

	return r.buf, r.buf.Length()
}

func (r *opfsReader) Seek(offset int64, whence int) (int64, error) {
	if r.closed {
		return 0, tgutil.NewError("the reader is closed")
	}

	var at int64

	switch whence {
	case io.SeekStart:
		at = offset
	case io.SeekCurrent:
		at = r.off + offset
	case io.SeekEnd:
		at = r.size + offset
	default:
		return 0, tgutil.Errorf("unknown whence %d", whence)
	}

	if at < 0 {
		return 0, tgutil.Errorf("seek to %d, before the start of the file", at)
	}

	r.off = at
	return at, nil
}

// Close releases the reader. The handle stays open unless this was the last
// reader on a file the store has already dropped: it is the body's, and the
// other readers are still on it.
func (r *opfsReader) Close() error {
	if r.closed {
		return nil
	}

	r.closed = true
	r.buf = js.Undefined()

	r.body.release()

	return nil
}
