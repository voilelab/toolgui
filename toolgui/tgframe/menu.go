package tgframe

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

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

	// Accelerator is the key combination a text item also fires on, in the
	// normalized spelling [Menu.Text] stored it under, and empty for an item
	// that declared none.
	Accelerator string `json:"accelerator,omitzero"`

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

// MenuTextConf is the optional part of a [Menu.Text] item.
type MenuTextConf struct {
	// Accelerator is the key combination that fires the item without the menu
	// being opened, written in a platform independent spelling:
	// "CmdOrCtrl+O", "CmdOrCtrl+Shift+F5", "Ctrl+plus". Modifiers are
	// CmdOrCtrl, OptionOrAlt, Shift and Ctrl; the key is one printable ASCII
	// character or one of the named keys (backspace, tab, enter, escape,
	// left, right, up, down, space, delete, home, end, page up, page down,
	// f1 to f24, plus).
	//
	// The desktop hangs it off the native menu item and lets the OS dispatch
	// it. The browser has no such service, so the shell listens for the
	// keystroke itself -- and a combination the browser has already taken is
	// not reliably the app's. See the menu documentation for which ones.
	//
	// [App.SetMenu] panics on one it cannot serve, see [ErrAccelerator].
	Accelerator string
}

// Text adds an item reading label that reports a click under id. Read the
// click with [MenuClicked].
//
//	m.Text("Open", "file_open", &tgframe.MenuTextConf{
//		Accelerator: "CmdOrCtrl+O",
//	})
func (m *Menu) Text(label, id string, conf ...*MenuTextConf) *Menu {
	cf := oneMenuTextConf(conf)

	m.nodes = append(m.nodes, &MenuNode{
		Type:        MenuNodeText,
		Label:       label,
		ID:          MenuID(id),
		Accelerator: cf.Accelerator,
	})

	return m
}

// oneMenuTextConf resolves Text's variadic conf the way [OneConf] resolves a
// component's. It is its own function because a menu item has no id to
// configure -- Text takes it -- so MenuTextConf embeds no [Base] and is not a
// [Conf].
func oneMenuTextConf(conf []*MenuTextConf) *MenuTextConf {
	if len(conf) > 1 {
		panic(fmt.Sprintf(
			"toolgui: Menu.Text takes at most one conf, got %d", len(conf)))
	}

	if len(conf) == 0 || conf[0] == nil {
		return &MenuTextConf{}
	}

	return conf[0]
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

// ErrMenuItem is what a menu the app cannot serve is reported with: an item
// with no label, a text item with no id, two items sharing one, or an
// accelerator the app cannot serve ([ErrAccelerator]).
var ErrMenuItem = tgutil.NewError("invalid menu item")

// checkNodes walks nodes, checking each one, adding every text item's click id
// to ids, and normalizing every accelerator in place -- accels is the set of
// the ones seen, so two items cannot declare the same combination in
// different words.
//
// It writes to the nodes it is given, so it is walked over the snapshot
// [App.SetMenu] took rather than over the caller's own tree.
func checkNodes(nodes []*MenuNode, ids, accels map[string]bool) error {
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

			if err := checkAccelerator(node, accels); err != nil {
				return tgutil.Errorf("%w", err)
			}
		case MenuNodeSubmenu:
			if node.Label == "" {
				return tgutil.Errorf("%w: a submenu needs a label", ErrMenuItem)
			}

			if err := checkNodes(node.Children, ids, accels); err != nil {
				return tgutil.Errorf("%w", err)
			}
		default:
			return tgutil.Errorf("%w: unknown type `%s`", ErrMenuItem, node.Type)
		}
	}

	return nil
}

// checkAccelerator normalizes node's accelerator in place and records it in
// accels. An item that declared none is left alone.
func checkAccelerator(node *MenuNode, accels map[string]bool) error {
	if node.Accelerator == "" {
		return nil
	}

	accel, err := parseAccelerator(node.Accelerator)
	if err != nil {
		return tgutil.Errorf("%w: `%s`: %w", ErrMenuItem, node.Label, err)
	}

	// Two items on one combination is one of them never firing, and which
	// one is whichever the walk reaches first -- not something to leave to
	// the tree's shape.
	if accels[accel] {
		return tgutil.Errorf("%w: two items share the accelerator `%s`",
			ErrMenuItem, accel)
	}

	accels[accel] = true
	node.Accelerator = accel

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
