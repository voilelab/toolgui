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
func Iframe(c *tgframe.Container, html string, conf ...*IframeConf) {
	cf := tgframe.OneConf("Iframe", conf)

	comp := newIframeComponent(html, cf.Script)

	if cf.Width != "" {
		comp.Width = cf.Width
	}

	if cf.Height != "" {
		comp.Height = cf.Height
	}

	tgframe.SetConfID(comp, cf)

	c.AddComponent(comp)
}

// IframeValue unmarshals the latest value the iframe with the given id sent
// through window.update into out. The id is the one passed as
// [IframeConf.ID].
//
// The frontend keys the value by the iframe's own component id, so an iframe
// can only write to its own state.
func IframeValue(s *tgframe.State, id string, out any) error {
	return s.GetObject(tcutil.NormalID(iframeComponentName, id), out)
}
