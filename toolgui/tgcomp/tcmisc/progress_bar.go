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

	SendNotifyPack tgframe.SendNotifyPackFunc `json:"-"`
}

func newProgressBarComponent(value int, label string, sendNotifyPack tgframe.SendNotifyPackFunc) *progressBarComponent {
	return &progressBarComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: progressBarComponentName,
		},
		Value: value,
		Label: label,

		SendNotifyPack: sendNotifyPack,
	}
}

// SetValue sets the value of the progress bar. Value should be between 0 and 100.
func (p *progressBarComponent) SetValue(value int) {
	p.Value = value
	p.SendNotifyPack(tgframe.NewNotifyPackUpdate(p))
}

// SetLabel sets the label of the progress bar.
func (p *progressBarComponent) SetLabel(label string) {
	p.Label = label
	p.SendNotifyPack(tgframe.NewNotifyPackUpdate(p))
}

// Remove removes the progress bar component.
func (p *progressBarComponent) Remove() {
	p.SendNotifyPack(tgframe.NewNotifyPackDelete(p))
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
func ProgressBar(c *tgframe.Container, value int, label string) *progressBarComponent {
	comp := newProgressBarComponent(value, label, c.SendNotifyPack)
	c.AddComponent(comp)
	return comp
}

// ProgressBarWithID creates a progress bar with a user specific id.
func ProgressBarWithID(c *tgframe.Container, value int, label string, id string) *progressBarComponent {
	comp := newProgressBarComponent(value, label, c.SendNotifyPack)
	comp.SetID(id)
	c.AddComponent(comp)
	return comp
}
