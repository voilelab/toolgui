package tccontent

import (
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &subtitleComponent{}
var subtitleComponentName = "subtitle_component"

type subtitleComponent struct {
	*tgframe.BaseComponent
	Text string `json:"text"`
}

func newSubtitleComponent(text string) *subtitleComponent {
	return &subtitleComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: subtitleComponentName,
		},
		Text: text,
	}
}

// SubtitleConf is the configuration for the Subtitle component.
//
//tgcomp:export
type SubtitleConf struct {
	tgframe.Base
}

// Subtitle create a subtitle.
//
//tgcomp:export
func Subtitle(c *tgframe.Container, text string, conf ...*SubtitleConf) {
	cf := tgframe.OneConf("Subtitle", conf)

	comp := newSubtitleComponent(text)
	tgframe.SetConfIDIn(c, comp, cf)
	c.AddComponent(comp)
}
