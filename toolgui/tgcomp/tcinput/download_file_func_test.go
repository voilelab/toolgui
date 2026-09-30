package tcinput

import (
	"errors"
	"io"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgjson"
)

// drawDownloadFileFunc draws the component and returns what it reported, the
// number of times gen ran, and every component the container sent.
func drawDownloadFileFunc(t *testing.T, s *tgframe.State, body []byte,
	genErr error, conf ...*DownloadFileConf) (bool, int, []map[string]any) {
	t.Helper()

	var packs []tgframe.NotifyPack
	c := tgframe.NewContainer("test", s, func(pack tgframe.NotifyPack) {
		packs = append(packs, pack)
	})

	calls := 0
	got := DownloadFileFunc(c, "Export", func() ([]byte, error) {
		calls++
		return body, genErr
	}, conf...)

	var comps []map[string]any
	for _, pack := range packs {
		bs, err := tgjson.Marshal(pack)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		var out struct {
			Component map[string]any `json:"component"`
		}
		if err := tgjson.Unmarshal(bs, &out); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		comps = append(comps, out.Component)
	}

	return got, calls, comps
}

// TestDownloadFileFuncWithoutClick checks a run that is not about the button
// neither makes the file nor offers one.
func TestDownloadFileFuncWithoutClick(t *testing.T) {
	s := tgframe.NewState()
	defer s.Destroy()

	got, calls, comps := drawDownloadFileFunc(t, s, []byte("body"), nil)
	if got || calls != 0 {
		t.Errorf("clicked = %v, gen ran %d times, want false and 0", got, calls)
	}

	if len(comps) != 1 {
		t.Fatalf("got %d components, want 1", len(comps))
	}

	props := comps[0]
	if props["lazy"] != true {
		t.Error("expect the pack to mark the button lazy")
	}
	if props["token"] != "" || props["serial"] != nil {
		t.Errorf("token/serial = %v/%v, want none", props["token"], props["serial"])
	}
}

// TestDownloadFileFuncOnClick checks the click run makes the file, offers it
// by token, and marks it with a serial that is new on every click, even when
// the bytes and so the token are not.
func TestDownloadFileFuncOnClick(t *testing.T) {
	s := tgframe.NewState()
	defer s.Destroy()

	conf := &DownloadFileConf{Filename: "out.txt", MIME: "text/plain"}
	s.SetClickID("download_file_component_Export")

	got, calls, comps := drawDownloadFileFunc(t, s, []byte("body"), nil, conf)
	if !got || calls != 1 {
		t.Fatalf("clicked = %v, gen ran %d times, want true and 1", got, calls)
	}

	props := comps[0]
	token, _ := props["token"].(string)
	serial, _ := props["serial"].(string)
	if token == "" || serial == "" {
		t.Fatalf("token/serial = %q/%q, want both", token, serial)
	}
	if props["filename"] != "out.txt" || props["mime"] != "text/plain" {
		t.Errorf("filename/mime = %v/%v", props["filename"], props["mime"])
	}

	fp, err := s.GetDownload(token).Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer fp.Close()

	bs, err := io.ReadAll(fp)
	if err != nil || string(bs) != "body" {
		t.Errorf("stored %q, %v, want body", bs, err)
	}

	_, _, again := drawDownloadFileFunc(t, s, []byte("body"), nil, conf)
	if again[0]["token"] != token {
		t.Error("same bytes, expect the same token")
	}
	if again[0]["serial"] == serial {
		t.Error("expect a new serial on every click")
	}
}

// TestDownloadFileFuncError checks a failing gen leaves the button to retry
// with and shows the error, and offers nothing.
func TestDownloadFileFuncError(t *testing.T) {
	s := tgframe.NewState()
	defer s.Destroy()

	s.SetClickID("download_file_component_Export")

	got, _, comps := drawDownloadFileFunc(t, s, nil, errors.New("boom"))
	if !got {
		t.Error("clicked = false, want true")
	}

	if len(comps) != 2 {
		t.Fatalf("got %d components, want the button and an error", len(comps))
	}
	if comps[0]["token"] != "" || comps[0]["serial"] != nil {
		t.Error("expect no file offered")
	}
	if comps[1]["name"] != tgframe.ErrorComponentName {
		t.Errorf("second component = %v, want the error", comps[1]["name"])
	}
}
