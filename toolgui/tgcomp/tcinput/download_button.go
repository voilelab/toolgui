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
	tgframe.Base

	// MIME specifies the Multipurpose Internet Mail Extension (MIME) type of the downloaded content.
	// Defaults to "application/octet-stream" if not provided.
	MIME string

	// Color defines the color of the download button.
	Color tcutil.Color

	// Disabled indicates whether the download button should be initially disabled.
	Disabled bool

	// Filename sets the suggested filename for the downloaded content when clicked.
	Filename string
}

// DownloadButton create a download button component.
func DownloadButton(c *tgframe.Container, text string, body []byte, conf ...*DownloadButtonConf) bool {
	cf := tgframe.OneConf("DownloadButton", conf)

	b64Body := base64.RawStdEncoding.EncodeToString([]byte(body))

	mime := "application/octet-stream"
	if cf.MIME != "" {
		mime = cf.MIME
	}

	uri := fmt.Sprintf("data:%s;base64,%s", mime, b64Body)
	comp := newDownloadButtonComponent(text, uri)

	if cf.Filename != "" {
		comp.Filename = cf.Filename
	}

	comp.Color = cf.Color
	comp.Disabled = cf.Disabled
	tgframe.SetConfID(comp, cf)

	c.AddComponent(comp)
	return c.State.GetClickID() == comp.ID
}
