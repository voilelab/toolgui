//go:build js && wasm

package tgframe

import (
	"crypto/rand"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall/js"
	"time"

	"github.com/voilelab/toolgui/toolgui/internal/opfs"
	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// A wasm build keeps uploads in the browser's origin private file system
// (OPFS), streamed in chunks so they stay out of the heap.
//
// Only reads and writes on an open handle are synchronous. Opening handles is
// a promise, which can't be awaited under the store's lock or on the JS
// callback stack, so all async work runs ahead of demand on our own
// goroutines.
const (
	// opfsRootName is the one directory this package owns in the shared
	// origin root.
	opfsRootName = "toolgui-state"

	// opfsLockName marks a state's directory as in use. Its sync access
	// handle is held for the state's life; since handles are exclusive, the
	// startup sweep can tell orphaned directories from live ones.
	opfsLockName = ".lock"

	// opfsStatePrefix plus a millisecond stamp names a state's directory;
	// the stamp lets the sweep skip ones still being set up.
	opfsStatePrefix = "state-"

	// opfsPoolSize is how many files are kept open ahead of demand.
	opfsPoolSize = 8

	// opfsChunkSize is the bytes moved per read or write call.
	opfsChunkSize = opfs.ChunkSize

	// opfsSweepGrace is how long a new state directory is left alone by the
	// sweep.
	opfsSweepGrace = time.Minute

	// opfsRemoveTries is how many times directory removal is retried. It
	// fails while the page still writes an upload after a page switch, which
	// ends on its own.
	opfsRemoveTries = 24
)

// opfsRemoveGap is the delay between removal tries; a var for tests.
var opfsRemoveGap = 5 * time.Second

// opfsRoot is the parent of the states' directories, opened once on a
// goroutine.
type opfsRoot struct {
	once   sync.Once
	ready  chan struct{}
	handle js.Value
	err    error
}

var opfsStateRoot = &opfsRoot{ready: make(chan struct{})}

// get returns the handle, opening it on first use. It blocks, so never call
// it from a [js.Func] callback.
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
	origin, err := opfs.Origin()
	if err != nil {
		return js.Undefined(), tgutil.Errorf("%w", err)
	}

	root, err := opfs.AwaitCall(origin, "getDirectoryHandle", opfsRootName, opfs.Create)
	if err != nil {
		return js.Undefined(), tgutil.Errorf("%w", err)
	}

	return root, nil
}

// opfsStateName names a new state directory: prefix, timestamp, random
// suffix.
func opfsStateName() string {
	return opfsStatePrefix + strconv.FormatInt(time.Now().UnixMilli(), 10) +
		"-" + rand.Text()
}

// opfsStateAge reports the age of a state directory, false for names we
// didn't write.
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

// opfsSweep removes state directories left by dead tabs, using lock files
// to skip ones other live tabs hold.
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

		if _, err := opfs.AwaitCall(root, "removeEntry", name, opfs.Recursive); err != nil {
			slog.Warn("remove an orphaned state directory",
				"dir", name, "error", err)
			continue
		}

		slog.Info("removed an orphaned state directory", "dir", name)
	}
}

// opfsOrphaned reports whether a state directory belongs to nobody. When
// unsure it says false, leaving it to the next startup.
func opfsOrphaned(root js.Value, name string) bool {
	age, ours := opfsStateAge(name)
	if !ours {
		return false
	}

	// A young directory may be mid-setup (lock file made, handle not yet
	// taken), so skip it.
	if age <= opfsSweepGrace {
		return false
	}

	dir, err := opfs.AwaitCall(root, "getDirectoryHandle", name)
	if err != nil {
		// Not a directory, or gone between the listing and here.
		return false
	}

	lockFile, err := opfs.AwaitCall(dir, "getFileHandle", opfsLockName)
	if err != nil {
		// Old enough to judge, and it never got as far as a lock file.
		return true
	}

	lock, err := opfs.AwaitCall(lockFile, "createSyncAccessHandle")
	if err != nil {
		// Held by a live state.
		return false
	}

	if _, err := opfs.Call(lock, "close"); err != nil {
		slog.Warn("close a swept lock", "dir", name, "error", err)
	}

	return true
}

// opfsNames lists a directory, one promise per entry.
func opfsNames(dir js.Value) ([]string, error) {
	it, err := opfs.Call(dir, "keys")
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	var names []string

	for {
		res, err := opfs.AwaitCall(it, "next")
		if err != nil {
			return nil, tgutil.Errorf("%w", err)
		}

		if res.Get("done").Bool() {
			return names, nil
		}

		names = append(names, res.Get("value").String())
	}
}

// opfsBodies keeps one state's uploads in its own directory. The directory,
// lock and files are made on run's goroutine.
type opfsBodies struct {
	// ready is closed once setup is done or failed; dir, lock and err are
	// read only after. dir is set early so a failed setup can still remove
	// it.
	ready chan struct{}
	name  string
	dir   js.Value
	lock  *opfs.SyncFile
	err   error

	// pool holds files already open, waiting to be handed out.
	pool chan *opfsBody

	// done is closed by destroy to stop run; stopped is closed by run on
	// exit, before the directory is removed.
	done    chan struct{}
	stopped chan struct{}

	// filled says the pool has been full once. Before that, take may wait
	// for startup; after, an empty pool is an error, since waiting on the JS
	// callback stack would deadlock the refill.
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

// run opens the state's directory and keeps the pool filled. It is the only
// goroutine that awaits, keeping the rest synchronous.
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
				// Out of quota or directory gone; pooled files still work.
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
			// Pool is full: startup is over.
			b.filled.Store(true)

			select {
			case b.pool <- body:
			case <-b.done:
				// destroy closed this handle too.
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

	dir, err := opfs.AwaitCall(root, "getDirectoryHandle", b.name, opfs.Create)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	// Set before later failures so destroy can remove it.
	b.dir = dir

	lock, err := opfs.OpenSyncFile(dir, opfsLockName)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	b.lock = lock

	return nil
}

// create makes and opens one more file. The handle stays open for the
// file's life, so reads and writes never await.
func (b *opfsBodies) create() (*opfsBody, error) {
	b.mu.Lock()
	b.seq++
	// Counter name, never the untrusted upload name.
	name := strconv.Itoa(b.seq)
	b.mu.Unlock()

	file, err := opfs.OpenSyncFile(b.dir, name)
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	return &opfsBody{bodies: b, name: name, file: file}, nil
}

func (b *opfsBodies) newBody() (fileBody, error) {
	body, err := b.take()
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	if !b.track(body) {
		// destroy already closed it.
		body.close()
		return nil, tgutil.NewError("the state's files are gone")
	}

	return body, nil
}

// reserve names an empty file for the page to write and tracks it for
// destroy. No handle is opened, since it would block the page's writable
// stream; [opfsBody.adopt] takes one when the page is done. The pool (open
// handles) is not used.
func (b *opfsBodies) reserve() (*opfsBody, error) {
	// Never wait on the JS callback stack. Setup finishes long before a user
	// can pick a file.
	select {
	case <-b.ready:
	default:
		return nil, tgutil.NewError("no file storage yet: the state's" +
			" directory is still being opened")
	}

	if b.err != nil {
		return nil, tgutil.Errorf("%w", b.err)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.destroyed {
		return nil, tgutil.NewError("the state's files are gone")
	}

	b.seq++

	body := &opfsBody{bodies: b, name: strconv.Itoa(b.seq)}
	b.live[body] = struct{}{}

	return body, nil
}

// dirPath is the state's directory path from the OPFS root, for the page to
// reach reserved files.
func (b *opfsBodies) dirPath() []string {
	return []string{opfsRootName, b.name}
}

// take hands out a pooled open file without waiting.
//
// Once the pool has filled, an empty pool is an error rather than a wait,
// since the caller may be on the JS callback stack. Waiting only happens
// during startup, before any upload component can render.
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
		// run has stopped; only the pool is left.
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

// track adds body to the set destroy closes. It returns false once
// destroyed; the caller then closes body itself.
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

// destroy closes every handle and removes the directory. Handles close
// first, since open files block removal.
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

// removeDir drops the state's directory once run has stopped, including one
// a failed setup left.
//
// It retries: a page switch mid-upload leaves the page's writable stream open
// in it, which blocks removal until the write ends.
func (b *opfsBodies) removeDir() {
	<-b.stopped

	if !b.dir.Truthy() {
		// setup never got as far as making one.
		return
	}

	// Release the lock (if any) first, so the next sweep can judge a
	// directory this fails to remove.
	if b.lock != nil {
		if err := b.lock.Close(); err != nil {
			slog.Error("release a state directory lock",
				"dir", b.name, "error", err)
		}
	}

	root, err := opfsStateRoot.get()
	if err != nil {
		return
	}

	var last error

	for try := range opfsRemoveTries + 1 {
		if try > 0 {
			time.Sleep(opfsRemoveGap)
		}

		_, last = opfs.AwaitCall(root, "removeEntry", b.name, opfs.Recursive)
		if last == nil {
			return
		}

		if strings.Contains(last.Error(), "NotFoundError") {
			// Already gone.
			return
		}
	}

	// Left for the next startup's sweep.
	slog.Error("remove a state's file directory", "dir", b.name, "error", last)
}

// opfsBody is one file and its sync access handle, opened once and kept.
//
// Handles are exclusive, so all readers share it: each reader keeps its own
// offset and locks per read.
//
// A reserved body has no handle until [opfsBody.adopt].
type opfsBody struct {
	bodies *opfsBodies
	name   string

	lock sync.Mutex
	file *opfs.SyncFile

	// readers counts open readers; removed says the store dropped the file.
	// Like an unlinked file, it lives until the last reader closes.
	readers int
	removed bool

	// done says the body is dropped or closed. (No handle doesn't imply
	// done: reserved bodies have none yet.)
	done bool
}

// adopt takes over the sync access handle the page opened on a reserved
// file and returns its size. Synchronous, since the page opens the handle.
//
// The page must close its writable stream first, or opening the handle fails
// with NoModificationAllowedError.
func (b *opfsBody) adopt(handle js.Value) (int64, error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.done {
		return 0, tgutil.NewError("the file is gone")
	}

	if b.file != nil {
		return 0, tgutil.NewError("the file is open already")
	}

	if !handle.Truthy() {
		return 0, tgutil.NewError("no file handle")
	}

	b.file = opfs.NewSyncFile(handle)

	size, err := b.file.Size()
	if err != nil {
		return 0, tgutil.Errorf("%w", err)
	}

	return size, nil
}

func (b *opfsBody) open() (FileReader, error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.file == nil {
		return nil, tgutil.Errorf("%w", b.noHandle())
	}

	size, err := b.file.Size()
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	b.readers++

	// Cap the reader at the current size so appends don't affect it.
	return &opfsReader{body: b, size: size}, nil
}

func (b *opfsBody) write(r io.Reader, atEnd bool) (int64, error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.file == nil {
		return 0, tgutil.Errorf("%w", b.noHandle())
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
			if _, err := b.file.WriteAt(buf[:n], at+total); err != nil {
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

// writeStart returns the write offset, truncating first unless appending.
func (b *opfsBody) writeStart(atEnd bool) (int64, error) {
	if !atEnd {
		if err := b.file.Truncate(0); err != nil {
			return 0, tgutil.Errorf("%w", err)
		}

		return 0, nil
	}

	return b.file.Size()
}

// noHandle says why there is no handle: closed, or not handed over yet. It
// must be called with the lock held.
func (b *opfsBody) noHandle() error {
	if b.done {
		return tgutil.NewError("the file is closed")
	}

	return tgutil.NewError("the file has not been handed over yet")
}

// remove drops the file once the last open reader closes, matching the
// other builds.
//
// [fileBodies.destroy] doesn't wait for readers.
func (b *opfsBody) remove() {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.removed = true

	if b.readers > 0 {
		return
	}

	b.drop()
}

// release gives back a reader, dropping a removed file after the last one.
func (b *opfsBody) release() {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.readers--

	if b.readers > 0 || !b.removed {
		return
	}

	b.drop()
}

// drop closes the handle, then removes the file. It must be called with the
// lock held. An unfilled reservation may have no file; NotFoundError is just
// logged.
func (b *opfsBody) drop() {
	if b.done {
		// Already dropped or closed.
		return
	}

	b.done = true
	b.shut()
	b.bodies.forget(b)

	p, err := opfs.Call(b.bodies.dir, "removeEntry", b.name)
	if err != nil {
		slog.Error("remove a file", "name", b.name, "error", err)
		return
	}

	// Don't wait: we may be under the store's lock or on the JS stack.
	opfs.Detach(p, "remove a file")
}

// close shuts the handle and keeps the file. Used by destroy.
func (b *opfsBody) close() {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.done = true
	b.shut()
}

// shut closes the handle. It must be called with the lock held.
func (b *opfsBody) shut() {
	if b.file == nil {
		return
	}

	if err := b.file.Close(); err != nil {
		slog.Error("close a file handle", "name", b.name, "error", err)
	}

	b.file = nil
}

// opfsReader reads one body through its shared handle, keeping its own
// offset.
type opfsReader struct {
	body *opfsBody
	size int64

	off    int64
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

// readAt fills p from off in chunks, up to the size at open.
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

	r.body.lock.Lock()
	defer r.body.lock.Unlock()

	if r.body.file == nil {
		return 0, tgutil.NewError("the file is closed")
	}

	// Positional read: random access without loading the whole file.
	read, err := r.body.file.ReadAt(p, off)
	if err != nil && err != io.EOF {
		return read, tgutil.Errorf("%w", err)
	}

	if read < want {
		return read, io.EOF
	}

	return read, nil
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

// Close releases the reader. The shared handle stays open unless this was
// the last reader of a dropped file.
func (r *opfsReader) Close() error {
	if r.closed {
		return nil
	}

	r.closed = true

	r.body.release()

	return nil
}
