package tccontent

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &latexComponent{}
var latexComponentName = "latex_component"

type latexComponent struct {
	*tgframe.BaseComponent
	Latex string `json:"latex"`
}

func newLatexComponent(text string) *latexComponent {
	return &latexComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: latexComponentName,
			ID:   tcutil.NormalID(latexComponentName, text),
		},
		Latex: text,
	}
}

// Text show a text.
func Latex(c *tgframe.Container, text string) {
	comp := newLatexComponent(text)
	c.AddComponent(comp)
}

// TextWithID create a text component with a user specific id.
func LatexWithID(c *tgframe.Container, text string, id string) {
	comp := newLatexComponent(text)
	comp.SetID(id)
	c.AddComponent(comp)
}
