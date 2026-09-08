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
}

func newProgressBarComponent(value int, label string) *progressBarComponent {
	return &progressBarComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: progressBarComponentName,
		},
		Value: value,
		Label: label,
	}
}

// ProgressBarHandle is what [ProgressBar] hands out: the bar it drew, to be
// moved along while the page function works. It is a value like any other, so
// the work may take it as a parameter or keep it in a struct field:
//
//	type importer struct{ bar *tgcomp.ProgressBarHandle }
type ProgressBarHandle struct {
	comp      *progressBarComponent
	container *tgframe.Container
}

// SetValue sets the value of the progress bar. Value should be between 0 and 100.
func (p *ProgressBarHandle) SetValue(value int) {
	p.comp.Value = value
	p.container.SendNotifyPack(tgframe.NewNotifyPackUpdate(p.comp))
}

// SetLabel sets the label of the progress bar.
func (p *ProgressBarHandle) SetLabel(label string) {
	p.comp.Label = label
	p.container.SendNotifyPack(tgframe.NewNotifyPackUpdate(p.comp))
}

// Remove takes the progress bar off the screen and gives its id back, so the
// id can be reused in the same run and its state does not leak to whatever
// lands on the id next.
func (p *ProgressBarHandle) Remove() {
	p.container.RemoveComponent(p.comp)
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
func ProgressBar(c *tgframe.Container, value int, label string, conf ...*ProgressBarConf) *ProgressBarHandle {
	cf := tgframe.OneConf("ProgressBar", conf)

	comp := newProgressBarComponent(value, label)
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)
	return &ProgressBarHandle{comp: comp, container: c}
}
