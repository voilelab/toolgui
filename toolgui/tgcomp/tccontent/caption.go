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
//
//tgcomp:export
type CaptionConf struct {
	tgframe.Base
}

// Caption show a small dimmed text, for a note next to what it explains.
//
//tgcomp:export
func Caption(c *tgframe.Container, text string, conf ...*CaptionConf) {
	cf := tgframe.OneConf("Caption", conf)

	comp := newCaptionComponent(text)
	tgframe.SetConfIDIn(c, comp, cf)
	c.AddComponent(comp)
}
