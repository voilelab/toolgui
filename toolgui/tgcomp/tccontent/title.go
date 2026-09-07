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
	// ID is the unique identifier for this title component.
	ID string
}

// Title show a title. conf may be nil.
func Title(c *tgframe.Container, text string, conf *TitleConf) {
	comp := newTitleComponent(text)
	if conf != nil && conf.ID != "" {
		comp.SetID(conf.ID)
	}

	c.AddComponent(comp)
}
