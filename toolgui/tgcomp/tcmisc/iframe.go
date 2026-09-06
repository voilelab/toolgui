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
	// Script allows the iframe to run javascript.
	//
	// Warning: an iframe with Script enabled is NOT isolated from the app.
	// Only pass html you trust as much as your own app code.
	Script bool

	// Width is the css width of the iframe, default is "100%".
	Width string

	// Height is the css height of the iframe, default is "150px".
	// "auto" tracks the guest's own height, which needs the guest to call
	// window.toolgui.autoHeight().
	Height string

	// ID is the unique identifier for this iframe component.
	ID string
}

// Iframe show a html.
// script is used to allow the iframe to run javascript. (notice that this is not secure)
func Iframe(c *tgframe.Container, html string, script bool) {
	IframeWithConf(c, html, &IframeConf{Script: script})
}

// IframeWithID create a html component with a user specific id.
// script is used to allow the iframe to run javascript. (notice that this is not secure)
func IframeWithID(c *tgframe.Container, html string, script bool, id string) {
	IframeWithConf(c, html, &IframeConf{Script: script, ID: id})
}

// IframeWithConf show a html with a custom configuration.
func IframeWithConf(c *tgframe.Container, html string, conf *IframeConf) {
	if conf == nil {
		conf = &IframeConf{}
	}

	comp := newIframeComponent(html, conf.Script)

	if conf.Width != "" {
		comp.Width = conf.Width
	}

	if conf.Height != "" {
		comp.Height = conf.Height
	}

	if conf.ID != "" {
		comp.SetID(conf.ID)
	}

	c.AddComponent(comp)
}

// IframeValue unmarshals the latest value the iframe with the given id sent
// through window.update into out. The id is the one passed to IframeWithID or
// IframeConf.ID.
//
// The frontend keys the value by the iframe's own component id, so an iframe
// can only write to its own state.
func IframeValue(s *tgframe.State, id string, out any) error {
	return s.GetObject(tcutil.NormalID(iframeComponentName, id), out)
}
