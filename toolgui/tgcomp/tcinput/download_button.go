package tcinput

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &downloadButtonComponent{}
var downloadButtonComponentName = "download_button_component"

type downloadButtonComponent struct {
	*tgframe.BaseComponent
	Text     string       `json:"text"`
	URI      string       `json:"uri"`
	Filename string       `json:"filename"`
	Color    tcutil.Color `json:"color"`
	Disabled bool         `json:"disabled"`
}

func newDownloadButtonComponent(text, uri string) *downloadButtonComponent {
	return &downloadButtonComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: downloadButtonComponentName,
			ID:   tcutil.NormalID(downloadButtonComponentName, text),
		},
		Text:     text,
		URI:      uri,
		Filename: fmt.Sprintf("%x", md5.Sum([]byte(uri))),
	}
}

type DownloadButtonConf struct {
	// MIME specifies the Multipurpose Internet Mail Extension (MIME) type of the downloaded content.
	// Defaults to "application/octet-stream" if not provided.
	MIME string

	// Color defines the color of the download button.
	Color tcutil.Color

	// Disabled indicates whether the download button should be initially disabled.
	Disabled bool

	// Filename sets the suggested filename for the downloaded content when clicked.
	Filename string

	ID string
}

// DownloadButton create a download button component.
func DownloadButton(s *tgframe.State, c *tgframe.Container, text string, body []byte) bool {
	return DownloadButtonWithConf(s, c, text, body, nil)
}

// DownloadButtonWithConf create a download button component with a user specific config.
// If no configuration is provided, a new one with default values is used.
func DownloadButtonWithConf(s *tgframe.State, c *tgframe.Container, text string, body []byte, conf *DownloadButtonConf) bool {
	if conf == nil {
		conf = &DownloadButtonConf{}
	}

	b64Body := base64.RawStdEncoding.EncodeToString([]byte(body))

	mime := "application/octet-stream"
	if conf.MIME != "" {
		mime = conf.MIME
	}

	uri := fmt.Sprintf("data:%s;base64,%s", mime, b64Body)
	comp := newDownloadButtonComponent(text, uri)

	if conf.Filename != "" {
		comp.Filename = conf.Filename
	}

	comp.Color = conf.Color
	comp.Disabled = conf.Disabled

	if conf.ID != "" {
		comp.SetID(conf.ID)
	}

	c.AddComponent(comp)
	return s.GetClickID() == comp.ID
}
