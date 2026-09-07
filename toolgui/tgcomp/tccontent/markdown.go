package tccontent

import (
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &markdownComponent{}
var markdownComponentName = "markdown_component"

type markdownComponent struct {
	*tgframe.BaseComponent
	Markdown string `json:"text"`
}

func newMarkdownComponent(text string) *markdownComponent {
	return &markdownComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: markdownComponentName,
		},
		Markdown: text,
	}
}

// MarkdownConf is the configuration for the Markdown component.
type MarkdownConf struct {
	tgframe.Base
}

// Markdown render markdown to html.
func Markdown(c *tgframe.Container, markdown string, conf ...*MarkdownConf) {
	cf := tgframe.OneConf("Markdown", conf)

	comp := newMarkdownComponent(markdown)
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)
}
