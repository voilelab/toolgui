//go:build js && wasm

package tgframe

import (
	"syscall/js"
	"testing"
	"time"
)

// These are the browser half of a download: Go writes the file into the origin
// private file system and the page reads it from where
// [Download.BrowserLocation] says, which is what keeps the bytes out of the
// pack.

// downloadSize is what the test offers. Well past what a pack should carry as
// base64, and past a chunk of a write, so a file written or read in one go
// rather than a chunk at a time would not pass here.
const downloadSize = 8 << 20

// pageReadsDownload does the page's half. getFile answers a blob backed by
// what is in the file system rather than a copy of it, so nothing here holds
// the file either; the bytes are checked a chunk at a time through a slice of
// that blob.
func pageReadsDownload(t *testing.T, d *Download) {
	t.Helper()

	dir, name, err := d.BrowserLocation()
	if err != nil {
		t.Fatalf("BrowserLocation: %v", err)
	}

	handle, err := opfsAwaitCall(opfsWalk(t, dir), "getFileHandle", name)
	if err != nil {
		t.Fatalf("open the download: %v", err)
	}

	blob, err := opfsAwaitCall(handle, "getFile")
	if err != nil {
		t.Fatalf("read the download: %v", err)
	}

	if got := int64(blob.Get("size").Float()); got != downloadSize {
		t.Fatalf("the page sees %d bytes, want %d", got, downloadSize)
	}

	for off := int64(0); off < downloadSize; off += opfsChunkSize {
		part, err := opfsCall(blob, "slice", float64(off),
			float64(off+opfsChunkSize))
		if err != nil {
			t.Fatalf("slice at %d: %v", off, err)
		}

		buf, err := opfsAwaitCall(part, "arrayBuffer")
		if err != nil {
			t.Fatalf("read at %d: %v", off, err)
		}

		bs := make([]byte, opfsChunkSize)
		js.CopyBytesToGo(bs, opfsUint8Array.New(buf))

		for i, b := range bs {
			if want := downloadByteAt(off + int64(i)); b != want {
				t.Fatalf("byte %d = %d, want %d", off+int64(i), b, want)
			}
		}
	}
}

// downloadByteAt is what the file holds at an offset, so any byte of it is
// known without keeping a copy to compare against.
func downloadByteAt(off int64) byte {
	return byte(off % 251)
}

// downloadBytes builds the file. It is the one copy of it the test holds, and
// it goes no further than the store's write.
func downloadBytes() []byte {
	bs := make([]byte, downloadSize)
	for i := range bs {
		bs[i] = downloadByteAt(int64(i))
	}

	return bs
}

// TestBrowserDownloadIsReadableByThePage checks the page reads every byte of
// what the run offered, while Go still holds a sync access handle on the file.
// The handle is exclusive against a second one and against a writable stream,
// not against getFile.
func TestBrowserDownloadIsReadableByThePage(t *testing.T) {
	s, bodies := opfsState(t)
	defer s.Destroy()
	opfsPooled(t, bodies)

	d, err := s.SetDownload("comp", "big.bin", "application/octet-stream",
		downloadBytes())
	if err != nil {
		t.Fatalf("SetDownload: %v", err)
	}

	dir, _, err := d.BrowserLocation()
	if err != nil {
		t.Fatalf("BrowserLocation: %v", err)
	}

	if len(dir) != 2 || dir[0] != opfsRootName {
		t.Errorf("dir = %v, want the state's own directory under %q",
			dir, opfsRootName)
	}

	pageReadsDownload(t, d)
}

// TestBrowserDownloadReplacedTakesItsFile checks a run that offers different
// bytes leaves nothing behind: the file the token before it named is gone from
// the origin private file system, rather than held against the origin's quota
// for the life of the tab.
func TestBrowserDownloadReplacedTakesItsFile(t *testing.T) {
	s, bodies := opfsState(t)
	defer s.Destroy()
	opfsPooled(t, bodies)

	first, err := s.SetDownload("comp", "a.txt", "text/plain", []byte("old"))
	if err != nil {
		t.Fatalf("SetDownload: %v", err)
	}

	dir, name, err := first.BrowserLocation()
	if err != nil {
		t.Fatalf("BrowserLocation: %v", err)
	}

	if _, err := s.SetDownload("comp", "a.txt", "text/plain",
		[]byte("new")); err != nil {
		t.Fatalf("SetDownload again: %v", err)
	}

	// The removal is asked for without waiting -- the store's methods may run
	// where nothing can -- so the file goes within a few turns of the event
	// loop rather than at once.
	for range 400 {
		if _, err := opfsAwaitCall(opfsWalk(t, dir), "getFileHandle", name); err != nil {
			return
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Error("expect the replaced file to be removed")
}
