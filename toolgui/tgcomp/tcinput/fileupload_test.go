package tcinput

import (
	"io"
	"strings"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

const testFileUploadID = "fileupload_component_File"

// newFileUploadContainer builds the container the component reads its state
// through, so that a test can seed s and then hand it straight to FileUpload.
func newFileUploadContainer(s *tgframe.State) *tgframe.Container {
	return tgframe.NewContainer("test", s, func(pack tgframe.NotifyPack) {})
}

func TestFileUploadWithoutPick(t *testing.T) {
	s := tgframe.NewState()
	defer s.Destroy()

	if FileUpload(newFileUploadContainer(s), "File", "") != nil {
		t.Error("expect no file object before a pick")
	}
}

// TestFileUploadWithoutContent checks a pick whose upload never landed reads
// as no file, rather than as a file with nothing in it.
func TestFileUploadWithoutContent(t *testing.T) {
	s := tgframe.NewState()
	defer s.Destroy()

	s.Set(testFileUploadID, map[string]any{"name": "a.txt", "size": 5})

	if FileUpload(newFileUploadContainer(s), "File", "") != nil {
		t.Error("expect no file object when the content is missing")
	}
}

func TestFileUpload(t *testing.T) {
	s := tgframe.NewState()
	defer s.Destroy()

	s.Set(testFileUploadID, map[string]any{
		"name": "a.txt",
		"type": "text/plain",
		// The browser's size is not what a reader gets, so the component
		// answers with the size of what it stored.
		"size": 100,
	})

	if _, err := s.WriteFile(testFileUploadID, "a.txt",
		strings.NewReader("hello")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fileObj := FileUpload(newFileUploadContainer(s), "File", "")
	if fileObj == nil {
		t.Fatal("expect a file object")
	}

	if fileObj.Name != "a.txt" || fileObj.Type != "text/plain" {
		t.Errorf("got %+v, want a.txt of text/plain", fileObj)
	}

	if fileObj.Size != 5 {
		t.Errorf("Size = %d, want 5", fileObj.Size)
	}

	fp, err := fileObj.Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer fp.Close()

	bs, err := io.ReadAll(fp)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	if string(bs) != "hello" {
		t.Errorf("ReadAll = %q, want hello", bs)
	}

	bs, err = fileObj.Bytes()
	if err != nil {
		t.Fatalf("Bytes: %v", err)
	}

	if string(bs) != "hello" {
		t.Errorf("Bytes = %q, want hello", bs)
	}
}

// TestFileObjectWithoutFile checks reading a file object that was never tied
// to content fails instead of panicking.
func TestFileObjectWithoutFile(t *testing.T) {
	fileObj := &FileObject{Name: "a.txt"}

	if _, err := fileObj.Open(); err == nil {
		t.Error("expect an error opening a file object with no content")
	}

	if _, err := fileObj.Bytes(); err == nil {
		t.Error("expect an error reading a file object with no content")
	}
}
