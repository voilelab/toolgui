//go:build js && wasm

package tgframe

import (
	"strings"
	"testing"
)

// TestStateFilesInMemory checks the browser keeps an upload in memory. There
// is no filesystem in a tab: wasm_exec.js answers open and mkdir with ENOSYS,
// so a store that reached for one would fail every upload.
func TestStateFilesInMemory(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	file, err := s.WriteFile("comp", "a.txt", strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	body, ok := file.body.(*memBody)
	if !ok {
		t.Fatalf("body is %T, want a file in memory", file.body)
	}

	if string(body.bs) != "hello" {
		t.Errorf("the stored file holds %q, want hello", body.bs)
	}

	// A file the key no longer holds is dropped, so the tab isn't left
	// holding every file the user ever picked.
	if _, err := s.WriteFile("comp", "b.txt", strings.NewReader("new")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if body.bs != nil {
		t.Errorf("expect the replaced file to be dropped, it holds %q", body.bs)
	}
}
