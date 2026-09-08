package tcmisc

import (
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &progressBarComponent{}
var progressBarComponentName = "progress_bar_component"

type progressBarComponent struct {
	*tgframe.BaseComponent
	Value int    `json:"value"`
	Label string `json:"label"`

	container *tgframe.Container `json:"-"`
}

func newProgressBarComponent(value int, label string, container *tgframe.Container) *progressBarComponent {
	return &progressBarComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: progressBarComponentName,
		},
		Value: value,
		Label: label,

		container: container,
	}
}

// SetValue sets the value of the progress bar. Value should be between 0 and 100.
func (p *progressBarComponent) SetValue(value int) {
	p.Value = value
	p.container.SendNotifyPack(tgframe.NewNotifyPackUpdate(p))
}

// SetLabel sets the label of the progress bar.
func (p *progressBarComponent) SetLabel(label string) {
	p.Label = label
	p.container.SendNotifyPack(tgframe.NewNotifyPackUpdate(p))
}

// Remove takes the progress bar off the screen and gives its id back, so the
// id can be reused in the same run and its state does not leak to whatever
// lands on the id next.
func (p *progressBarComponent) Remove() {
	p.container.RemoveComponent(p)
}

// ProgressBarConf is the configuration for the ProgressBar component.
type ProgressBarConf struct {
	tgframe.Base
}

// ProgressBar creates a new progress bar component.
// Example:
// ```go
// bar := tgcomp.ProgressBar(c, 50, "Progress")
//
//	for i := 0; i <= 100; i++ {
//		bar.SetValue(i)
//		time.Sleep(100 * time.Millisecond)
//	}
//
// bar.SetLabel("Completed")
// ```
func ProgressBar(c *tgframe.Container, value int, label string, conf ...*ProgressBarConf) *progressBarComponent {
	cf := tgframe.OneConf("ProgressBar", conf)

	comp := newProgressBarComponent(value, label, c)
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)
	return comp
}
