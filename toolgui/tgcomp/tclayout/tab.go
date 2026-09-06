package tclayout

import (
	"strings"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &tabComponent{}
var tabComponentName = "tab_component"

type tabComponent struct {
	*tgframe.BaseComponent
	Tabs []string `json:"tabs"`
}

func newTabComponent(tabs []string) *tabComponent {
	tabJoinedName := strings.Join(tabs, ",")

	return &tabComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: tabComponentName,
			ID:   tcutil.NormalID(tabComponentName, tabJoinedName),
		},
		Tabs: tabs,
	}
}

// Tab creates a new tab component
func Tab(c *tgframe.Container, tabs []string) []*tgframe.Container {
	comp := newTabComponent(tabs)
	c.AddComponent(comp)

	ret := make([]*tgframe.Container, len(tabs))
	for i, tab := range tabs {
		ret[i] = tgframe.NewContainer(comp.ID+"_"+tab, c.SendNotifyPack)
		c.SendNotifyPack(tgframe.NewNotifyPackCreate(comp.ID, ret[i]))
	}

	return ret
}

// Tab2 create 2 tabs.
func Tab2(c *tgframe.Container, tab1, tab2 string) (*tgframe.Container, *tgframe.Container) {
	retTabs := Tab(c, []string{tab1, tab2})
	return retTabs[0], retTabs[1]
}

// Tab3 create 3 tabs.
func Tab3(c *tgframe.Container, tab1, tab2, tab3 string) (*tgframe.Container, *tgframe.Container, *tgframe.Container) {
	retTabs := Tab(c, []string{tab1, tab2, tab3})
	return retTabs[0], retTabs[1], retTabs[2]
}

// Tab4 create 4 tabs.
func Tab4(c *tgframe.Container, tab1, tab2, tab3, tab4 string) (*tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container) {
	retTabs := Tab(c, []string{tab1, tab2, tab3, tab4})
	return retTabs[0], retTabs[1], retTabs[2], retTabs[3]
}

// Tab5 create 5 tabs.
func Tab5(c *tgframe.Container, tab1, tab2, tab3, tab4, tab5 string) (*tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container) {
	retTabs := Tab(c, []string{tab1, tab2, tab3, tab4, tab5})
	return retTabs[0], retTabs[1], retTabs[2], retTabs[3], retTabs[4]
}
