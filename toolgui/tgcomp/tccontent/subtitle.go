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

// Subtitle create a subtitle.
func Subtitle(c *tgframe.Container, text string) {
	c.AddComponent(newSubtitleComponent(text))
}

// SubtitleWithID create a subtitle component with a user specific id.
func SubtitleWithID(c *tgframe.Container, text string, id string) {
	comp := newSubtitleComponent(text)
	comp.SetID(id)
	c.AddComponent(comp)
}
