//go:build js && wasm

package tgframe

import (
	"crypto/rand"
	"io"
	"strconv"
	"strings"
	"syscall/js"
	"testing"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// These need a browser: the origin private file system exists nowhere else, and
// the handles the store reads and writes through are a dedicated Web Worker's
// only. `task test_wasm` is what runs them there.

// opfsState makes a state and waits until its files can be written. A page
// never waits like this -- newBody is called where nothing may block, and the
// pool is filled ahead of demand for exactly that reason -- but a test that
// stores a file in the same breath as making the state has to.
func opfsState(t *testing.T) (*State, *opfsBodies) {
	t.Helper()

	s := NewState()

	bodies, ok := s.files.bodies.(*opfsBodies)
	if !ok {
		t.Fatalf("bodies are %T, want the origin private file system", s.files.bodies)
	}

	<-bodies.ready

	if bodies.err != nil {
		t.Fatalf("open the state's directory: %v", bodies.err)
	}

	return s, bodies
}

// opfsPooled waits until the state is done starting up: the pool has filled, so
// every file it keeps open ahead of demand is there and taking one no longer
// waits for anything. A page gets here within a few turns of the event loop,
// long before a user can pick a file.
func opfsPooled(t *testing.T, bodies *opfsBodies) {
	t.Helper()

	for range 400 {
		if bodies.filled.Load() {
			return
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Fatal("the file pool never filled")
}

// opfsEntry answers whether a directory holds an entry under name.
func opfsEntry(t *testing.T, dir js.Value, name string, isDir bool) bool {
	t.Helper()

	method := "getFileHandle"
	if isDir {
		method = "getDirectoryHandle"
	}

	_, err := opfsAwaitCall(dir, method, name)
	return err == nil
}

// opfsWaitGone waits for an entry to go. Removal is started and not waited
// for, because the callers are where nothing may block, so a test has to give
// the promise its turns of the event loop.
func opfsWaitGone(t *testing.T, dir js.Value, name string, isDir bool) {
	t.Helper()

	for range 200 {
		if !opfsEntry(t, dir, name, isDir) {
			return
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Errorf("expect %q to be gone", name)
}

// TestStateFilesInOPFS checks the browser keeps an upload in the origin private
// file system rather than the tab's heap, and that a file the key no longer
// holds is removed instead of being left there for the origin's quota to trip
// over.
func TestStateFilesInOPFS(t *testing.T) {
	s, bodies := opfsState(t)
	defer s.Destroy()

	file, err := s.WriteFile("comp", "a.txt", strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	body, ok := file.body.(*opfsBody)
	if !ok {
		t.Fatalf("body is %T, want a file in the origin private file system", file.body)
	}

	if !opfsEntry(t, bodies.dir, body.name, false) {
		t.Fatalf("expect %q in the state's directory", body.name)
	}

	bs, err := file.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	if string(bs) != "hello" {
		t.Errorf("the stored file holds %q, want hello", bs)
	}

	if _, err := s.WriteFile("comp", "b.txt", strings.NewReader("new")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	opfsWaitGone(t, bodies.dir, body.name, false)
}

// TestStateFileReadAt checks a reader reads at an offset without pulling the
// whole file in, which is what archive/zip and the image decoders ask of it.
func TestStateFileReadAt(t *testing.T) {
	s, _ := opfsState(t)
	defer s.Destroy()

	// Longer than one chunk, so a read is more than one call into the browser
	// and the offsets have to add up.
	want := []byte(strings.Repeat("0123456789abcdef", opfsChunkSize/8))

	file, err := s.WriteFile("comp", "a.bin", strings.NewReader(string(want)))
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if file.Size() != int64(len(want)) {
		t.Errorf("Size = %d, want %d", file.Size(), len(want))
	}

	fp, err := file.Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer fp.Close()

	at := make([]byte, 100)
	if _, err := fp.ReadAt(at, int64(len(want))-100); err != nil {
		t.Fatalf("ReadAt: %v", err)
	}

	if string(at) != string(want[len(want)-100:]) {
		t.Errorf("ReadAt at the end = %q, want %q", at, want[len(want)-100:])
	}

	// A read that runs off the end comes up short and says so.
	over := make([]byte, 64)
	if n, err := fp.ReadAt(over, int64(len(want))-10); n != 10 || err != io.EOF {
		t.Errorf("ReadAt over the end = (%d, %v), want (10, EOF)", n, err)
	}

	got, err := io.ReadAll(fp)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	if string(got) != string(want) {
		t.Errorf("ReadAll gave %d bytes, want %d", len(got), len(want))
	}
}

// TestStateFileReadersAreIndependent checks two readers over one file keep
// their own position. They share the body's handle -- the browser hands a
// file's out exclusively -- so the position cannot live on it.
func TestStateFileReadersAreIndependent(t *testing.T) {
	s, _ := opfsState(t)
	defer s.Destroy()

	file, err := s.WriteFile("comp", "a.txt", strings.NewReader("hello file"))
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	first, err := file.Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer first.Close()

	second, err := file.Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer second.Close()

	head := make([]byte, 5)
	if _, err := io.ReadFull(first, head); err != nil {
		t.Fatalf("ReadFull: %v", err)
	}

	if string(head) != "hello" {
		t.Errorf("the first reader gave %q, want hello", head)
	}

	// The second one has not moved.
	rest, err := io.ReadAll(second)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	if string(rest) != "hello file" {
		t.Errorf("the second reader gave %q, want hello file", rest)
	}

	if _, err := first.Seek(6, io.SeekStart); err != nil {
		t.Fatalf("Seek: %v", err)
	}

	tail, err := io.ReadAll(first)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	if string(tail) != "file" {
		t.Errorf("the first reader gave %q after a seek, want file", tail)
	}
}

// TestStateFileAppend checks a file built a chunk at a time ends up whole,
// which is what a transport that can only carry a piece per call needs.
func TestStateFileAppend(t *testing.T) {
	s, _ := opfsState(t)
	defer s.Destroy()

	file, err := s.NewFile("a.txt")
	if err != nil {
		t.Fatalf("NewFile: %v", err)
	}

	for _, part := range []string{"hello", " ", "file"} {
		if err := file.Append(strings.NewReader(part)); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	if file.Size() != 10 {
		t.Errorf("Size = %d, want 10", file.Size())
	}

	bs, err := file.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	if string(bs) != "hello file" {
		t.Errorf("Bytes = %q, want hello file", bs)
	}
}

// TestStateSetFileFromACallback checks a file can be stored from inside a
// JavaScript callback, which is where an upload arrives. Nothing on that path
// may wait on a promise: a call into Go holds the event loop for its whole
// length, so a promise awaited there would never settle and the tab would stop
// dead. If this ever regresses it hangs rather than fails, and the test times
// out.
func TestStateSetFileFromACallback(t *testing.T) {
	s, bodies := opfsState(t)
	defer s.Destroy()

	opfsPooled(t, bodies)

	var stored error

	done := make(chan struct{})

	fn := js.FuncOf(func(_ js.Value, _ []js.Value) any {
		_, stored = s.SetFile("comp", "a.txt", []byte("hello"))
		close(done)
		return nil
	})
	defer fn.Release()

	// Through a timer, so the browser calls it with Go parked -- the same way
	// the bridge's uploadFile is reached, rather than Go calling itself.
	js.Global().Call("setTimeout", fn, 0)
	<-done

	if stored != nil {
		t.Fatalf("SetFile from a callback: %v", stored)
	}

	bs, err := s.GetFile("comp").Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	if string(bs) != "hello" {
		t.Errorf("Bytes = %q, want hello", bs)
	}
}

// TestStateNewFileRunsOutRatherThanWaiting checks that asking for more files in
// one JavaScript callback than the pool holds is an error. Waiting there for
// the pool to refill would stop the event loop the refill needs, and the tab
// with it. If that ever regresses this hangs rather than fails.
func TestStateNewFileRunsOutRatherThanWaiting(t *testing.T) {
	s, bodies := opfsState(t)
	defer s.Destroy()

	opfsPooled(t, bodies)

	var failed error

	done := make(chan struct{})

	fn := js.FuncOf(func(_ js.Value, _ []js.Value) any {
		defer close(done)

		// Twice the pool, so it runs out however full it was to start with.
		for range opfsPoolSize * 2 {
			if _, err := s.NewFile("a.txt"); err != nil {
				failed = err
				return nil
			}
		}

		return nil
	})
	defer fn.Release()

	js.Global().Call("setTimeout", fn, 0)
	<-done

	if failed == nil {
		t.Error("expect taking more files than the pool holds to fail")
	}
}

// TestStateFileReaderSurvivesReplacement checks a reader opened before an
// upload replaced its file goes on reading what it opened, the way an unlinked
// file does for a descriptor already held. The file itself goes once the last
// reader closes.
func TestStateFileReaderSurvivesReplacement(t *testing.T) {
	s, bodies := opfsState(t)
	defer s.Destroy()

	file, err := s.WriteFile("comp", "a.txt", strings.NewReader("hello file"))
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	body, ok := file.body.(*opfsBody)
	if !ok {
		t.Fatalf("body is %T, want a file in the origin private file system", file.body)
	}

	fp, err := file.Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer fp.Close()

	head := make([]byte, 5)
	if _, err := io.ReadFull(fp, head); err != nil {
		t.Fatalf("ReadFull: %v", err)
	}

	// The upload that takes the key drops the body this reader is on.
	if _, err := s.WriteFile("comp", "b.txt", strings.NewReader("new")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	rest, err := io.ReadAll(fp)
	if err != nil {
		t.Fatalf("ReadAll after the file was replaced: %v", err)
	}

	if string(rest) != " file" {
		t.Errorf("the reader gave %q after the file was replaced, want \" file\"", rest)
	}

	if err := fp.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	opfsWaitGone(t, bodies.dir, body.name, false)
}

// TestStateDestroyRemovesDirectory checks the whole directory goes with the
// state. A tab that reloads all day must not leave the origin full of the
// files of sessions that are over.
func TestStateDestroyRemovesDirectory(t *testing.T) {
	s, bodies := opfsState(t)

	if _, err := s.WriteFile("comp", "a.txt", strings.NewReader("hello")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	root, err := opfsStateRoot.get()
	if err != nil {
		t.Fatalf("open the state root: %v", err)
	}

	s.Destroy()

	opfsWaitGone(t, root, bodies.name, true)

	// The store is done handing files out, rather than writing them into a
	// directory that is on its way to being removed.
	if _, err := s.WriteFile("comp", "b.txt", strings.NewReader("new")); err == nil {
		t.Error("expect WriteFile to fail on a destroyed state")
	}
}

// TestStateDirectoryGoesAfterAPartialSetup checks a state whose setup made its
// directory and then failed still takes the directory with it. Nothing else
// would: the sweep runs once at startup and not again, so a tab that hit this
// on every page switch would pile them up for the rest of its life.
func TestStateDirectoryGoesAfterAPartialSetup(t *testing.T) {
	root, err := opfsStateRoot.get()
	if err != nil {
		t.Fatalf("open the state root: %v", err)
	}

	name := opfsStateName()

	dir, err := opfsAwaitCall(root, "getDirectoryHandle", name, opfsCreate)
	if err != nil {
		t.Fatalf("make %q: %v", name, err)
	}

	// What run leaves behind when it got the directory and then failed at the
	// lock -- out of quota, say. There is no lock handle to close, and err is
	// set, and the directory still has to go.
	bodies := &opfsBodies{
		ready:   make(chan struct{}),
		pool:    make(chan *opfsBody, opfsPoolSize),
		done:    make(chan struct{}),
		stopped: make(chan struct{}),
		live:    map[*opfsBody]struct{}{},
		name:    name,
		dir:     dir,
		err:     tgutil.NewError("setup failed after the directory was made"),
	}
	close(bodies.ready)
	close(bodies.stopped)

	bodies.destroy()

	opfsWaitGone(t, root, name, true)
}

// TestOPFSSweep checks startup clears what a crashed tab left behind without
// touching what a live one is using. Storage is per origin, so the same app in
// another tab is reading directories in here.
func TestOPFSSweep(t *testing.T) {
	root, err := opfsStateRoot.get()
	if err != nil {
		t.Fatalf("open the state root: %v", err)
	}

	// A state that crashed: its lock file is there, and nobody holds it.
	crashed := opfsFixture(t, root, opfsStatePrefix+"1000-crashed", true)

	// A state that never got as far as its lock file, long enough ago that it
	// cannot still be setting itself up.
	stalled := opfsFixture(t, root, opfsStatePrefix+"1000-stalled", false)

	// The same, from this millisecond: a state being made right now looks like
	// this for a turn or two of the event loop.
	fresh := opfsFixture(t, root,
		opfsStatePrefix+strconv.FormatInt(time.Now().UnixMilli(), 10)+"-fresh", false)

	// And a directory of this millisecond whose lock file is there but not yet
	// held, which is what a state looks like between making the file and taking
	// its handle. Sweeping on the lock alone would take this one: the sweep runs
	// while this program's own first state is still setting itself up, and
	// while another tab's may be.
	opening := opfsFixture(t, root,
		opfsStatePrefix+strconv.FormatInt(time.Now().UnixMilli(), 10)+"-opening", true)

	// Somebody else's directory. The origin's root is shared, and ours is only
	// ours by convention.
	foreign := opfsFixture(t, root, "not-a-toolgui-state-"+rand.Text(), true)

	// A live state, whose lock this worker is holding.
	live, bodies := opfsState(t)
	defer live.Destroy()

	opfsSweep(root)

	for _, name := range []string{crashed, stalled} {
		if opfsEntry(t, root, name, true) {
			t.Errorf("expect the abandoned %q to be swept", name)
		}
	}

	for _, name := range []string{fresh, opening, foreign, bodies.name} {
		if !opfsEntry(t, root, name, true) {
			t.Errorf("expect %q to survive the sweep", name)
		}
	}
}

// opfsFixture makes a state directory the sweep will find, with or without the
// lock file that says a state is using it.
func opfsFixture(t *testing.T, root js.Value, name string, locked bool) string {
	t.Helper()

	dir, err := opfsAwaitCall(root, "getDirectoryHandle", name, opfsCreate)
	if err != nil {
		t.Fatalf("make %q: %v", name, err)
	}

	if locked {
		if _, err := opfsAwaitCall(dir, "getFileHandle", opfsLockName, opfsCreate); err != nil {
			t.Fatalf("make the lock in %q: %v", name, err)
		}
	}

	// Half of these are here to be swept, so a removal that finds nothing is
	// the expected outcome rather than something to report.
	t.Cleanup(func() {
		_, _ = opfsAwaitCall(root, "removeEntry", name, opfsRecursive)
	})

	return name
}
