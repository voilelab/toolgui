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

func newDownloadButtonComponent(text string) *downloadButtonComponent {
	return &downloadButtonComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: downloadButtonComponentName,
			ID:   tcutil.NormalID(downloadButtonComponentName, text),
		},
		Text: text,
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

	comp := downloadButtonComponentFor(text, cf)
	comp.URI = uri

	comp.Filename = fmt.Sprintf("%x", md5.Sum([]byte(uri)))
	if cf.Filename != "" {
		comp.Filename = cf.Filename
	}

	c.AddComponent(comp)
	return c.State.GetClickID() == comp.ID
}

// DownloadButtonClicked reports whether the click this run is handling is the
// one on the download button text and conf describe. It reads the click, it
// does not draw the button: give it the same text and conf the
// [DownloadButton] call gets. It's [ButtonClicked] for a download button.
//
// It takes no body: the button's id comes from text (or the conf id), not from
// what it hands over, so a page can ask about the click before it has the
// bytes to offer.
func DownloadButtonClicked(c *tgframe.Container, text string,
	conf ...*DownloadButtonConf) bool {

	comp := downloadButtonComponentFor(text,
		tgframe.OneConf("DownloadButtonClicked", conf))

	return clicked(c, comp.ID)
}

// downloadButtonComponentFor builds the component text and conf describe, all
// but the body. Both entry points go through it, so the id
// DownloadButtonClicked reads is the one DownloadButton draws.
func downloadButtonComponentFor(text string,
	cf *DownloadButtonConf) *downloadButtonComponent {

	comp := newDownloadButtonComponent(text)
	comp.Color = cf.Color
	comp.Disabled = cf.Disabled
	tgframe.SetConfID(comp, cf)

	return comp
}
