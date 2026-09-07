package tccontent

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &badgeComponent{}
var badgeComponentName = "badge_component"

type badgeComponent struct {
	*tgframe.BaseComponent
	Text  string       `json:"text"`
	Color tcutil.Color `json:"color"`
}

func newBadgeComponent(text string) *badgeComponent {
	return &badgeComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: badgeComponentName,
		},
		Text: text,
	}
}

// BadgeConf is the configuration for the Badge component.
type BadgeConf struct {
	tgframe.Base

	// Color is the color of the badge. Default is tcutil.ColorNull, which
	// leaves it neutral.
	Color tcutil.Color
}

// Badge show a short label, for a status or a tag next to other content.
func Badge(c *tgframe.Container, text string, conf ...*BadgeConf) {
	cf := tgframe.OneConf("Badge", conf)

	comp := newBadgeComponent(text)
	comp.Color = cf.Color
	tgframe.SetConfID(comp, cf)
	c.AddComponent(comp)
}
