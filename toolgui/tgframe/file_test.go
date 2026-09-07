package tgframe

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestStateWriteFile(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	file, err := s.WriteFile("comp", "a.txt", strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if file.Name() != "a.txt" {
		t.Errorf("Name = %q, want a.txt", file.Name())
	}

	if file.Size() != 5 {
		t.Errorf("Size = %d, want 5", file.Size())
	}

	bs, err := s.GetFile("comp").Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	if string(bs) != "hello" {
		t.Errorf("Bytes = %q, want hello", bs)
	}
}

func TestStateGetFileMissing(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	if s.GetFile("comp") != nil {
		t.Error("expect no file before an upload")
	}
}

// TestStateWriteFileReplaces checks a second upload to the same component
// takes the first one's place, and its disk with it.
func TestStateWriteFileReplaces(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	first, err := s.WriteFile("comp", "a.txt", strings.NewReader("old"))
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	second, err := s.WriteFile("comp", "b.txt", strings.NewReader("new"))
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	bs, err := second.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	if string(bs) != "new" {
		t.Errorf("Bytes = %q, want new", bs)
	}

	if _, err := os.Stat(first.path); !os.IsNotExist(err) {
		t.Errorf("expect the replaced file to be gone, stat gave %v", err)
	}
}

func TestStateAppendFile(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	if _, err := s.WriteFile("comp", "a.txt", strings.NewReader("hello")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	file, err := s.AppendFile("comp", strings.NewReader(" file"))
	if err != nil {
		t.Fatalf("AppendFile: %v", err)
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

func TestStateAppendFileWithoutWrite(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	if _, err := s.AppendFile("comp", strings.NewReader("x")); err == nil {
		t.Error("expect an error appending to a file that was never written")
	}
}

// TestFileOpen checks the reader a page gets can both stream and read at an
// offset, which is what archive and image decoders ask for.
func TestFileOpen(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	file, err := s.WriteFile("comp", "a.txt", strings.NewReader("hello file"))
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fp, err := file.Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer fp.Close()

	bs, err := io.ReadAll(fp)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	if string(bs) != "hello file" {
		t.Errorf("ReadAll = %q, want hello file", bs)
	}

	at := make([]byte, 4)
	if _, err := fp.ReadAt(at, 6); err != nil {
		t.Fatalf("ReadAt: %v", err)
	}

	if string(at) != "file" {
		t.Errorf("ReadAt = %q, want file", at)
	}
}

// TestStateDestroyRemovesFiles checks a state that goes away takes the
// uploads it was holding with it.
func TestStateDestroyRemovesFiles(t *testing.T) {
	s := NewState()

	file, err := s.WriteFile("comp", "a.txt", strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	s.Destroy()

	if _, err := os.Stat(file.path); !os.IsNotExist(err) {
		t.Errorf("expect the file to be gone, stat gave %v", err)
	}

	if s.GetFile("comp") != nil {
		t.Error("expect no file after Destroy")
	}
}

// TestStateDestroyWithoutFiles checks a state that never saw an upload leaves
// no directory behind to clean up.
func TestStateDestroyWithoutFiles(t *testing.T) {
	s := NewState()
	s.Destroy()

	if s.fileDir != "" {
		t.Errorf("fileDir = %q, want empty", s.fileDir)
	}
}

func TestStateSetFile(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	if _, err := s.SetFile("comp", "a.txt", []byte("hello")); err != nil {
		t.Fatalf("SetFile: %v", err)
	}

	bs, err := s.GetFile("comp").Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	if !bytes.Equal(bs, []byte("hello")) {
		t.Errorf("Bytes = %q, want hello", bs)
	}
}
