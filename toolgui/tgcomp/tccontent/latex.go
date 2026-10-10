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
//
//tgcomp:export
type LatexConf struct {
	tgframe.Base
}

// Latex renders text as LaTeX.
//
//tgcomp:export
func Latex(c *tgframe.Container, text string, conf ...*LatexConf) {
	cf := tgframe.OneConf("Latex", conf)

	comp := newLatexComponent(text)
	tgframe.SetConfIDIn(c, comp, cf)
	c.AddComponent(comp)
}
