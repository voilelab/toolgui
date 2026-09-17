package tgframe

import (
	"io"
	"testing"
)

// readDownload reads the whole file behind a download, the way a transport
// serving it does.
func readDownload(t *testing.T, d *Download) string {
	t.Helper()

	fp, err := d.Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer fp.Close()

	bs, err := io.ReadAll(fp)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	return string(bs)
}

// TestSetDownloadKeepsBytesOutOfTheToken checks a download is fetched by a
// token that says nothing about what it names, and hands back what was
// offered.
func TestSetDownloadKeepsBytesOutOfTheToken(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	d, err := s.SetDownload("comp", "report.csv", "text/csv", []byte("a,b\n1,2\n"))
	if err != nil {
		t.Fatalf("SetDownload: %v", err)
	}

	if d.Token() == "" {
		t.Fatal("expect a token")
	}

	if d.Name() != "report.csv" || d.MIME() != "text/csv" {
		t.Errorf("Name/MIME = %q/%q, want report.csv/text/csv", d.Name(), d.MIME())
	}

	if d.Size() != 8 {
		t.Errorf("Size = %d, want 8", d.Size())
	}

	if got := readDownload(t, d); got != "a,b\n1,2\n" {
		t.Errorf("content = %q, want the offered bytes", got)
	}

	if got := s.GetDownload(d.Token()); got != d {
		t.Errorf("GetDownload = %v, want the download just offered", got)
	}
}

// TestSetDownloadSameBytesKeepsTheToken checks a rerun that offers the same
// file again changes nothing: the pack stays as it was and the token the
// client already holds goes on working.
func TestSetDownloadSameBytesKeepsTheToken(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	first, err := s.SetDownload("comp", "a.txt", "text/plain", []byte("same"))
	if err != nil {
		t.Fatalf("SetDownload: %v", err)
	}

	second, err := s.SetDownload("comp", "a.txt", "text/plain", []byte("same"))
	if err != nil {
		t.Fatalf("SetDownload again: %v", err)
	}

	if second.Token() != first.Token() {
		t.Errorf("token rotated on the same bytes: %q then %q",
			first.Token(), second.Token())
	}

	if got := readDownload(t, second); got != "same" {
		t.Errorf("content = %q, want same", got)
	}
}

// TestSetDownloadNewBytesKeepTheShownToken checks what a client holding a
// token can rely on: different bytes are a new file under a new token, and the
// token of the run before stays fetchable -- that is the button still on the
// screen until the replacement pack lands.
func TestSetDownloadNewBytesKeepTheShownToken(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	first, err := s.SetDownload("comp", "a.txt", "text/plain", []byte("old"))
	if err != nil {
		t.Fatalf("SetDownload: %v", err)
	}

	second, err := s.SetDownload("comp", "a.txt", "text/plain", []byte("new"))
	if err != nil {
		t.Fatalf("SetDownload again: %v", err)
	}

	if second.Token() == first.Token() {
		t.Fatal("expect a new token for new bytes")
	}

	if got := s.GetDownload(second.Token()); got == nil {
		t.Fatal("expect the new token to serve")
	} else if content := readDownload(t, got); content != "new" {
		t.Errorf("content = %q, want new", content)
	}

	// The click that was on screen when the rerun landed still saves what it
	// was offering.
	if got := s.GetDownload(first.Token()); got == nil {
		t.Error("expect the token of the run before to still serve")
	} else if content := readDownload(t, got); content != "old" {
		t.Errorf("content = %q, want old", content)
	}
}

// TestSetDownloadKeepsOneGenerationOnly checks the other end of that: a third
// run retires the first, so a page reoffering on every run holds two files per
// download rather than a history of them.
func TestSetDownloadKeepsOneGenerationOnly(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	tokens := make([]string, 0, 3)
	for _, body := range []string{"one", "two", "three"} {
		d, err := s.SetDownload("comp", "a.txt", "text/plain", []byte(body))
		if err != nil {
			t.Fatalf("SetDownload %q: %v", body, err)
		}

		tokens = append(tokens, d.Token())
	}

	if got := s.GetDownload(tokens[0]); got != nil {
		t.Error("expect the token two runs back to be retired")
	}

	for _, token := range tokens[1:] {
		if got := s.GetDownload(token); got == nil {
			t.Errorf("expect %q to still serve", token)
		}
	}
}

// TestSetDownloadPerOwner checks two components offering a file each keep
// their own, the way two fileuploads keep their own upload.
func TestSetDownloadPerOwner(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	one, err := s.SetDownload("one", "a.txt", "text/plain", []byte("first"))
	if err != nil {
		t.Fatalf("SetDownload: %v", err)
	}

	two, err := s.SetDownload("two", "a.txt", "text/plain", []byte("second"))
	if err != nil {
		t.Fatalf("SetDownload: %v", err)
	}

	if one.Token() == two.Token() {
		t.Fatal("expect a token of its own per component")
	}

	if got := readDownload(t, s.GetDownload(one.Token())); got != "first" {
		t.Errorf("content = %q, want first", got)
	}

	if got := readDownload(t, s.GetDownload(two.Token())); got != "second" {
		t.Errorf("content = %q, want second", got)
	}
}

// TestGetDownloadOfAnotherState checks a token is only ever found in the state
// that offered it. It is what stops a token being a bearer of anything on its
// own: one page's output cannot be read through another page's connection.
func TestGetDownloadOfAnotherState(t *testing.T) {
	mine := NewState()
	defer mine.Destroy()

	theirs := NewState()
	defer theirs.Destroy()

	d, err := theirs.SetDownload("comp", "a.txt", "text/plain", []byte("secret"))
	if err != nil {
		t.Fatalf("SetDownload: %v", err)
	}

	if got := mine.GetDownload(d.Token()); got != nil {
		t.Error("expect another state's token to be found nowhere here")
	}

	if got := mine.GetDownload(""); got != nil {
		t.Error("expect no download for an empty token")
	}

	if got := mine.GetDownload("not-a-token"); got != nil {
		t.Error("expect no download for a token nothing handed out")
	}
}

// TestCloneSharesDownloads checks the clone a run draws on and the state a
// transport holds are the same store. The run offers the file; the fetch
// arrives on the other one.
func TestCloneSharesDownloads(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	run := s.Clone()

	d, err := run.SetDownload("comp", "a.txt", "text/plain", []byte("shared"))
	if err != nil {
		t.Fatalf("SetDownload: %v", err)
	}

	if got := s.GetDownload(d.Token()); got != d {
		t.Fatal("expect the run's download on the state it was cloned from")
	}
}

// TestDeleteReleasesTheDownload is what a cleared slot ends up calling: a
// download whose component has left the page goes with it, token and bytes
// both, rather than being held until the state ends.
func TestDeleteReleasesTheDownload(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	d, err := s.SetDownload("comp", "a.txt", "text/plain", []byte("gone"))
	if err != nil {
		t.Fatalf("SetDownload: %v", err)
	}

	s.Delete("comp")

	if got := s.GetDownload(d.Token()); got != nil {
		t.Error("expect the token of a released component to be retired")
	}

	// The next component to land on that id offers its own file, rather than
	// finding the one before it.
	again, err := s.SetDownload("comp", "a.txt", "text/plain", []byte("gone"))
	if err != nil {
		t.Fatalf("SetDownload again: %v", err)
	}

	if again.Token() == d.Token() {
		t.Error("expect a token of its own after the release")
	}

	if got := readDownload(t, again); got != "gone" {
		t.Errorf("content = %q, want gone", got)
	}
}
