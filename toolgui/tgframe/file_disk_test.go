//go:build !js || !wasm

package tgframe

import (
	"os"
	"strings"
	"testing"
)

// TestStateFilesOnDisk checks a server keeps an upload on disk rather than in
// the process, which is the whole point of streaming it there.
func TestStateFilesOnDisk(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	file, err := s.WriteFile("comp", "a.txt", strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	body, ok := file.body.(*diskBody)
	if !ok {
		t.Fatalf("body is %T, want a file on disk", file.body)
	}

	bs, err := os.ReadFile(body.path)
	if err != nil {
		t.Fatalf("read the stored file: %v", err)
	}

	if string(bs) != "hello" {
		t.Errorf("the stored file holds %q, want hello", bs)
	}

	// A file the key no longer holds is deleted: a session that uploads all
	// day shouldn't fill the disk with them.
	if _, err := s.WriteFile("comp", "b.txt", strings.NewReader("new")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := os.Stat(body.path); !os.IsNotExist(err) {
		t.Errorf("expect the replaced file to be gone, stat gave %v", err)
	}

	dir := s.files.bodies.(*diskBodies).dir

	s.Destroy()

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("expect the directory to be gone, stat gave %v", err)
	}
}

// TestStateDestroyWithoutFiles checks a state that never saw an upload leaves
// no directory behind to clean up.
func TestStateDestroyWithoutFiles(t *testing.T) {
	s := NewState()
	s.Destroy()

	if dir := s.files.bodies.(*diskBodies).dir; dir != "" {
		t.Errorf("file dir = %q, want empty", dir)
	}
}
