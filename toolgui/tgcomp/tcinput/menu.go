package tcinput

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &menuComponent{}
var menuComponentName = "menu_component"

// menuItem is one entry of the dropdown. It carries an id of its own because
// the click that comes back names the item, not the menu.
type menuItem struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type menuComponent struct {
	*tgframe.BaseComponent
	Label    string       `json:"label"`
	Items    []menuItem   `json:"items"`
	Color    tcutil.Color `json:"color"`
	Disabled bool         `json:"disabled"`
}

func newMenuComponent(label string) *menuComponent {
	return &menuComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: menuComponentName,
			ID:   tcutil.NormalID(menuComponentName, label),
		},
		Label: label,
	}
}

// setItems builds the item ids off the component's own id, so it has to run
// after [tgframe.SetConfID] has settled what that id is.
func (c *menuComponent) setItems(items []string) {
	c.Items = make([]menuItem, len(items))
	for i, item := range items {
		c.Items[i] = menuItem{ID: menuItemID(c.ID, i), Label: item}
	}
}

// menuItemID is the click id of the item at index under a menu's id.
func menuItemID(menuID string, index int) string {
	return fmt.Sprintf("%s_%d", menuID, index)
}

// MenuConf is the configuration for the Menu component.
type MenuConf struct {
	tgframe.Base

	// Color defines the color of the button that opens the menu.
	Color tcutil.Color

	// Disabled is true if the button that opens the menu is disabled.
	Disabled bool
}

// Menu create a button with a list of actions behind it, and return the index
// of the item clicked. 0-indexed, return nil if no item is clicked.
//
// The click goes through the same event a [Button] press does, so the index
// is there for exactly the run that handles the click and nil again on the
// next one. Which item it was is the item's own id, derived from the menu's,
// so two menus sharing a label need a Conf.ID between them the way two
// buttons do.
//
// Whether the dropdown is open is the client's, like a popover's: the items
// are written every run, open or not.
func Menu(c *tgframe.Container, label string, items []string,
	conf ...*MenuConf) *int {

	cf := tgframe.OneConf("Menu", conf)

	comp := newMenuComponent(label)
	comp.Color = cf.Color
	comp.Disabled = cf.Disabled
	tgframe.SetConfID(comp, cf)
	comp.setItems(items)

	c.AddComponent(comp)

	clickID := c.State.GetClickID()
	for i, item := range comp.Items {
		if item.ID == clickID {
			idx := i
			return &idx
		}
	}

	return nil
}
