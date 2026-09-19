package tgwails

import (
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgframe"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
)

// testMenu is the tree the desktop tests translate: one submenu holding all
// three node kinds, and a nested submenu under it.
func testMenu() *tgframe.Menu {
	return tgframe.NewMenu().
		Submenu("File", func(m *tgframe.Menu) {
			m.Text("Open", "file_open", &tgframe.MenuTextConf{
				Accelerator: "CmdOrCtrl+Shift+O",
			})
			m.Submenu("More", func(m *tgframe.Menu) {
				m.Text("Reload", "file_reload")
			})
			m.Separator()
			m.Text("Quit", "file_quit")
		})
}

// clickIDs builds the menu with a callback recording what it is handed, and
// returns the menu plus a function reading the ids clicked so far.
//
// The tree goes through an App first, which is where the one Run builds the
// window from comes from: SetMenu is what normalizes an accelerator, so a
// test translating the Menu itself would be translating a spelling the
// window never sees.
func clickIDs(goos string, frameless bool, m *tgframe.Menu) (
	*menu.Menu, func() []string) {

	var got []string
	native := buildMenu(goos, frameless, appNodes(m), func(id string) {
		got = append(got, id)
	})

	return native, func() []string { return got }
}

// appNodes is the tree as the window receives it.
func appNodes(m *tgframe.Menu) []*tgframe.MenuNode {
	app := tgframe.NewApp()
	app.SetMenu(m)

	return app.AppConf().Menu
}

// click fires the item's callback the way Wails does.
func click(item *menu.MenuItem) {
	item.Click(&menu.CallbackData{MenuItem: item})
}

// TestBuildMenuTree is the translation itself: the same tree, in Wails' own
// types, with the three node kinds each landing on their counterpart.
func TestBuildMenuTree(t *testing.T) {
	native, _ := clickIDs("linux", false, testMenu())

	if len(native.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(native.Items))
	}

	file := native.Items[0]
	if file.Type != menu.SubmenuType || file.Label != "File" {
		t.Fatalf("top item = %s %q, want a submenu labelled File",
			file.Type, file.Label)
	}

	items := file.SubMenu.Items
	if len(items) != 4 {
		t.Fatalf("len(File items) = %d, want 4", len(items))
	}

	want := []struct {
		kind  menu.Type
		label string
	}{
		{menu.TextType, "Open"},
		{menu.SubmenuType, "More"},
		{menu.SeparatorType, ""},
		{menu.TextType, "Quit"},
	}
	for i, w := range want {
		if items[i].Type != w.kind || items[i].Label != w.label {
			t.Errorf("File item %d = %s %q, want %s %q",
				i, items[i].Type, items[i].Label, w.kind, w.label)
		}
	}

	nested := items[1].SubMenu.Items
	if len(nested) != 1 || nested[0].Label != "Reload" {
		t.Fatalf("unexpected nested submenu: %v", nested)
	}
}

// TestBuildMenuClickIDs is why the tree is worth translating: an item reports
// the click id the app declared, prefix and all, so the run handling it reads
// the same id tgframe.MenuClicked asks about.
func TestBuildMenuClickIDs(t *testing.T) {
	native, ids := clickIDs("linux", false, testMenu())

	file := native.Items[0].SubMenu.Items
	click(file[0])
	click(file[1].SubMenu.Items[0])
	click(file[3])

	want := []string{
		tgframe.MenuID("file_open"),
		tgframe.MenuID("file_reload"),
		tgframe.MenuID("file_quit"),
	}

	got := ids()
	if len(got) != len(want) {
		t.Fatalf("clicked ids = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("clicked id %d = %q, want %q", i, got[i], want[i])
		}
	}

	// A separator has nothing to click, so it carries no callback to fire.
	if file[2].Click != nil {
		t.Error("a separator carries a click callback")
	}
}

// TestBuildMenuMacRoles covers the shortcuts that come from the standard
// menus rather than from the webview: replacing the macOS menubar without
// them would cost the window Cmd+C, Cmd+V and Cmd+Q.
func TestBuildMenuMacRoles(t *testing.T) {
	native, _ := clickIDs("darwin", false, testMenu())

	want := []menu.Role{menu.AppMenuRole, menu.EditMenuRole, menu.WindowMenuRole}
	if len(native.Items) != len(want)+1 {
		t.Fatalf("len(Items) = %d, want %d", len(native.Items), len(want)+1)
	}

	for i, role := range want {
		if native.Items[i].Role != role {
			t.Errorf("Items[%d].Role = %d, want %d",
				i, native.Items[i].Role, role)
		}
	}

	// The app's own tree follows the standard menus rather than replacing
	// them.
	if native.Items[len(want)].Label != "File" {
		t.Errorf("Items[%d].Label = %q, want File",
			len(want), native.Items[len(want)].Label)
	}

	// A frameless window drops the Window menu, whose entries do nothing
	// there -- the same rule Wails applies to a window with no menu.
	frameless, _ := clickIDs("darwin", true, testMenu())
	for _, item := range frameless.Items {
		if item.Role == menu.WindowMenuRole {
			t.Error("a frameless window got the Window menu")
		}
	}
}

// TestBuildMenuNoRolesOffMac keeps roles where Wails serves them. They are
// macOS only, so anywhere else the app's tree is the whole menubar.
func TestBuildMenuNoRolesOffMac(t *testing.T) {
	for _, goos := range []string{"linux", "windows"} {
		native, _ := clickIDs(goos, false, testMenu())

		if len(native.Items) != 1 || native.Items[0].Label != "File" {
			t.Errorf("%s: unexpected menubar: %v", goos, native.Items)
		}
	}
}

// TestBuildMenuNone is the app that declares no menu. Off macOS it gets no
// menubar at all; macOS gets the standard menus anyway, without which the
// window would have no editing shortcuts.
func TestBuildMenuNone(t *testing.T) {
	for _, goos := range []string{"linux", "windows"} {
		if native := buildMenu(goos, false, nil, func(string) {}); native != nil {
			t.Errorf("%s: want no menu, got %v", goos, native.Items)
		}
	}

	native := buildMenu("darwin", false, nil, func(string) {})
	if native == nil {
		t.Fatal("darwin: want the standard menus, got no menu")
	}

	want := []menu.Role{menu.AppMenuRole, menu.EditMenuRole, menu.WindowMenuRole}
	if len(native.Items) != len(want) {
		t.Fatalf("darwin: len(Items) = %d, want %d",
			len(native.Items), len(want))
	}
	for i, role := range want {
		if native.Items[i].Role != role {
			t.Errorf("darwin: Items[%d].Role = %d, want %d",
				i, native.Items[i].Role, role)
		}
	}
}

// TestNativeMenu covers the entry point Run uses. Whatever the platform puts
// in front of it, the app's own tree is what the menubar ends with.
func TestNativeMenu(t *testing.T) {
	native := nativeMenu(false, appNodes(testMenu()), func(string) {})

	if native == nil {
		t.Fatal("nativeMenu returned no menu for an app that declares one")
	}

	last := native.Items[len(native.Items)-1]
	if last.Type != menu.SubmenuType || last.Label != "File" {
		t.Fatalf("last item = %s %q, want a submenu labelled File",
			last.Type, last.Label)
	}
}

// TestBuildMenuAccelerator is the desktop's whole half of the feature: the
// combination the app declared is hung off the native item, and the OS
// dispatches it. An item that declared none carries none, rather than an
// empty accelerator the menu would draw a gap for.
func TestBuildMenuAccelerator(t *testing.T) {
	native, _ := clickIDs("linux", false, testMenu())

	items := native.Items[0].SubMenu.Items

	open := items[0].Accelerator
	if open == nil {
		t.Fatal("Open carries no accelerator")
	}

	if open.Key != "o" {
		t.Errorf("Open accelerator key = %q, want %q", open.Key, "o")
	}

	want := []keys.Modifier{keys.CmdOrCtrlKey, keys.ShiftKey}
	if len(open.Modifiers) != len(want) {
		t.Fatalf("Open accelerator modifiers = %v, want %v",
			open.Modifiers, want)
	}
	for i := range want {
		if open.Modifiers[i] != want[i] {
			t.Errorf("Open accelerator modifier %d = %q, want %q",
				i, open.Modifiers[i], want[i])
		}
	}

	if items[3].Accelerator != nil {
		t.Errorf("Quit carries the accelerator %v, want none",
			items[3].Accelerator)
	}
}

// TestAcceleratorUnknown: a combination this build cannot parse costs the
// item its shortcut, not its place in the menu.
func TestAcceleratorUnknown(t *testing.T) {
	if got := accelerator(""); got != nil {
		t.Errorf("accelerator(\"\") = %v, want nil", got)
	}

	if got := accelerator("Hyper+o"); got != nil {
		t.Errorf("accelerator(\"Hyper+o\") = %v, want nil", got)
	}
}
