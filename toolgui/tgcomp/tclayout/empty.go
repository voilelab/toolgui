package tclayout

import (
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &emptyComponent{}
var emptyComponentName = "empty_component"

type emptyComponent struct {
	*tgframe.BaseComponent
}

func newEmptyComponent() *emptyComponent {
	return &emptyComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: emptyComponentName,
		},
	}
}

// EmptyConf is the configuration for the Empty component. The container an
// empty hands out derives its id from the empty's; give none and it carries
// none, and the components inside are still placed by position.
type EmptyConf struct {
	tgframe.Base
}

// EmptyContainer is what [Empty] hands out: a place in the page that can be
// written and written over.
type EmptyContainer = tgframe.Slot

// Empty reserves a place in the page and hands back a slot to write it with.
// Writing the slot again takes the previous contents off the screen instead of
// adding to them, so a page function can show progress and then replace it
// with the result:
//
//	slot := tgcomp.Empty(c)
//	slot.With(func(c *tgframe.Container) {
//		tgcomp.Text(c, "Querying…")
//	})
//
//	head, rows := query()
//
//	slot.With(func(c *tgframe.Container) {
//		tgcomp.Table(c, head, rows)
//	})
//
// The slot starts empty on every run, whatever the last run left in it.
func Empty(c *tgframe.Container, conf ...*EmptyConf) *EmptyContainer {
	cf := tgframe.OneConf("Empty", conf)

	comp := newEmptyComponent()
	tgframe.SetConfID(comp, cf)

	emptyComp := c.AddComponent(comp)
	return c.AddSlotTo(emptyComp, "inner", 0)
}
