// Package tgwails runs a ToolGUI app in a desktop window with Wails v1.
//
// It's the desktop counterpart of tgexec.WebExecutor: same app, same pages,
// same frontend components, with bound methods and Wails events in place of
// HTTP and websockets.
//
//	app := tgframe.NewApp()
//	app.AddPage("index", "Index", Main)
//	tgwails.NewExecutor(app, &tgwails.Conf{Title: "My Tool"}).Run()
package tgwails

import (
	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgutil"

	"github.com/wailsapp/wails"
)

// Conf configures the desktop window. The zero value is usable: see
// [DefaultConf] for what the empty fields fall back to.
type Conf struct {
	// Title shows in the title bar.
	Title string

	// Width and Height are the window size in pixels.
	Width, Height int

	// Colour is the window background, as "#fff", "rgb(255,255,255)" or
	// "rgba(255,255,255,1)".
	Colour string

	// DisableResize stops the user resizing the window.
	DisableResize bool

	// DisableInspector hides the webview devtools.
	DisableInspector bool
}

// DefaultConf return the config used for the fields a caller leaves empty.
func DefaultConf() *Conf {
	return &Conf{
		Title:  "ToolGUI",
		Width:  1024,
		Height: 768,
		Colour: "#ffffff",
	}
}

func (c *Conf) withDefaults() *Conf {
	def := DefaultConf()
	if c == nil {
		return def
	}

	filled := *c
	if filled.Title == "" {
		filled.Title = def.Title
	}
	if filled.Width == 0 {
		filled.Width = def.Width
	}
	if filled.Height == 0 {
		filled.Height = def.Height
	}
	if filled.Colour == "" {
		filled.Colour = def.Colour
	}

	return &filled
}

// Executor is a desktop ui executor for ToolGUI.
type Executor struct {
	app  *tgframe.App
	conf *Conf
}

// NewExecutor return an Executor serving app in a desktop window.
// A nil conf uses [DefaultConf].
func NewExecutor(app *tgframe.App, conf *Conf) *Executor {
	return &Executor{
		app:  app,
		conf: conf.withDefaults(),
	}
}

// Run open the window and block until the user closes it.
func (e *Executor) Run() error {
	wailsApp := wails.CreateApp(&wails.AppConfig{
		Title:            e.conf.Title,
		Width:            e.conf.Width,
		Height:           e.conf.Height,
		Colour:           e.conf.Colour,
		Resizable:        !e.conf.DisableResize,
		DisableInspector: e.conf.DisableInspector,

		JS:  JS,
		CSS: CSS,
	})

	wailsApp.Bind(NewToolGUI(e.app))

	err := wailsApp.Run()
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	return nil
}
