// Package tgwails runs a ToolGUI app in a desktop window with Wails v2.
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

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// RGBA is a window colour. [options.NewRGBA] builds one.
type RGBA = options.RGBA

// Conf configures the desktop window. The zero value is usable: see
// [DefaultConf] for what the empty fields fall back to.
type Conf struct {
	// Title shows in the title bar.
	Title string

	// Width and Height are the window size in pixels.
	Width, Height int

	// Background is the window background colour.
	Background *RGBA

	// DisableResize stops the user resizing the window.
	DisableResize bool

	// Frameless drops the window decorations.
	Frameless bool
}

// DefaultConf return the config used for the fields a caller leaves empty.
func DefaultConf() *Conf {
	return &Conf{
		Title:      "ToolGUI",
		Width:      1024,
		Height:     768,
		Background: options.NewRGBA(255, 255, 255, 1),
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
	if filled.Background == nil {
		filled.Background = def.Background
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
	backend := NewToolGUI(e.app)

	err := wails.Run(&options.App{
		Title:            e.conf.Title,
		Width:            e.conf.Width,
		Height:           e.conf.Height,
		BackgroundColour: e.conf.Background,
		DisableResize:    e.conf.DisableResize,
		Frameless:        e.conf.Frameless,

		AssetServer: &assetserver.Options{Assets: Assets},

		OnStartup:  backend.start,
		OnShutdown: backend.shutdown,
		Bind:       []any{backend},
	})
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	return nil
}
