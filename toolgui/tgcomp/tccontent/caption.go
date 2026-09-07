package tccontent

import (
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &captionComponent{}
var captionComponentName = "caption_component"

type captionComponent struct {
	*tgframe.BaseComponent
	Text string `json:"text"`
}

func newCaptionComponent(text string) *captionComponent {
	return &captionComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: captionComponentName,
		},
		Text: text,
	}
}

// CaptionConf is the configuration for the Caption component.
type CaptionConf struct {
	tgframe.Base
}

// Caption show a small dimmed text, for a note next to what it explains.
func Caption(c *tgframe.Container, text string, conf ...*CaptionConf) {
	cf := tgframe.OneConf(conf)

	comp := newCaptionComponent(text)
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)
}
