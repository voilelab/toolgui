package tgwails

import (
	goruntime "runtime"

	"github.com/voilelab/toolgui/toolgui/tgframe"

	"github.com/wailsapp/wails/v2/pkg/menu"
)

// nativeMenu translates the tree [tgframe.App.SetMenu] declared into the one
// Wails hangs off the window. Same declaration the web menubar draws from:
// only the thing drawing it changes.
//
// It returns nil where there is nothing to draw -- an app with no menu, on a
// platform that adds no standard ones -- leaving the window without a
// menubar.
func nativeMenu(frameless bool, nodes []*tgframe.MenuNode,
	click func(id string)) *menu.Menu {

	return buildMenu(goruntime.GOOS, frameless, nodes, click)
}

// buildMenu is [nativeMenu] with the platform named, so a test can build the
// menu of a platform it is not running on.
func buildMenu(goos string, frameless bool, nodes []*tgframe.MenuNode,
	click func(id string)) *menu.Menu {

	roles := roleItems(goos, frameless)
	if len(roles) == 0 && len(nodes) == 0 {
		return nil
	}

	native := menu.NewMenu()
	for _, item := range roles {
		native.Append(item)
	}

	appendNodes(native, nodes, click)

	return native
}

// roleItems returns the standard menus that go in front of the app's own
// entries.
//
// Roles are macOS only -- Wails v2 says so, and the constants for them are
// commented out everywhere else -- so nothing goes in front on the other
// platforms. macOS needs them: Cmd+C, Cmd+V and Cmd+Q are actions of the
// standard App and Edit menus rather than of the webview, so a window whose
// menu replaces them has none of those shortcuts.
//
// This is what Wails itself puts in when no menu is set, Frameless rule
// included, so declaring a menu does not cost the window a standard one.
func roleItems(goos string, frameless bool) []*menu.MenuItem {
	if goos != "darwin" {
		return nil
	}

	items := []*menu.MenuItem{menu.AppMenu(), menu.EditMenu()}
	if !frameless {
		// The Window menu's entries do nothing on a frameless window.
		items = append(items, menu.WindowMenu())
	}

	return items
}

// appendNodes writes nodes into native, recursing through submenus.
//
// Text, separator and submenu are the three kinds a [tgframe.Menu] can hold,
// and SetMenu turns away a tree carrying anything else, so an unknown node is
// nothing to report on here -- it is dropped.
func appendNodes(native *menu.Menu, nodes []*tgframe.MenuNode,
	click func(id string)) {

	for _, node := range nodes {
		switch node.Type {
		case tgframe.MenuNodeSeparator:
			native.AddSeparator()
		case tgframe.MenuNodeText:
			// The id is the one the app declared, prefix and all, so the run
			// handling the click reads it back with tgframe.MenuClicked just
			// as it does the web menubar's.
			native.AddText(node.Label, nil, func(*menu.CallbackData) {
				click(node.ID)
			})
		case tgframe.MenuNodeSubmenu:
			appendNodes(native.AddSubmenu(node.Label), node.Children, click)
		}
	}
}
