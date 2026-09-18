package main

import (
	"archive/zip"
	"crypto/md5"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"strconv"

	"github.com/voilelab/toolgui/docs/demos"
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

//go:embed main.go
var code string

// The demo plugin ships as the files it is made of, served under
// /plugin/colorpicker/.
//
//go:embed plugins/colorpicker
var colorPickerAssets embed.FS

const readme = `
# [ToolGUI](https://github.com/voilelab/toolgui)

This Go package provides a framework for rapidly building interactive data
dashboards and web applications. It aims to offer a similar development
experience to Streamlit for Python users.

> [!WARNING]
> ⚠️ Under Development:
> 
> The API for this package is still under development,
> and may be subject to changes in the future.

## Example

` + "```go" + `
package main

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgexec"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func main() {
	app := tgframe.NewApp()
	app.AddPage("index", "Index", func(p *tgframe.Params) error {
		tgcomp.Text(p.Main, "Hello world")
		return nil
	})

	tgexec.NewWebExecutor(app).StartService("127.0.0.1:3001")
}
` + "```"

func SourceCodePage(p *tgframe.Params) error {
	tgcomp.Title(p.Main, "Example for ToolGUI")
	tgcomp.Code(p.Main, code)
	return nil
}

func MainPage(p *tgframe.Params) error {
	tgcomp.Markdown(p.Main, readme)
	return nil
}

// demoMenu is the menubar the app declares. Its items are read back on the
// menu page, and nowhere else: the tree belongs to the app, so the same items
// are there whichever page is open.
func demoMenu() *tgframe.Menu {
	return tgframe.NewMenu().
		Submenu("File", func(m *tgframe.Menu) {
			m.Text("Say hello", "hello")
			m.Submenu("More", func(m *tgframe.Menu) {
				m.Text("Say hello loudly", "hello_loud")
			})
			m.Separator()
			m.Text("Clear the log", "clear")
		}).
		Submenu("Help", func(m *tgframe.Menu) {
			m.Text("About", "about")
		})
}

// menuLogKey is where the menu page keeps what has been picked. The state is
// the session's, so the log survives a page func that only draws it.
const menuLogKey = "demo_menu_log"

func MenuPage(p *tgframe.Params) error {
	tgcomp.Title(p.Main, "App Menu")
	tgcomp.Text(p.Main, "The menubar above the app comes from App.SetMenu."+
		" Pick an item and the run handling the click appends to this log.")

	log, _ := p.State.Get[[]string](menuLogKey)

	// Numbered, so a line says which run wrote it: picking the same item
	// twice reads as two entries rather than as one that may not have moved.
	pick := func(what string) {
		log = append(log, fmt.Sprintf("%d. %s", len(log)+1, what))
	}

	switch {
	case tgframe.MenuClicked(p, "hello"):
		pick("File > Say hello")
	case tgframe.MenuClicked(p, "hello_loud"):
		pick("FILE > MORE > SAY HELLO LOUDLY")
	case tgframe.MenuClicked(p, "clear"):
		log = nil
	case tgframe.MenuClicked(p, "about"):
		pick("Help > About: toolgui " + tgframe.Version())
	}

	p.State.Set(menuLogKey, log)

	if len(log) == 0 {
		tgcomp.Caption(p.Main, "Nothing picked yet.")
		return nil
	}

	for _, line := range log {
		tgcomp.Text(p.Main, line)
	}

	return nil
}

func SidebarPage(p *tgframe.Params) error {
	if tgcomp.Checkbox(p.Main, "Show sidebar") {
		tgcomp.Text(p.Sidebar, "Sidebar is here")
	}

	tgcomp.Code(p.Main, code)
	return nil
}

// headerRow names the two columns every example is laid out in.
func headerRow(p *tgframe.Params) {
	compCol, codeCol := tgcomp.EqColumn2(
		p.Main, &tgcomp.ColumnConf{ID: "header_of_rows"})
	tgcomp.Subtitle(compCol, "Component")
	tgcomp.Subtitle(codeCol, "Code")
}

// blockRow draws one example: what it renders on the left, and on the right
// the source it was rendered from -- the same slice of docs/demos the book
// includes, so the two cannot disagree.
//
// A failing example takes the page with it, as it did when the demo wrote its
// own rows, which is what the error example is there to show.
func blockRow(p *tgframe.Params, b demos.Block) error {
	compCol, codeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: b.ID})

	err := b.Run(&tgframe.Params{
		Context: p.Context,
		State:   p.State,
		Main:    compCol,
		Sidebar: p.Sidebar,
	})
	if err != nil {
		return err
	}

	tgcomp.Code(codeCol, b.Code)
	return nil
}

// dividers draws the line between two examples. The input page gives each one
// an id of its own, which is where the "1", "2", ... in its DOM come from.
type dividers struct {
	numbered bool
	drawn    int
}

func (d *dividers) draw(c *tgframe.Container) {
	d.drawn++

	if !d.numbered {
		tgcomp.Divider(c)
		return
	}

	tgcomp.Divider(c, &tgcomp.DividerConf{ID: strconv.Itoa(d.drawn)})
}

// demoPage is a component's own page, the one /#/<component> lands on.
func demoPage(d *demos.Demo) tgframe.RunFunc {
	return func(p *tgframe.Params) error {
		headerRow(p)

		div := &dividers{}
		for i, b := range d.Blocks {
			if i != 0 {
				div.draw(p.Main)
			}

			if err := blockRow(p, b); err != nil {
				return err
			}
		}

		if d.Sidebar == nil {
			return nil
		}

		return d.Sidebar(p)
	}
}

// groupPage is a coarse page: every example of a category, the way the demo
// was organised before each component had a page.
func groupPage(g *demos.Group) tgframe.RunFunc {
	return func(p *tgframe.Params) error {
		headerRow(p)

		div := &dividers{numbered: g.NumberDividers}
		if g.LeadDivider {
			div.draw(p.Main)
		}

		drawn := 0
		sidebars := []tgframe.RunFunc{}

		for _, d := range g.Demos {
			for _, b := range d.Blocks {
				if drawn != 0 {
					div.draw(p.Main)
				}
				drawn++

				if err := blockRow(p, b); err != nil {
					return err
				}
			}

			if d.Sidebar != nil {
				sidebars = append(sidebars, d.Sidebar)
			}
		}

		// A side column example is about the side column rather than about a
		// row, so it is written once the rows are done.
		for _, run := range sidebars {
			if err := run(p); err != nil {
				return err
			}
		}

		return nil
	}
}

func getFiles(p *tgframe.Params, f *tcinput.FileObject) ([]string, error) {
	fp, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	// The key covers the content, and hashing it through a stream keeps the
	// file off the heap.
	hash := md5.New()
	if _, err := io.Copy(hash, fp); err != nil {
		return nil, err
	}

	// The key names the function as well as the file, so moving these two
	// calls around does not change what they read.
	key := fmt.Sprintf("getFiles_%s_%s_%x", f.Name, f.Type, hash.Sum(nil))

	if v, ok := p.State.GetFuncCache[[]string](key); ok {
		slog.Info("cache found")
		return v, nil
	}

	// zip reads at an offset, which it can do straight against the file.
	cbzFp, err := zip.NewReader(fp, int64(f.Size))
	if err != nil {
		return nil, err
	}

	ret := []string{}
	for _, f := range cbzFp.File {
		ret = append(ret, f.Name)
	}

	p.State.SetFuncCache(key, ret)
	return ret, nil
}

func FuncCachePage(p *tgframe.Params) error {
	cbzfile := tgcomp.FileUpload(p.Sidebar, "CBZ File", "application/x-cbz")

	if cbzfile == nil {
		return nil
	}

	files, err := getFiles(p, cbzfile)
	if err != nil {
		return err
	}

	for i, f := range files {
		tgcomp.Text(p.Main, fmt.Sprintf("%d: %s", i, f))
	}

	return nil
}

// newApp build the demo. main lives in main_server.go and main_wasm.go: the
// pages are the same either way, only the executor differs.
//
// The examples come from docs/demos, which is also where the book includes
// them from, so a page here is a table entry rather than a function of its
// own.
func newApp() *tgframe.App {
	app := tgframe.NewApp()

	// The title trails the page title in the browser tab, and names the app
	// in the manifest main_server.go sets unless that gives its own name.
	app.SetTitle("ToolGUI Demo")

	// The menubar is the app's, so it is there on every page. Embedded in the
	// book the row is dropped along with the rest of the app's chrome.
	app.SetMenu(demoMenu())

	app.AddPage("index", "Index", MainPage)

	// A category page, then a page per component under it: the nav reads the
	// way the book's contents do, and a reader following a link from the book
	// lands on the one component they came for.
	for _, g := range demos.Groups() {
		app.AddPage(g.Name, g.Title, groupPage(g))

		for _, d := range g.Demos {
			app.AddPage(d.Name, d.Title, demoPage(d))
		}
	}

	app.AddPage("sidebar", "Sidebar", SidebarPage)
	app.AddPage("menu", "App Menu", MenuPage)
	app.AddPage("function_cache", "Function Cache", FuncCachePage)
	app.AddPage("code", "Source Code", SourceCodePage)

	return app
}

// addPluginDemo adds the plugin page and the files its plugin is made of.
// It's the server build's to call: a plugin is loaded over a url, and the
// browser build has no executor serving one.
func addPluginDemo(app *tgframe.App) error {
	assets, err := fs.Sub(colorPickerAssets, "plugins/colorpicker")
	if err != nil {
		return err
	}

	if err := app.AddPluginAssets("colorpicker", assets); err != nil {
		return err
	}

	d := demos.Plugin()
	app.AddPage(d.Name, d.Title, demoPage(d))
	return nil
}
