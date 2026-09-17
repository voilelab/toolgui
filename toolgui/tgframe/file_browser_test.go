//go:build js && wasm

package tgframe

import (
	"io"
	"strings"
	"syscall/js"
	"testing"
	"time"
)

// These need a browser, like the rest of the origin private file system's
// tests. What they stand in for is the page: an upload is written by
// JavaScript with a writable stream and read by Go through a sync access
// handle, so a test of that split has to drive both halves.

// browserUploadSize is the file the streaming test puts through. It is far
// past what the tab would hold as a base64 string beside a copy of the bytes,
// which is what this transport used to ask of it, and nothing here holds more
// than a chunk of it at a time.
const browserUploadSize = 200 << 20

// browserChunkSize is how much of the big file is made at once, which is the
// same as the store moves per call: the test holds no more of it than the code
// under test does. The pattern repeats every chunk, so a byte anywhere in the
// file is known from its offset alone.
const browserChunkSize = opfsChunkSize

// opfsWalk opens a directory named from the origin private file system's root
// down, the way the page walks to a reserved file.
func opfsWalk(t *testing.T, path []string) js.Value {
	t.Helper()

	dir, err := opfsAwaitCall(js.Global().Get("navigator").Get("storage"), "getDirectory")
	if err != nil {
		t.Fatalf("open the origin private file system: %v", err)
	}

	for _, name := range path {
		dir, err = opfsAwaitCall(dir, "getDirectoryHandle", name)
		if err != nil {
			t.Fatalf("open %q: %v", name, err)
		}
	}

	return dir
}

// browserFile opens the reserved file the way the page does: it is the page
// that creates it, because Go reserved a name and deliberately no handle.
func browserFile(t *testing.T, upload *BrowserUpload) js.Value {
	t.Helper()

	file, err := opfsAwaitCall(opfsWalk(t, upload.Dir()), "getFileHandle",
		upload.Name(), opfsCreate)
	if err != nil {
		t.Fatalf("open the reserved file: %v", err)
	}

	return file
}

// browserWrite streams a blob into the reserved file and hands back the sync
// access handle Go reads it through, in the order the page has to use: the
// writable stream is closed -- pipeTo does it -- before the handle is asked
// for.
func browserWrite(t *testing.T, upload *BrowserUpload, blob js.Value) js.Value {
	t.Helper()

	file := browserFile(t, upload)

	writable, err := opfsAwaitCall(file, "createWritable",
		map[string]any{"keepExistingData": false})
	if err != nil {
		t.Fatalf("open a writable stream: %v", err)
	}

	stream, err := opfsCall(blob, "stream")
	if err != nil {
		t.Fatalf("stream the blob: %v", err)
	}

	if _, err := opfsAwaitCall(stream, "pipeTo", writable); err != nil {
		t.Fatalf("pipe the blob into the file: %v", err)
	}

	handle, err := opfsAwaitCall(file, "createSyncAccessHandle")
	if err != nil {
		t.Fatalf("open a sync access handle: %v", err)
	}

	return handle
}

// browserBlob makes a blob of text, the way a picked file arrives.
func browserBlob(t *testing.T, text string) js.Value {
	t.Helper()

	return js.Global().Get("Blob").New([]any{text})
}

// TestBrowserUploadRoundTrip checks a file the page wrote arrives as a file a
// page function can read, with the size it was written at. Nothing about the
// content crosses the boundary: what does is a place to write and, afterwards,
// a handle on what was written.
func TestBrowserUploadRoundTrip(t *testing.T) {
	s, _ := opfsState(t)
	defer s.Destroy()

	upload, err := s.NewBrowserUpload()
	if err != nil {
		t.Fatalf("NewBrowserUpload: %v", err)
	}

	file, err := upload.Take("a.txt", browserWrite(t, upload, browserBlob(t, "hello file")))
	if err != nil {
		t.Fatalf("Take: %v", err)
	}

	s.PutFile("comp", file)

	if file.Name() != "a.txt" {
		t.Errorf("Name = %q, want a.txt", file.Name())
	}

	if file.Size() != 10 {
		t.Errorf("Size = %d, want 10", file.Size())
	}

	bs, err := s.GetFile("comp").Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	if string(bs) != "hello file" {
		t.Errorf("Bytes = %q, want hello file", bs)
	}
}

// TestBrowserUploadHoldsAreExclusive pins the one ordering this transport has
// to keep. A writable stream and a sync access handle are exclusive holds on
// the same file, so the page closes its stream and only then is Go handed a
// handle; either the other way round or both at once is
// NoModificationAllowedError.
//
// It is not a browser quirk to work around. A writable stream is the only way
// to copy a picked file without the whole of it going through the tab's heap,
// and a sync access handle is the only way to read one back without awaiting a
// promise, which is what an upload on the JavaScript callback stack cannot do.
// Each side needs the hold the other cannot have.
func TestBrowserUploadHoldsAreExclusive(t *testing.T) {
	s, _ := opfsState(t)
	defer s.Destroy()

	upload, err := s.NewBrowserUpload()
	if err != nil {
		t.Fatalf("NewBrowserUpload: %v", err)
	}

	file := browserFile(t, upload)

	writable, err := opfsAwaitCall(file, "createWritable",
		map[string]any{"keepExistingData": false})
	if err != nil {
		t.Fatalf("open a writable stream: %v", err)
	}

	// Go asking for the read side while the page is still writing.
	if _, err := opfsAwaitCall(file, "createSyncAccessHandle"); err == nil {
		t.Error("expect a sync access handle to be refused while the page is writing")
	} else if !strings.Contains(err.Error(), "NoModificationAllowedError") {
		t.Errorf("a sync access handle during a write failed with %v,"+
			" want NoModificationAllowedError", err)
	}

	if _, err := opfsAwaitCall(writable, "write", js.ValueOf("hello file")); err != nil {
		t.Fatalf("write: %v", err)
	}

	if _, err := opfsAwaitCall(writable, "close"); err != nil {
		t.Fatalf("close the writable stream: %v", err)
	}

	handle, err := opfsAwaitCall(file, "createSyncAccessHandle")
	if err != nil {
		t.Fatalf("open a sync access handle after the write was closed: %v", err)
	}

	// And the page asking for the write side back while Go is reading.
	if _, err := opfsAwaitCall(file, "createWritable",
		map[string]any{"keepExistingData": false}); err == nil {
		t.Error("expect a writable stream to be refused while Go holds the file")
	} else if !strings.Contains(err.Error(), "NoModificationAllowedError") {
		t.Errorf("a writable stream during a read failed with %v,"+
			" want NoModificationAllowedError", err)
	}

	taken, err := upload.Take("a.txt", handle)
	if err != nil {
		t.Fatalf("Take: %v", err)
	}

	bs, err := taken.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	if string(bs) != "hello file" {
		t.Errorf("Bytes = %q, want hello file", bs)
	}
}

// TestBrowserUploadDiscardLeavesNothing checks an upload that broke off halfway
// leaves neither a part-written file in the origin private file system nor
// anything a page function could read. Out of quota, cancelled, or a stream
// that broke all end here.
func TestBrowserUploadDiscardLeavesNothing(t *testing.T) {
	s, bodies := opfsState(t)
	defer s.Destroy()

	upload, err := s.NewBrowserUpload()
	if err != nil {
		t.Fatalf("NewBrowserUpload: %v", err)
	}

	file := browserFile(t, upload)

	writable, err := opfsAwaitCall(file, "createWritable",
		map[string]any{"keepExistingData": false})
	if err != nil {
		t.Fatalf("open a writable stream: %v", err)
	}

	if _, err := opfsAwaitCall(writable, "write", js.ValueOf("half a f")); err != nil {
		t.Fatalf("write: %v", err)
	}

	// What pipeTo does to the stream when anything on the way in fails: the
	// half that was written goes with the swap file rather than into place.
	if _, err := opfsAwaitCall(writable, "abort"); err != nil {
		t.Fatalf("abort the writable stream: %v", err)
	}

	upload.Discard()

	if s.GetFile("comp") != nil {
		t.Error("expect an upload that failed to reach no component")
	}

	// Taking it now is the worker coming back after the page gave up.
	handle, err := opfsAwaitCall(file, "createSyncAccessHandle")
	if err == nil {
		if _, err := upload.Take("a.txt", handle); err == nil {
			t.Error("expect a discarded upload not to be taken")
		}

		if _, err := opfsCall(handle, "close"); err != nil {
			t.Errorf("close the handle: %v", err)
		}
	}

	opfsWaitGone(t, bodies.dir, upload.Name(), false)
}

// TestBrowserUploadAppend checks the chunked path works on a file the page
// wrote. It is the same handle either way once the file is taken over, so a
// page function can go on adding to an upload the way a transport that carries
// a piece per call builds one up.
func TestBrowserUploadAppend(t *testing.T) {
	s, _ := opfsState(t)
	defer s.Destroy()

	upload, err := s.NewBrowserUpload()
	if err != nil {
		t.Fatalf("NewBrowserUpload: %v", err)
	}

	file, err := upload.Take("a.txt", browserWrite(t, upload, browserBlob(t, "hello")))
	if err != nil {
		t.Fatalf("Take: %v", err)
	}

	for _, part := range []string{" ", "file"} {
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

// TestBrowserUploadFromACallback checks both halves of the handover can be done
// from inside a JavaScript callback, which is where an upload arrives. Nothing
// on that path may wait on a promise: a call into Go holds the event loop for
// its whole length, so one awaited there would never settle and the tab would
// stop dead. If this regresses it hangs rather than fails, and the test times
// out.
func TestBrowserUploadFromACallback(t *testing.T) {
	s, bodies := opfsState(t)
	defer s.Destroy()

	opfsPooled(t, bodies)

	upload, err := onTimeout(t, func() (*BrowserUpload, error) {
		return s.NewBrowserUpload()
	})
	if err != nil {
		t.Fatalf("NewBrowserUpload from a callback: %v", err)
	}

	// Between the two calls the page has the file to itself, and everything
	// waited for is waited for here.
	handle := browserWrite(t, upload, browserBlob(t, "hello file"))

	file, err := onTimeout(t, func() (*File, error) {
		return upload.Take("a.txt", handle)
	})
	if err != nil {
		t.Fatalf("Take from a callback: %v", err)
	}

	if file.Size() != 10 {
		t.Errorf("Size = %d, want 10", file.Size())
	}
}

// onTimeout runs fn through a browser timer, so it is called with Go parked --
// the same way the bridge is reached, rather than Go calling itself.
func onTimeout[T any](t *testing.T, fn func() (T, error)) (T, error) {
	t.Helper()

	var (
		value T
		err   error
	)

	done := make(chan struct{})

	cb := js.FuncOf(func(_ js.Value, _ []js.Value) any {
		value, err = fn()
		close(done)
		return nil
	})
	defer cb.Release()

	js.Global().Call("setTimeout", cb, 0)
	<-done

	return value, err
}

// TestBrowserUploadStreamsABigFile checks a file far larger than the tab would
// hold goes in whole. The old transport made the whole of it a base64 string
// and handed that to Go to decode, so a file this size meant several copies in
// the heap at once, two of them a third larger again; this one is a stream into
// the file system on one side and reads at an offset on the other, and neither
// side ever holds more than a chunk.
func TestBrowserUploadStreamsABigFile(t *testing.T) {
	s, _ := opfsState(t)
	defer s.Destroy()

	upload, err := s.NewBrowserUpload()
	if err != nil {
		t.Fatalf("NewBrowserUpload: %v", err)
	}

	file, err := upload.Take("big.bin", browserWrite(t, upload, browserBigBlob(t)))
	if err != nil {
		// An origin with no room for it is the machine's answer, not a broken
		// transport, and it is the one this reports rather than fails on.
		if strings.Contains(err.Error(), "QuotaExceeded") {
			t.Skipf("no room in the origin for %d bytes: %v", browserUploadSize, err)
		}

		t.Fatalf("Take: %v", err)
	}

	if file.Size() != browserUploadSize {
		t.Fatalf("Size = %d, want %d", file.Size(), browserUploadSize)
	}

	fp, err := file.Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer fp.Close()

	// Spot checks rather than the whole file: reading it all back into Go would
	// be the very thing this transport stopped doing.
	offsets := []int64{
		0,
		browserChunkSize - 8,
		browserUploadSize / 2,
		browserUploadSize - 64,
	}

	got := make([]byte, 64)

	for _, off := range offsets {
		if _, err := io.ReadFull(io.NewSectionReader(fp, off, 64), got); err != nil {
			t.Fatalf("read at %d: %v", off, err)
		}

		for i, b := range got {
			if want := browserByteAt(off + int64(i)); b != want {
				t.Fatalf("byte %d is %d, want %d", off+int64(i), b, want)
			}
		}
	}
}

// browserByteAt is what the big file holds at an offset. The pattern repeats
// every chunk, so any byte of it is known without keeping the file around to
// compare against.
func browserByteAt(off int64) byte {
	return byte((off % browserChunkSize) % 251)
}

// browserBigBlob makes the big file as a blob, which is where a picked file of
// this size lives too: the browser keeps a blob's bytes of its own accord, out
// of the heap the wasm program shares. One chunk is made here and the blob is
// that chunk over and over, so Go holds a megabyte rather than the file.
func browserBigBlob(t *testing.T) js.Value {
	t.Helper()

	bs := make([]byte, browserChunkSize)
	for i := range bs {
		bs[i] = browserByteAt(int64(i))
	}

	chunk := opfsUint8Array.New(browserChunkSize)
	js.CopyBytesToJS(chunk, bs)

	parts := make([]any, browserUploadSize/browserChunkSize)
	for i := range parts {
		parts[i] = chunk
	}

	return js.Global().Get("Blob").New(parts)
}

// TestBrowserUploadOutlivingItsStateIsCleanedUp checks a state whose directory
// the page is still writing in is removed once the page lets go, rather than
// left for the origin's quota to carry.
//
// A page switch mid-upload is where this happens: the session goes, its
// directory is removed, and the browser refuses because a writable stream is
// still open on a file inside. The refusal is temporary -- the write ends one
// way or the other -- so the removal asks again.
func TestBrowserUploadOutlivingItsStateIsCleanedUp(t *testing.T) {
	// Short enough for a test to sit through a try or two.
	gap := opfsRemoveGap
	opfsRemoveGap = 20 * time.Millisecond
	t.Cleanup(func() { opfsRemoveGap = gap })

	root, err := opfsStateRoot.get()
	if err != nil {
		t.Fatalf("open the state root: %v", err)
	}

	s, bodies := opfsState(t)

	upload, err := s.NewBrowserUpload()
	if err != nil {
		t.Fatalf("NewBrowserUpload: %v", err)
	}

	file := browserFile(t, upload)

	writable, err := opfsAwaitCall(file, "createWritable",
		map[string]any{"keepExistingData": false})
	if err != nil {
		t.Fatalf("open a writable stream: %v", err)
	}

	if _, err := opfsAwaitCall(writable, "write", js.ValueOf("half a f")); err != nil {
		t.Fatalf("write: %v", err)
	}

	// The page switch: the session is gone while the page is still writing.
	s.Destroy()

	// Long enough for a removal that only tried once to have tried and given
	// up, so this fails rather than passes by being slow.
	time.Sleep(10 * opfsRemoveGap)

	if !opfsEntry(t, root, bodies.name, true) {
		t.Fatal("expect the directory to survive while the page is writing in it")
	}

	if _, err := opfsAwaitCall(writable, "abort"); err != nil {
		t.Fatalf("abort the writable stream: %v", err)
	}

	opfsWaitGone(t, root, bodies.name, true)
}
