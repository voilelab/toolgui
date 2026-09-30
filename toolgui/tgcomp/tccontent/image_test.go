package tccontent

import (
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func TestDetectImageMIME(t *testing.T) {
	tests := []struct {
		name string
		bs   []byte
		want string
	}{
		{"png", []byte("\x89PNG\x0D\x0A\x1A\x0A"), "image/png"},
		{"jpeg", []byte("\xFF\xD8\xFF\xE0"), "image/jpeg"},
		{"gif", []byte("GIF89a"), "image/gif"},
		{"webp", []byte("RIFF\x00\x00\x00\x00WEBPVP8 "), "image/webp"},
		{"unknown", []byte("hello"), "image/png"},
		{"empty", nil, "image/png"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectImageMIME(tt.bs, "image/png"); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestImageBytesDataURI(t *testing.T) {
	comp := addMetric(t, func(c *tgframe.Container) {
		Image(c, []byte("GIF89a"))
	})
	if got, want := comp["src"], "data:image/gif;base64,R0lGODlh"; got != want {
		t.Errorf("src = %v, want %v", got, want)
	}
}
