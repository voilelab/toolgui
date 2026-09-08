package tcmisc

import (
	"sync"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tclayout"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &spinnerComponent{}
var spinnerComponentName = "spinner_component"

type spinnerComponent struct {
	*tgframe.BaseComponent
	Label string `json:"label"`
}

func newSpinnerComponent(label string) *spinnerComponent {
	return &spinnerComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: spinnerComponentName,
		},
		Label: label,
	}
}

// SpinnerConf is the configuration for the Spinner component.
type SpinnerConf struct {
	tgframe.Base
}

// Spinner shows that the page function is busy, and returns the function that
// takes it down again. It sits in an [tclayout.Empty] slot, so taking it down
// leaves the page as if it had never been there.
//
// Call it with defer and the spinner is taken down however the work ends,
// including a panic:
//
//	defer tgcomp.Spinner(c, "Loading…")()
//
// The returned function may be called more than once; only the first call
// does anything.
func Spinner(c *tgframe.Container, label string, conf ...*SpinnerConf) func() {
	cf := tgframe.OneConf("Spinner", conf)

	slot := tclayout.Empty(c, &tclayout.EmptyConf{ID: cf.ID})
	slot.With(func(c *tgframe.Container) {
		c.AddComponent(newSpinnerComponent(label))
	})

	return sync.OnceFunc(slot.Clear)
}
