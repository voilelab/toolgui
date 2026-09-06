package tgframe

import (
	"errors"
	"log"

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
	return NewContainer(MainContainerID, nil).ID
}

// SidebarContainerID is the id of sidebar container.
// The creation of root container won't trigger SendNotifyPackFunc.
const SidebarContainerID string = "container_sidebar"

func realSidebarContainerID() string {
	return NewContainer(SidebarContainerID, nil).ID
}

type Params struct {
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

	hashPageNameMode bool
	showVersion      bool
}

// AppConf store configs for frontend
type AppConf struct {
	PageNames []string               `json:"page_names"`
	PageConfs map[string]*PageConfig `json:"page_confs"`

	HashPageNameMode bool `json:"hash_page_name_mode"`

	// Version is the toolgui version, and ShowVersion whether the side nav
	// should show it.
	Version     string `json:"version"`
	ShowVersion bool   `json:"show_version"`

	MainContainerID    string `json:"main_container_id"`
	SidebarContainerID string `json:"sidebar_container_id"`
}

// NewApp return App
func NewApp() *App {
	return &App{
		pageNames: make([]string, 0),
		pageConfs: make(map[string]*PageConfig),
		pageFuncs: make(map[string]RunFunc),

		showVersion: true,
	}
}

// SetHashPageMode set value of hash page name mode flag.
func (app *App) SetHashPageNameMode(v bool) {
	app.hashPageNameMode = v
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

		MainContainerID:    realMainContainerID(),
		SidebarContainerID: realSidebarContainerID(),

		HashPageNameMode: app.hashPageNameMode,

		Version:     Version(),
		ShowVersion: app.showVersion,
	}
}

// Run run a page which named `name` with state.
// Return a error wrap with ErrPanic if encounter panic.
func (app *App) RunWithHandlingPanic(
	name string, state *State, notifyFunc SendNotifyPackFunc) (err error) {

	defer func() {
		r := recover()
		if r != nil {
			log.Println("Panic", r)
			err = tgutil.Errorf("%w: %v", ErrPanic, r)
		}
	}()

	err = app.Run(name, state, notifyFunc)
	return
}

// Run run a page which named `name` with state.
func (app *App) Run(name string, state *State, notifyFunc SendNotifyPackFunc) error {
	pageFunc, ok := app.pageFuncs[name]
	if !ok {
		return tgutil.Errorf("%w: `%s`", ErrPageNotFound, name)
	}

	newMain := NewContainer(MainContainerID, notifyFunc)
	newSidebar := NewContainer(SidebarContainerID, notifyFunc)

	err := pageFunc(&Params{
		State:   state,
		Main:    newMain,
		Sidebar: newSidebar,
	})
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	return nil
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
