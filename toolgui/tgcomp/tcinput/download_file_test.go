package tcinput

import (
	"crypto/md5"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgjson"
)

// drawDownloadFile draws the component against a state of the test's own and
// returns the json the container would have sent the client for it.
func drawDownloadFile(t *testing.T, s *tgframe.State, text string, body []byte,
	conf ...*DownloadFileConf) map[string]any {
	t.Helper()

	var packs []tgframe.NotifyPack
	c := tgframe.NewContainer("test", s, func(pack tgframe.NotifyPack) {
		packs = append(packs, pack)
	})

	DownloadFile(c, text, body, conf...)

	if len(packs) != 1 {
		t.Fatalf("got %d packs, want 1", len(packs))
	}

	bs, err := tgjson.Marshal(packs[0])
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out struct {
		Component map[string]any `json:"component"`
	}
	if err := tgjson.Unmarshal(bs, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	return out.Component
}

// TestDownloadFilePackCarriesATokenOnly is the point of the component: the
// pack says where to fetch the file and nothing about what is in it, so a big
// output does not travel through the render tree as base64.
func TestDownloadFilePackCarriesATokenOnly(t *testing.T) {
	s := tgframe.NewState()
	defer s.Destroy()

	body := []byte(strings.Repeat("secret-payload", 100))

	props := drawDownloadFile(t, s, "Report", body, &DownloadFileConf{
		Filename: "report.csv",
		MIME:     "text/csv",
	})

	token, _ := props["token"].(string)
	if token == "" {
		t.Fatal("expect a token in the pack")
	}

	if props["filename"] != "report.csv" || props["mime"] != "text/csv" {
		t.Errorf("filename/mime = %v/%v, want report.csv/text/csv",
			props["filename"], props["mime"])
	}

	if _, ok := props["uri"]; ok {
		t.Error("expect no uri in the pack")
	}

	for key, value := range props {
		if s, ok := value.(string); ok && s != token &&
			strings.Contains(s, "secret-payload") {
			t.Errorf("%q carries the content", key)
		}
	}

	download := s.GetDownload(token)
	if download == nil {
		t.Fatal("expect the token to name a download on the state")
	}

	fp, err := download.Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer fp.Close()

	got, err := io.ReadAll(fp)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if string(got) != string(body) {
		t.Error("the stored file is not what was offered")
	}
}

// TestDownloadFileDefaults checks the type and the filename a caller left out.
func TestDownloadFileDefaults(t *testing.T) {
	s := tgframe.NewState()
	defer s.Destroy()

	body := []byte("123")

	props := drawDownloadFile(t, s, "Download", body)

	if props["mime"] != "application/octet-stream" {
		t.Errorf("mime = %v, want application/octet-stream", props["mime"])
	}

	want := fmt.Sprintf("%x", md5.Sum(body))
	if props["filename"] != want {
		t.Errorf("filename = %v, want %s", props["filename"], want)
	}
}

// TestDownloadFileRerunKeepsTheToken checks a page that draws the same file
// again offers it under the same token, so the pack does not change and the
// token the client holds keeps working.
func TestDownloadFileRerunKeepsTheToken(t *testing.T) {
	s := tgframe.NewState()
	defer s.Destroy()

	one := drawDownloadFile(t, s, "Report", []byte("body"))
	two := drawDownloadFile(t, s, "Report", []byte("body"))

	if one["token"] != two["token"] {
		t.Errorf("token rotated on a rerun: %v then %v", one["token"], two["token"])
	}
}
