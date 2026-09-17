package tcmisc

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &iframeComponent{}
var iframeComponentName = "iframe_component"

const (
	// defaultIframeWidth is used when IframeConf.Width is empty.
	defaultIframeWidth = "100%"

	// defaultIframeHeight is used when IframeConf.Height is empty.
	// Without it the browser falls back to its own 300x150 default.
	defaultIframeHeight = "150px"
)

type iframeComponent struct {
	*tgframe.BaseComponent
	Html   string `json:"html"`
	Script bool   `json:"script"`
	Width  string `json:"width"`
	Height string `json:"height"`
}

func newIframeComponent(html string, script bool) *iframeComponent {
	return &iframeComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: iframeComponentName,
			ID:   tcutil.HashedID(iframeComponentName, []byte(html)),
		},
		Html:   html,
		Script: script,
		Width:  defaultIframeWidth,
		Height: defaultIframeHeight,
	}
}

// IframeConf is the configuration for the Iframe component.
type IframeConf struct {
	tgframe.Base

	// Script allows the iframe to run javascript. The iframe runs on an
	// opaque origin either way, so it cannot reach the app; it talks to it
	// through window.toolgui.
	Script bool

	// Width is the css width of the iframe, default is "100%".
	Width string

	// Height is the css height of the iframe, default is "150px".
	// "auto" tracks the guest's own height, which needs the guest to call
	// window.toolgui.autoHeight().
	Height string
}

// Iframe shows HTML in a sandboxed iframe.
//
// Read what its guest sends back with [IframeValue].
func Iframe(c *tgframe.Container, html string, conf ...*IframeConf) {
	c.AddComponent(iframeComponentFor(html, tgframe.OneConf("Iframe", conf)))
}

// IframeValue returns the latest value the guest of the iframe html and conf
// describe sent through window.toolgui.update. It reads the value, it does not
// draw the iframe: give it the same html and conf the [Iframe] call gets.
//
// It returns nil when the guest has no value for the page, so a page tells
// that apart from a value that is the zero T. A guest that has not sent yet and
// one that sent null both read as nil: null is how a guest says it has nothing,
// not a value of its own.
//
//	if v := tcmisc.IframeValue[picked](c, html, conf); v != nil {
//		use(*v)
//	}
//
// The frontend keys the value by the iframe's own component id, so a guest can
// only write to its own state.
func IframeValue[T any](c *tgframe.Container, html string, conf ...*IframeConf) *T {
	comp := iframeComponentFor(html, tgframe.OneConf("IframeValue", conf))

	return frameValue[T](c, "iframe", comp.ID)
}

// iframeComponentFor builds the component html and conf describe. Both entry
// points go through it, so the id IframeValue reads is the one Iframe draws.
func iframeComponentFor(html string, cf *IframeConf) *iframeComponent {
	comp := newIframeComponent(html, cf.Script)

	if cf.Width != "" {
		comp.Width = cf.Width
	}

	if cf.Height != "" {
		comp.Height = cf.Height
	}

	tgframe.SetConfID(comp, cf)

	return comp
}
