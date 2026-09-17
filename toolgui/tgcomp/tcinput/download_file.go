package tcinput

import (
	"crypto/md5"
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgutil"
)

var _ tgframe.Component = &downloadFileComponent{}
var downloadFileComponentName = "download_file_component"

type downloadFileComponent struct {
	*tgframe.BaseComponent
	Text string `json:"text"`

	// Token is what the client fetches the file by. The content is not here:
	// that is the whole difference from [DownloadButton].
	Token    string       `json:"token"`
	Filename string       `json:"filename"`
	MIME     string       `json:"mime"`
	Color    tcutil.Color `json:"color"`
	Disabled bool         `json:"disabled"`
}

func newDownloadFileComponent(text string) *downloadFileComponent {
	return &downloadFileComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: downloadFileComponentName,
			ID:   tcutil.NormalID(downloadFileComponentName, text),
		},
		Text: text,
	}
}

// DownloadFileConf is the configuration for the DownloadFile component.
type DownloadFileConf struct {
	tgframe.Base

	// MIME specifies the Multipurpose Internet Mail Extension (MIME) type of
	// the downloaded content. Defaults to "application/octet-stream" if not
	// provided.
	MIME string

	// Color defines the color of the download button.
	Color tcutil.Color

	// Disabled indicates whether the download button should be initially
	// disabled.
	Disabled bool

	// Filename sets the suggested filename for the downloaded content when
	// clicked. Defaults to the content's MD5 in hex.
	Filename string
}

// DownloadFile create a button that hands the app user a file to download.
//
// The file is stored where the build keeps files -- on disk on a server, in
// the origin private file system in a browser build -- and the component's
// pack carries only an unguessable token to fetch it by. A click fetches the
// bytes through the state's own transport and saves them, so a big output
// costs the transport what it weighs rather than a third again as base64, and
// the link the browser follows is one the page built rather than a URI Go
// sent.
//
// [DownloadButton] is the one to reach for below a few hundred kilobytes: it
// puts the content in the pack as a data: URI, which is a round trip saved on
// a file the pack can carry comfortably. This one is for everything above
// that.
//
// It reports whether this run is handling a click on it, the same as
// [DownloadButton].
func DownloadFile(c *tgframe.Container, text string, body []byte,
	conf ...*DownloadFileConf) bool {
	cf := tgframe.OneConf("DownloadFile", conf)

	comp := downloadFileComponentFor(text, cf)

	comp.Filename = fmt.Sprintf("%x", md5.Sum(body))
	if cf.Filename != "" {
		comp.Filename = cf.Filename
	}

	if c.State == nil {
		c.Fail(tgutil.NewError("DownloadFile needs a state to keep the file in"))
		return false
	}

	// Under the component's id, so a rerun replaces the file this button
	// offers rather than piling up another one beside it.
	download, err := c.State.SetDownload(comp.ID, comp.Filename, comp.MIME, body)
	if err != nil {
		c.Fail(tgutil.Errorf("store the file to download: %w", err))
		return false
	}

	comp.Token = download.Token()

	c.AddComponent(comp)
	return c.State.GetClickID() == comp.ID
}

// DownloadFileClicked reports whether the click this run is handling is the
// one on the download file button text and conf describe. It reads the click,
// it does not draw the button: give it the same text and conf the
// [DownloadFile] call gets. It's [ButtonClicked] for a download file button.
//
// It takes no body, for the same reason [DownloadButtonClicked] does not: the
// button's id comes from text (or the conf id), not from the file.
func DownloadFileClicked(c *tgframe.Container, text string,
	conf ...*DownloadFileConf) bool {

	comp := downloadFileComponentFor(text,
		tgframe.OneConf("DownloadFileClicked", conf))

	return clicked(c, comp.ID)
}

// downloadFileComponentFor builds the component text and conf describe, all
// but the file itself. Both entry points go through it, so the id
// DownloadFileClicked reads is the one DownloadFile draws.
func downloadFileComponentFor(text string,
	cf *DownloadFileConf) *downloadFileComponent {

	comp := newDownloadFileComponent(text)
	tgframe.SetConfID(comp, cf)

	comp.MIME = "application/octet-stream"
	if cf.MIME != "" {
		comp.MIME = cf.MIME
	}

	comp.Color = cf.Color
	comp.Disabled = cf.Disabled

	return comp
}
