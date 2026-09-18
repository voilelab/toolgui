package tgframe

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"sync"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// ErrPageNotFound is the error that the page is not found.
var ErrPageNotFound = errors.New("page not found")

// ErrPanic is the error that the app panics.
var ErrPanic = errors.New("panic")

// MainContainerID is the id of root container.
// The creation of root container won't trigger SendNotifyPackFunc.
const MainContainerID string = "container_main"

func realMainContainerID() string {
	return containerID(MainContainerID)
}

// SidebarContainerID is the id of sidebar container.
// The creation of root container won't trigger SendNotifyPackFunc.
const SidebarContainerID string = "container_sidebar"

func realSidebarContainerID() string {
	return containerID(SidebarContainerID)
}

type Params struct {
	// Context is cancelled when the run is cut short: a new event arrived, or
	// the session closed. Hand it to whatever the page does slowly — a request,
	// a query — and that work stops as soon as the user moves on, instead of
	// running until the page next draws something.
	//
	// It is never nil.
	Context context.Context

	State   *State
	Main    *Container
	Sidebar *Container
}

// RunFunc is the type of a function handling page
type RunFunc func(*Params) error

// PageConfig stores basic setting of a page
type PageConfig struct {
	// Name should not duplicate to another page
	Name string `json:"name"`

	// Title will show as the title of page
	Title string `json:"title"`

	// Emoji will show as icon of a page
	Emoji string `json:"emoji"`
}

// App is an app
type App struct {
	pageNames []string
	pageConfs map[string]*PageConfig
	pageFuncs map[string]RunFunc

	title string

	// pluginAssets are the file sets served under [PluginAssetPrefix], by name.
	// A set is looked up per request, so the lock is what lets one be
	// registered while the app is already serving.
	pluginAssets map[string]fs.FS
	pluginLock   sync.RWMutex

	hashPageNameMode bool
	showVersion      bool

	// menu is the tree [App.SetMenu] took, and menuIDs the click ids in it.
	// Both are nil until an app declares a menu, which is what keeps the
	// menubar out of the frontend's DOM entirely.
	menu    *Menu
	menuIDs map[string]bool
}

// AppConf store configs for frontend
type AppConf struct {
	PageNames []string               `json:"page_names"`
	PageConfs map[string]*PageConfig `json:"page_confs"`

	// Title names the app itself, after the page title in the browser tab.
	Title string `json:"title"`

	HashPageNameMode bool `json:"hash_page_name_mode"`

	// Version is the toolgui version, and ShowVersion whether the side nav
	// should show it.
	Version     string `json:"version"`
	ShowVersion bool   `json:"show_version"`

	MainContainerID    string `json:"main_container_id"`
	SidebarContainerID string `json:"sidebar_container_id"`

	// Menu is the app's menu tree, absent for an app that declares none.
	Menu []*MenuNode `json:"menu,omitzero"`
}

// NewApp return App
func NewApp() *App {
	return &App{
		pageNames: make([]string, 0),
		pageConfs: make(map[string]*PageConfig),
		pageFuncs: make(map[string]RunFunc),

		pluginAssets: make(map[string]fs.FS),

		showVersion: true,
	}
}

// SetHashPageMode set value of hash page name mode flag.
func (app *App) SetHashPageNameMode(v bool) {
	app.hashPageNameMode = v
}

// SetTitle sets the app title. The browser tab shows it after the page title,
// and it names the app in the web manifest and in the desktop window.
//
//	app.SetTitle("My Tool")
func (app *App) SetTitle(v string) {
	app.title = v
}

// SetMenu declares the app's menu. The frontend draws it as a menubar above
// the app, and a click on one of its items is read with [MenuClicked].
//
//	app.SetMenu(tgframe.NewMenu().
//		Submenu("File", func(m *tgframe.Menu) {
//			m.Text("Open", "file_open")
//			m.Separator()
//			m.Text("Quit", "file_quit")
//		}))
//
// The tree belongs to the app rather than to a page func: it is declared once
// and stands for every run, which is the only shape a native menu -- with no
// diff to apply -- can take.
//
// It panics on a menu the app cannot serve, so a mistake is reported at
// startup rather than by a menubar that quietly misses an item. See
// [ErrMenuItem] for what counts as one. Passing nil drops the menu.
func (app *App) SetMenu(menu *Menu) {
	if menu == nil {
		app.menu, app.menuIDs = nil, nil
		return
	}

	ids := map[string]bool{}
	if err := collectIDs(menu.nodes, ids); err != nil {
		panic(err)
	}

	app.menu, app.menuIDs = menu, ids
}

// SetShowVersion set whether the side nav shows the toolgui version.
// It is shown by default.
func (app *App) SetShowVersion(v bool) {
	app.showVersion = v
}

// AddPage add a handled page by name, title, and runFunc.
//
//	app.AddPage("index", "Index", f})
func (app *App) AddPage(name, title string, runFunc RunFunc) {
	err := app.addPageByConfig(&PageConfig{
		Name:  name,
		Title: title,
	}, runFunc)
	if err != nil {
		panic(err)
	}
}

// AddPageByConfig add a handled page by name, title, icon, and runFunc.
//
//	app.AddPageByConfig(&tgframe.PageConfig{
//		Name:  "page1",
//		Title: "Page1",
//		Emoji: "🐱",
//	}, Page1)
func (app *App) AddPageByConfig(conf *PageConfig, runFunc RunFunc) {
	err := app.addPageByConfig(conf, runFunc)
	if err != nil {
		panic(err)
	}
}

func (app *App) addPageByConfig(conf *PageConfig, runFunc RunFunc) error {
	if conf == nil {
		return tgutil.NewError("nil config")
	}

	if conf.Name == "" || conf.Name == "api" || conf.Name == "static" {
		return tgutil.NewError("name should not be empty or 'api' or 'static'")
	}

	if _, exist := app.pageConfs[conf.Name]; exist {
		return tgutil.NewError("name duplicate")
	}

	app.pageFuncs[conf.Name] = runFunc
	app.pageConfs[conf.Name] = conf
	app.pageNames = append(app.pageNames, conf.Name)

	return nil
}

// AppConf return [AppConf].
func (app *App) AppConf() *AppConf {
	return &AppConf{
		PageNames: app.pageNames,
		PageConfs: app.pageConfs,

		Title: app.title,

		MainContainerID:    realMainContainerID(),
		SidebarContainerID: realSidebarContainerID(),

		HashPageNameMode: app.hashPageNameMode,

		Version:     Version(),
		ShowVersion: app.showVersion,

		Menu: app.menuNodes(),
	}
}

// menuNodes returns the menu tree, nil for an app with no menu.
func (app *App) menuNodes() []*MenuNode {
	if app.menu == nil {
		return nil
	}

	return app.menu.nodes
}

// RunWithHandlingPanic run a page which named `name` with state, with a
// context that is never cancelled. See [App.RunContextWithHandlingPanic].
func (app *App) RunWithHandlingPanic(
	name string, state *State, notifyFunc SendNotifyPackFunc) error {

	return app.RunContextWithHandlingPanic(
		context.Background(), name, state, notifyFunc)
}

// RunContextWithHandlingPanic run a page which named `name` with state and ctx.
// Return a error wrap with ErrPanic if encounter panic. A panicked error keeps
// its chain, so errors.Is still finds [ErrUpdateInterrupt] under [ErrPanic].
func (app *App) RunContextWithHandlingPanic(ctx context.Context,
	name string, state *State, notifyFunc SendNotifyPackFunc) (err error) {

	defer func() {
		r := recover()
		if r == nil {
			return
		}

		rErr, ok := r.(error)
		if !ok {
			log.Println("Panic", r)
			err = tgutil.Errorf("%w: %v", ErrPanic, r)
			return
		}

		err = tgutil.Errorf("%w: %w", ErrPanic, rErr)

		// An interrupt is how a cut run unwinds, not a failure worth a line.
		if !errors.Is(rErr, ErrUpdateInterrupt) {
			log.Println("Panic", r)
		}
	}()

	err = app.RunContext(ctx, name, state, notifyFunc)
	return
}

// Run run a page which named `name` with state, with a context that is never
// cancelled. See [App.RunContext].
func (app *App) Run(name string, state *State, notifyFunc SendNotifyPackFunc) error {
	return app.RunContext(context.Background(), name, state, notifyFunc)
}

// RunContext run a page which named `name` with state, and hands ctx to the
// page func as [Params.Context].
func (app *App) RunContext(ctx context.Context,
	name string, state *State, notifyFunc SendNotifyPackFunc) error {
	pageFunc, ok := app.pageFuncs[name]
	if !ok {
		return tgutil.Errorf("%w: `%s`", ErrPageNotFound, name)
	}

	// The menu is the app's, not the run's, so the ids it declares are told
	// to the state here rather than claimed like a component's. [MenuClicked]
	// reads them back to turn away a click on an item the app never declared.
	if state != nil {
		state.setMenuIDs(app.menuIDs)
	}

	run := newRunState()

	newMain := NewContainer(MainContainerID, state, notifyFunc)
	newMain.run = run
	newSidebar := NewContainer(SidebarContainerID, state, notifyFunc)
	newSidebar.run = run

	// The roots are never sent, so nothing else claims their ids. Claim them
	// here, or a container the page adds under one of their names would take
	// the root's id without colliding with anything.
	run.registerID(newMain)
	run.registerID(newSidebar)

	err := pageFunc(&Params{
		Context: ctx,
		State:   state,
		Main:    newMain,
		Sidebar: newSidebar,
	})

	// A cut run stopped partway, so the run before it is still what the
	// client is looking at. Leave that run's ids and released ids alone, or a
	// click or an upload naming one of its components would be turned away as
	// a name the page never drew. A run cut short at a notify pack panics out
	// before this, and one watching the context returns here.
	if ctx.Err() != nil {
		return NewPageError(err)
	}

	// The page function returned, so what it claimed is what is on the screen.
	// Record it: an upload names a component id, and the state is where that
	// name is checked.
	if state != nil {
		state.setRunIDs(run.ids)
	}

	if err != nil {
		// What the page function returned is what the page wants its user to
		// read, so it keeps its message instead of being masked -- and keeps
		// it whole, without a function path in front of it.
		return NewPageError(err)
	}

	// The page ran to the end, so an id it cleared and never wrote again names
	// nothing on the screen. Drop the state under it, or the next run would
	// hand that value to whatever lands on the id.
	if state != nil {
		for id := range run.released {
			state.Delete(id)
		}
	}

	// A component the page gave up on is the page's own report too.
	return NewPageError(run.err)
}

// HasPage return existence of page which named `name`.
func (app *App) HasPage(name string) bool {
	_, ok := app.pageFuncs[name]
	return ok
}

// FirstPage return the first page in the app.
func (app *App) FirstPage() (string, bool) {
	if len(app.pageNames) == 0 {
		return "", false
	}
	return app.pageNames[0], true
}
