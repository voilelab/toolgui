//go:build js && wasm

package tgwasm

import (
	"encoding/json"
	"strings"
	"syscall/js"
	"testing"
)

// askDownload asks the bridge where a token's file is, the way the worker
// does.
func askDownload(t *testing.T, b *bridge, token string) downloadSlot {
	t.Helper()

	answer := b.jsDownloadFile(js.Undefined(), []js.Value{js.ValueOf(token)})

	var got downloadSlot
	if err := json.Unmarshal([]byte(answer.(string)), &got); err != nil {
		t.Fatalf("read the download slot: %v", err)
	}

	return got
}

// readDownload does the page's half: it walks to the file the slot names and
// reads it with getFile, which is what hands the tab a blob backed by the file
// rather than a copy of it. The file is not created here -- it is Go's own and
// has to be there already.
func readDownload(t *testing.T, s downloadSlot) string {
	t.Helper()

	dir, err := await(js.Global().Get("navigator").Get("storage").Call("getDirectory"))
	if err != nil {
		t.Fatalf("open the origin private file system: %v", err)
	}

	for _, name := range s.Dir {
		dir, err = await(dir.Call("getDirectoryHandle", name))
		if err != nil {
			t.Fatalf("open %q: %v", name, err)
		}
	}

	file, err := await(dir.Call("getFileHandle", s.Name))
	if err != nil {
		t.Fatalf("open the download: %v", err)
	}

	blob, err := await(file.Call("getFile"))
	if err != nil {
		t.Fatalf("read the download: %v", err)
	}

	text, err := await(blob.Call("text"))
	if err != nil {
		t.Fatalf("read the blob: %v", err)
	}

	return text.String()
}

// TestDownloadFileNamesAFileThePageCanRead is the browser half of a download:
// the bridge answers with where the file is, and the page reads it there --
// while Go still holds a sync access handle on it, which is exclusive against
// another handle but not against this read.
func TestDownloadFileNamesAFileThePageCanRead(t *testing.T) {
	b := newBridge(testApp())
	b.jsStart(js.Undefined(), []js.Value{js.ValueOf("index")})
	defer stopped(b)()

	content := strings.Repeat("payload", 1000)

	download, err := b.state.SetDownload("comp", "a.txt", "text/plain",
		[]byte(content))
	if err != nil {
		t.Fatalf("SetDownload: %v", err)
	}

	slot := askDownload(t, b, download.Token())
	if slot.Error != "" {
		t.Fatalf("downloadFile: %s", slot.Error)
	}

	if slot.Name == "" || len(slot.Dir) == 0 {
		t.Fatalf("slot = %+v, want a file to read", slot)
	}

	if got := readDownload(t, slot); got != content {
		t.Errorf("the page read %d bytes, want the %d offered, identical",
			len(got), len(content))
	}
}

// TestDownloadFileOfAnUnknownToken checks a token nothing handed out names no
// file, which is also the answer for one a later run replaced.
func TestDownloadFileOfAnUnknownToken(t *testing.T) {
	b := newBridge(testApp())
	b.jsStart(js.Undefined(), []js.Value{js.ValueOf("index")})
	defer stopped(b)()

	if got := askDownload(t, b, "NOTATOKEN"); got.Error == "" {
		t.Errorf("slot = %+v, want no such download", got)
	}
}

// TestDownloadFileAfterAPageSwitch checks a token does not outlive the state
// that offered it. A page switch is a new state, and the file the old one kept
// goes with it.
func TestDownloadFileAfterAPageSwitch(t *testing.T) {
	b := newBridge(testApp())
	b.jsStart(js.Undefined(), []js.Value{js.ValueOf("index")})
	defer stopped(b)()

	download, err := b.state.SetDownload("comp", "a.txt", "text/plain",
		[]byte("gone"))
	if err != nil {
		t.Fatalf("SetDownload: %v", err)
	}

	b.jsStart(js.Undefined(), []js.Value{js.ValueOf("other")})

	if got := askDownload(t, b, download.Token()); got.Error == "" {
		t.Errorf("slot = %+v, want no such download", got)
	}
}

// TestDownloadFileWithoutASession checks the bridge answers rather than
// panicking when the page asks before start.
func TestDownloadFileWithoutASession(t *testing.T) {
	b := newBridge(testApp())

	if got := askDownload(t, b, "whatever"); got.Error == "" {
		t.Errorf("slot = %+v, want no session", got)
	}
}
