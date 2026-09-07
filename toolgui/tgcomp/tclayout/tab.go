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
	return &tabComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: tabComponentName,
			// The client keeps which tab is open, so the component needs a
			// name of its own to keep that across runs.
			ID: tcutil.NormalID(tabComponentName, strings.Join(tabs, ",")),
		},
		Tabs: tabs,
	}
}

// TabConf is the configuration for the Tab components.
type TabConf struct {
	tgframe.Base
}

// Tab creates a new tab component
func Tab(c *tgframe.Container, tabs []string, conf ...*TabConf) []*tgframe.Container {
	cf := tgframe.OneConf("Tab", conf)

	tabComp := newTabComponent(tabs)
	tgframe.SetConfID(tabComp, cf)

	comp := c.AddComponent(tabComp)

	ret := make([]*tgframe.Container, len(tabs))
	for i, tab := range tabs {
		ret[i] = c.AddContainerTo(comp, tab, i)
	}

	return ret
}

// Tab2 create 2 tabs.
func Tab2(c *tgframe.Container, tab1, tab2 string, conf ...*TabConf) (
	*tgframe.Container, *tgframe.Container) {

	retTabs := Tab(c, []string{tab1, tab2}, conf...)
	return retTabs[0], retTabs[1]
}

// Tab3 create 3 tabs.
func Tab3(c *tgframe.Container, tab1, tab2, tab3 string, conf ...*TabConf) (
	*tgframe.Container, *tgframe.Container, *tgframe.Container) {

	retTabs := Tab(c, []string{tab1, tab2, tab3}, conf...)
	return retTabs[0], retTabs[1], retTabs[2]
}

// Tab4 create 4 tabs.
func Tab4(c *tgframe.Container, tab1, tab2, tab3, tab4 string, conf ...*TabConf) (
	*tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container) {

	retTabs := Tab(c, []string{tab1, tab2, tab3, tab4}, conf...)
	return retTabs[0], retTabs[1], retTabs[2], retTabs[3]
}

// Tab5 create 5 tabs.
func Tab5(c *tgframe.Container, tab1, tab2, tab3, tab4, tab5 string, conf ...*TabConf) (
	*tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container,
	*tgframe.Container) {

	retTabs := Tab(c, []string{tab1, tab2, tab3, tab4, tab5}, conf...)
	return retTabs[0], retTabs[1], retTabs[2], retTabs[3], retTabs[4]
}
