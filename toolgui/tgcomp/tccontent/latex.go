package tccontent

import (
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
		},
		Latex: text,
	}
}

// LatexConf is the configuration for the Latex component.
type LatexConf struct {
	tgframe.Base
}

// Latex render text as latex.
func Latex(c *tgframe.Container, text string, conf ...*LatexConf) {
	cf := tgframe.OneConf("Latex", conf)

	comp := newLatexComponent(text)
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)
}
