package tgframe

import (
	"errors"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgjson"
)

func demoMenu() *Menu {
	return NewMenu().
		Submenu("File", func(m *Menu) {
			m.Text("Open", "file_open")
			m.Separator()
			m.Text("Quit", "file_quit")
		}).
		Text("Help", "help")
}

func TestMenuTree(t *testing.T) {
	nodes := demoMenu().Nodes()

	if len(nodes) != 2 {
		t.Fatalf("len(Nodes()) = %d, want 2", len(nodes))
	}

	file := nodes[0]
	if file.Type != MenuNodeSubmenu || file.Label != "File" {
		t.Fatalf("nodes[0] = %+v, want the File submenu", file)
	}

	want := []MenuNode{
		{Type: MenuNodeText, Label: "Open", ID: "menu_item_file_open"},
		{Type: MenuNodeSeparator},
		{Type: MenuNodeText, Label: "Quit", ID: "menu_item_file_quit"},
	}

	if len(file.Children) != len(want) {
		t.Fatalf("len(File.Children) = %d, want %d",
			len(file.Children), len(want))
	}

	for i, w := range want {
		got := file.Children[i]
		if got.Type != w.Type || got.Label != w.Label || got.ID != w.ID ||
			got.Children != nil {
			t.Errorf("File.Children[%d] = %+v, want %+v", i, *got, w)
		}
	}

	if nodes[1].ID != MenuID("help") {
		t.Errorf("nodes[1].ID = %q, want %q", nodes[1].ID, MenuID("help"))
	}
}

// TestMenuJSON pins what the frontend receives: a separator carries nothing
// but its type, and a text item carries the prefixed click id.
func TestMenuJSON(t *testing.T) {
	bs, err := tgjson.Marshal(demoMenu().Nodes())
	if err != nil {
		t.Fatal(err)
	}

	want := `[{"type":"submenu","label":"File","children":` +
		`[{"type":"text","label":"Open","id":"menu_item_file_open"},` +
		`{"type":"separator"},` +
		`{"type":"text","label":"Quit","id":"menu_item_file_quit"}]},` +
		`{"type":"text","label":"Help","id":"menu_item_help"}]`

	if string(bs) != want {
		t.Errorf("menu JSON =\n%s\nwant\n%s", bs, want)
	}
}

func TestAppConfMenu(t *testing.T) {
	app := NewApp()

	if app.AppConf().Menu != nil {
		t.Error("AppConf().Menu is set on an app with no menu, want nil")
	}

	// An app with no menu must not put the field on the wire either: the
	// frontend keys the menubar row off its absence.
	bs, err := tgjson.Marshal(app.AppConf())
	if err != nil {
		t.Fatal(err)
	}

	var conf map[string]any
	if err := tgjson.Unmarshal(bs, &conf); err != nil {
		t.Fatal(err)
	}

	if _, ok := conf["menu"]; ok {
		t.Error("AppConf JSON carries `menu` for an app with no menu")
	}

	app.SetMenu(demoMenu())
	if len(app.AppConf().Menu) != 2 {
		t.Errorf("len(AppConf().Menu) = %d, want 2", len(app.AppConf().Menu))
	}

	app.SetMenu(nil)
	if app.AppConf().Menu != nil {
		t.Error("AppConf().Menu is set after SetMenu(nil), want nil")
	}
}

func TestSetMenuInvalid(t *testing.T) {
	for _, tt := range []struct {
		name string
		menu *Menu
	}{
		{"no id", NewMenu().Text("Open", "")},
		{"no label", NewMenu().Text("", "open")},
		{"submenu with no label", NewMenu().Submenu("", func(*Menu) {})},
		{"duplicated id", NewMenu().Text("Open", "x").Text("Close", "x")},
		{"duplicated id across submenus", NewMenu().
			Submenu("File", func(m *Menu) { m.Text("Open", "x") }).
			Submenu("Edit", func(m *Menu) { m.Text("Undo", "x") })},
		{"unparseable accelerator", NewMenu().
			Text("Open", "open", &MenuTextConf{Accelerator: "Super+o"})},
		{"duplicated accelerator", NewMenu().
			Text("Open", "open", &MenuTextConf{Accelerator: "CmdOrCtrl+O"}).
			Text("Close", "close", &MenuTextConf{Accelerator: "CmdOrCtrl+o"})},
		{"duplicated accelerator across submenus", NewMenu().
			Submenu("File", func(m *Menu) {
				m.Text("Open", "open", &MenuTextConf{Accelerator: "CmdOrCtrl+O"})
			}).
			Submenu("Edit", func(m *Menu) {
				m.Text("Undo", "undo", &MenuTextConf{
					Accelerator: "cmdorctrl+o",
				})
			})},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					t.Fatal("SetMenu did not panic, want ErrMenuItem")
				}

				err, ok := r.(error)
				if !ok || !errors.Is(err, ErrMenuItem) {
					t.Fatalf("SetMenu panicked with %v, want ErrMenuItem", r)
				}
			}()

			NewApp().SetMenu(tt.menu)
		})
	}
}

// TestMenuClicked is the whole point of the id: the run handling the click
// sees it, the run after it does not.
func TestMenuClicked(t *testing.T) {
	app := NewApp()
	app.SetMenu(demoMenu())

	var seen struct {
		open, quit, undeclared bool
	}
	app.AddPage("index", "Index", func(p *Params) error {
		seen.open = MenuClicked(p, "file_open")
		seen.quit = MenuClicked(p, "file_quit")
		seen.undeclared = MenuClicked(p, "file_save")
		return nil
	})

	state := NewState()

	if err := app.Run("index", state, nil); err != nil {
		t.Fatal(err)
	}
	if seen.open || seen.quit {
		t.Errorf("MenuClicked before any click = %+v, want all false", seen)
	}

	(&EventClick{ID: MenuID("file_open")}).ApplyState(state)
	if err := app.Run("index", state, nil); err != nil {
		t.Fatal(err)
	}
	if !seen.open || seen.quit {
		t.Errorf("MenuClicked on the file_open run = %+v,"+
			" want open only", seen)
	}

	// A click on an id the menu never declared is not a click, however well
	// formed the id is.
	(&EventClick{ID: MenuID("file_save")}).ApplyState(state)
	if err := app.Run("index", state, nil); err != nil {
		t.Fatal(err)
	}
	if seen.undeclared || seen.open || seen.quit {
		t.Errorf("MenuClicked after an undeclared click = %+v,"+
			" want all false", seen)
	}
}

// TestMenuClickedWithoutMenu keeps an app that declares no menu from reporting
// clicks on one.
func TestMenuClickedWithoutMenu(t *testing.T) {
	app := NewApp()

	clicked := false
	app.AddPage("index", "Index", func(p *Params) error {
		clicked = MenuClicked(p, "file_open")
		return nil
	})

	state := NewState()
	(&EventClick{ID: MenuID("file_open")}).ApplyState(state)

	if err := app.Run("index", state, nil); err != nil {
		t.Fatal(err)
	}

	if clicked {
		t.Error("MenuClicked = true on an app with no menu, want false")
	}
}

// TestMenuIDNotAComponentID is what the reserved prefix buys: a menu item's
// click id is not one a component can be given, so the two spaces cannot
// collide.
func TestMenuIDNotAComponentID(t *testing.T) {
	comp := &BaseComponent{Name: "button_component"}
	comp.SetID("file_open")

	if comp.GetID() == MenuID("file_open") {
		t.Errorf("a component took the menu item's id %q", comp.GetID())
	}
}

// TestSetMenuSnapshotsTheTree is what keeps the tree and the declared ids one
// declaration: a caller that goes on using its Menu would otherwise grow the
// menubar without growing the set of ids a click is checked against, and the
// new item would draw but never report.
func TestSetMenuSnapshotsTheTree(t *testing.T) {
	menu := NewMenu().Text("Open", "file_open")

	app := NewApp()
	app.SetMenu(menu)

	menu.Text("Quit", "file_quit")
	menu.Nodes()[0].Label = "Opened"

	conf := app.AppConf()
	if len(conf.Menu) != 1 {
		t.Fatalf("len(AppConf().Menu) = %d after the caller added an item,"+
			" want 1", len(conf.Menu))
	}

	if conf.Menu[0].Label != "Open" {
		t.Errorf("AppConf().Menu[0].Label = %q after the caller renamed it,"+
			" want %q", conf.Menu[0].Label, "Open")
	}
}

// TestSetMenuSnapshotsSubmenus pins the same for a tree the caller reaches
// into rather than appends to.
func TestSetMenuSnapshotsSubmenus(t *testing.T) {
	var sub *Menu
	menu := NewMenu().Submenu("File", func(m *Menu) {
		m.Text("Open", "file_open")
		sub = m
	})

	app := NewApp()
	app.SetMenu(menu)

	sub.Text("Quit", "file_quit")

	children := app.AppConf().Menu[0].Children
	if len(children) != 1 {
		t.Fatalf("len(File.Children) = %d after the caller added an item,"+
			" want 1", len(children))
	}
}

// TestMenuAccelerator is the declaration both carriers read: one spelling on
// the item, normalized so the frontend and the native menu are handed the
// same thing however it was written.
func TestMenuAccelerator(t *testing.T) {
	app := NewApp()
	app.SetMenu(NewMenu().
		Submenu("File", func(m *Menu) {
			m.Text("Open", "file_open", &MenuTextConf{
				Accelerator: "shift+cmdorctrl+O",
			})
			m.Text("Quit", "file_quit")
		}))

	items := app.menu.Nodes()[0].Children
	if got := items[0].Accelerator; got != "CmdOrCtrl+Shift+o" {
		t.Errorf("Open.Accelerator = %q, want %q", got, "CmdOrCtrl+Shift+o")
	}

	// An item that declared none carries none, which is what keeps the field
	// out of the conf the frontend receives.
	if got := items[1].Accelerator; got != "" {
		t.Errorf("Quit.Accelerator = %q, want empty", got)
	}
}

// TestMenuAcceleratorJSON pins the field the frontend reads it under, and
// that an item without one does not carry it at all.
func TestMenuAcceleratorJSON(t *testing.T) {
	app := NewApp()
	app.SetMenu(NewMenu().
		Text("Open", "open", &MenuTextConf{Accelerator: "CmdOrCtrl+O"}).
		Text("Quit", "quit"))

	bs, err := tgjson.Marshal(app.menu.Nodes())
	if err != nil {
		t.Fatalf("Marshal() = %v", err)
	}

	want := `[{"type":"text","label":"Open","id":"menu_item_open",` +
		`"accelerator":"CmdOrCtrl+o"},` +
		`{"type":"text","label":"Quit","id":"menu_item_quit"}]`

	if string(bs) != want {
		t.Errorf("Marshal() = %s, want %s", bs, want)
	}
}

// TestSetMenuNormalizesTheSnapshot: the normalizing happens on the copy
// SetMenu keeps, so a Menu the caller still holds reads back the way they
// wrote it.
func TestSetMenuNormalizesTheSnapshot(t *testing.T) {
	menu := NewMenu().
		Text("Open", "open", &MenuTextConf{Accelerator: "cmdorctrl+O"})

	NewApp().SetMenu(menu)

	if got := menu.Nodes()[0].Accelerator; got != "cmdorctrl+O" {
		t.Errorf("Accelerator = %q, want the declared %q", got, "cmdorctrl+O")
	}
}
