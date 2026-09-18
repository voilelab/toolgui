package tgframe

import "github.com/voilelab/toolgui/toolgui/tgutil"

// MenuIDPrefix is what every menu item's click id carries in front of the id
// the app declares. Menu items live on the app rather than inside a run, yet
// they report through the same click id a component does; the prefix is what
// keeps the two apart. No component can produce it: a component's id is
// "<component name>_<label>", and every component name ends in "_component".
const MenuIDPrefix = "menu_item_"

// MenuID returns the click id the item declared under id reports through.
// [MenuClicked] is the usual way to read it; this is for a page that wants to
// compare against [State.GetClickID] itself.
func MenuID(id string) string {
	return MenuIDPrefix + id
}

// MenuNodeType is the kind of a [MenuNode].
type MenuNodeType string

const (
	// MenuNodeText is an item that reports a click.
	MenuNodeText MenuNodeType = "text"

	// MenuNodeSeparator is a line between items.
	MenuNodeSeparator MenuNodeType = "separator"

	// MenuNodeSubmenu is an item holding items of its own.
	MenuNodeSubmenu MenuNodeType = "submenu"
)

// MenuNode is one node of an app's menu tree, as the frontend receives it in
// [AppConf].
type MenuNode struct {
	Type MenuNodeType `json:"type"`

	// Label is what the item reads, on a text item and a submenu.
	Label string `json:"label,omitzero"`

	// ID is the click id of a text item, already carrying [MenuIDPrefix].
	ID string `json:"id,omitzero"`

	// Children are a submenu's own items.
	Children []*MenuNode `json:"children,omitzero"`
}

// Menu is a menu tree under construction, or one submenu of one. The tree is
// static: it is declared once on the app, not rebuilt per run, so the same
// declaration holds for a native menu that has no diff to apply.
//
//	app.SetMenu(tgframe.NewMenu().
//		Submenu("File", func(m *tgframe.Menu) {
//			m.Text("Open", "file_open")
//			m.Separator()
//			m.Text("Quit", "file_quit")
//		}))
type Menu struct {
	nodes []*MenuNode
}

// NewMenu returns an empty menu.
func NewMenu() *Menu {
	return &Menu{}
}

// Text adds an item reading label that reports a click under id. Read the
// click with [MenuClicked].
func (m *Menu) Text(label, id string) *Menu {
	m.nodes = append(m.nodes, &MenuNode{
		Type:  MenuNodeText,
		Label: label,
		ID:    MenuID(id),
	})

	return m
}

// Separator adds a line between the items around it.
func (m *Menu) Separator() *Menu {
	m.nodes = append(m.nodes, &MenuNode{Type: MenuNodeSeparator})
	return m
}

// Submenu adds an item reading label holding whatever build writes into it.
// Submenus nest.
func (m *Menu) Submenu(label string, build func(*Menu)) *Menu {
	sub := NewMenu()
	if build != nil {
		build(sub)
	}

	m.nodes = append(m.nodes, &MenuNode{
		Type:     MenuNodeSubmenu,
		Label:    label,
		Children: sub.nodes,
	})

	return m
}

// Nodes returns the tree the menu has been built into.
func (m *Menu) Nodes() []*MenuNode {
	return m.nodes
}

// cloneNodes returns a deep copy of nodes. [App.SetMenu] keeps one rather than
// the Menu it was handed: a caller still holding that Menu could otherwise add
// an item to it afterwards, and the tree would grow while the set of declared
// ids did not -- leaving an item the menubar draws whose clicks [MenuClicked]
// turns away as undeclared.
func cloneNodes(nodes []*MenuNode) []*MenuNode {
	if nodes == nil {
		return nil
	}

	out := make([]*MenuNode, len(nodes))
	for i, node := range nodes {
		cp := *node
		cp.Children = cloneNodes(node.Children)
		out[i] = &cp
	}

	return out
}

// ErrMenuItem is what a menu the app cannot serve is reported with:
// an item with no label, a text item with no id, or two items sharing one.
var ErrMenuItem = tgutil.NewError("invalid menu item")

// collectIDs walks nodes, checking each one and adding every text item's click
// id to ids.
func collectIDs(nodes []*MenuNode, ids map[string]bool) error {
	for _, node := range nodes {
		switch node.Type {
		case MenuNodeSeparator:
			continue
		case MenuNodeText:
			if node.Label == "" {
				return tgutil.Errorf("%w: a text item needs a label", ErrMenuItem)
			}

			// The prefix is always there, so the bare id is what the app
			// wrote and what it should be told about.
			id := node.ID[len(MenuIDPrefix):]
			if id == "" {
				return tgutil.Errorf("%w: `%s` needs an id", ErrMenuItem,
					node.Label)
			}

			if ids[node.ID] {
				return tgutil.Errorf("%w: two items share the id `%s`",
					ErrMenuItem, id)
			}

			ids[node.ID] = true
		case MenuNodeSubmenu:
			if node.Label == "" {
				return tgutil.Errorf("%w: a submenu needs a label", ErrMenuItem)
			}

			if err := collectIDs(node.Children, ids); err != nil {
				return tgutil.Errorf("%w", err)
			}
		default:
			return tgutil.Errorf("%w: unknown type `%s`", ErrMenuItem, node.Type)
		}
	}

	return nil
}

// MenuClicked reports whether the click this run is handling is the one on the
// menu item declared under id.
//
// A click naming an id the app's menu does not declare is not a click: the id
// comes from the client, and the menu is what says which ones exist.
//
//	if tgframe.MenuClicked(p, "file_open") {
//		open()
//	}
func MenuClicked(p *Params, id string) bool {
	clickID := MenuID(id)
	return p.State.GetClickID() == clickID && p.State.HasMenuID(clickID)
}
