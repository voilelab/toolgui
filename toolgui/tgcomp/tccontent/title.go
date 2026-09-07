package tccontent

import (
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &titleComponent{}
var titleComponentName = "title_component"

type titleComponent struct {
	*tgframe.BaseComponent
	Text string `json:"text"`
}

func newTitleComponent(text string) *titleComponent {
	return &titleComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: titleComponentName,
		},
		Text: text,
	}
}

// TitleConf is the configuration for the Title component.
type TitleConf struct {
	tgframe.Base
}

// Title show a title.
func Title(c *tgframe.Container, text string, conf ...*TitleConf) {
	cf := tgframe.OneConf("Title", conf)

	comp := newTitleComponent(text)
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)
}
