package tccontent

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
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
			ID:   tcutil.HashedID(markdownComponentName, []byte(text)),
		},
		Markdown: text,
	}
}

// Markdown render markdown to html.
func Markdown(c *tgframe.Container, markdown string) {
	comp := newMarkdownComponent(markdown)
	c.AddComponent(comp)
}

// Markdown create a markdown-rendering part with a user-specific id.
func MarkdownWithID(c *tgframe.Container, markdown string, id string) {
	comp := newMarkdownComponent(markdown)
	comp.SetID(id)
	c.AddComponent(comp)
}
