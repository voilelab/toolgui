package tgframe

import (
	"bytes"
	"io"
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
// takes the first one's place. Where the first one's content goes is the
// storage's business, checked where each storage is implemented.
func TestStateWriteFileReplaces(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	if _, err := s.WriteFile("comp", "a.txt", strings.NewReader("old")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	second, err := s.WriteFile("comp", "b.txt", strings.NewReader("new"))
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if s.GetFile("comp") != second {
		t.Fatal("expect the key to hold the second file")
	}

	bs, err := second.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	if string(bs) != "new" {
		t.Errorf("Bytes = %q, want new", bs)
	}
}

// TestStateNewFilePutFile checks a file being filled a chunk at a time is
// invisible under its key until it's handed over.
func TestStateNewFilePutFile(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	file, err := s.NewFile("a.txt")
	if err != nil {
		t.Fatalf("NewFile: %v", err)
	}

	if err := file.Append(strings.NewReader("hello")); err != nil {
		t.Fatalf("Append: %v", err)
	}

	if s.GetFile("comp") != nil {
		t.Error("expect no file under the key before PutFile")
	}

	if err := file.Append(strings.NewReader(" file")); err != nil {
		t.Fatalf("Append: %v", err)
	}

	s.PutFile("comp", file)

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

// TestStateNewFileIsSeparate checks two files opened at once get their own
// content, which is what keeps two uploads racing for one key apart.
func TestStateNewFileIsSeparate(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	first, err := s.NewFile("first.txt")
	if err != nil {
		t.Fatalf("NewFile: %v", err)
	}

	second, err := s.NewFile("second.txt")
	if err != nil {
		t.Fatalf("NewFile: %v", err)
	}

	if err := first.Append(strings.NewReader("one")); err != nil {
		t.Fatalf("Append: %v", err)
	}

	if err := second.Append(strings.NewReader("two")); err != nil {
		t.Fatalf("Append: %v", err)
	}

	bs, err := first.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	if string(bs) != "one" {
		t.Errorf("Bytes = %q, want one", bs)
	}
}

// TestStateCloneSharesFiles checks a clone and its original hand out separate
// paths, so neither can overwrite what the other stored.
func TestStateCloneSharesFiles(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	if _, err := s.WriteFile("comp", "a.txt", strings.NewReader("original")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	clone := s.Clone()
	if _, err := clone.WriteFile("other", "b.txt", strings.NewReader("clone")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	bs, err := s.GetFile("comp").Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	if string(bs) != "original" {
		t.Errorf("Bytes = %q, want original", bs)
	}

	// The store is shared, so what one writes the other reads.
	if clone.GetFile("comp") == nil {
		t.Error("expect the clone to see the original's file")
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

// TestStateDestroyRemovesFiles checks a state that goes away lets go of the
// uploads it was holding.
func TestStateDestroyRemovesFiles(t *testing.T) {
	s := NewState()

	if _, err := s.WriteFile("comp", "a.txt", strings.NewReader("hello")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	s.Destroy()

	if s.GetFile("comp") != nil {
		t.Error("expect no file after Destroy")
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
