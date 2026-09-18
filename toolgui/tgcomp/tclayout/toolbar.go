package tclayout

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &toolbarComponent{}

const toolbarComponentName = "toolbar_component"

// The ways a toolbar may line its items up, as the client receives them.
const (
	ToolbarJustifyStart   = "start"
	ToolbarJustifyEnd     = "end"
	ToolbarJustifyBetween = "between"
)

type toolbarComponent struct {
	*tgframe.BaseComponent

	Sticky  bool   `json:"sticky"`
	Justify string `json:"justify"`
}

func newToolbarComponent() *toolbarComponent {
	return &toolbarComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: toolbarComponentName,
		},
	}
}

// toolbarJustify settles what an unset conf means, and rejects one that names
// nothing the client can render.
func toolbarJustify(justify string) string {
	switch justify {
	case "":
		return ToolbarJustifyStart
	case ToolbarJustifyStart, ToolbarJustifyEnd, ToolbarJustifyBetween:
		return justify
	}

	panic(fmt.Sprintf("toolgui: unsupported toolbar justify: %q", justify))
}

// ToolbarConf is the configuration for the Toolbar component. The container a
// toolbar hands out derives its id from the toolbar's; give none and it
// carries none, and the components inside are still placed by position.
type ToolbarConf struct {
	tgframe.Base

	// Sticky keeps the row at the top of the page while the rest of it
	// scrolls under, on an opaque background so nothing shows through.
	Sticky bool

	// Justify is ToolbarJustifyStart (default), ToolbarJustifyEnd or
	// ToolbarJustifyBetween. Anything else panics.
	Justify string
}

// Toolbar creates a row of controls and returns the container they go into.
// What is written inside lines up horizontally, each item as wide as it needs
// to be, rather than taking a row of the page each:
//
//	bar := tgcomp.Toolbar(p.Main, &tgcomp.ToolbarConf{Sticky: true})
//	if tgcomp.Button(bar, "Run") {
//		run()
//	}
//	tgcomp.Select(bar, "Mode", modes)
//
// A row too wide for the viewport wraps instead of pushing the page sideways.
func Toolbar(c *tgframe.Container, conf ...*ToolbarConf) *tgframe.Container {
	cf := tgframe.OneConf("Toolbar", conf)

	comp := newToolbarComponent()
	comp.Sticky = cf.Sticky
	comp.Justify = toolbarJustify(cf.Justify)
	tgframe.SetConfID(comp, cf)

	toolbarComp := c.AddComponent(comp)
	return c.AddContainerTo(toolbarComp, "inner", 0)
}
