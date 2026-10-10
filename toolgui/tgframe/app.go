package tgframe

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"net/url"
	"slices"
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
	// Context is cancelled when the run is cut short (a new event, or the
	// session closed). Pass it to slow work so it stops when the user moves
	// on. Never nil.
	Context context.Context

	State   *State
	Main    *Container
	Sidebar *Container

	// Query is the page query, `group=a` in `/detail?group=a` or
	// `#/detail?group=a`. Never nil; each run gets its own copy.
	//
	// It is untrusted input: validate against a known set, and never use it
	// as a path, command or SQL unchecked. It is never written into the
	// State; to seed an input, set its Conf.Default.
	Query url.Values

	run *runState
}

// ReplaceQuery replaces the page query in the address bar with q, with no
// reload, new session or history entry. It replaces the whole query: pass
// every key to keep.
//
// It takes effect when the run ends, unless a newer event cut it; the last
// call wins. Later runs and reconnects read q as [Params.Query]. A q over
// [MaxQuerySize] fails the run. On the desktop the query is only kept in
// memory.
//
//	q := url.Values{}
//	q.Set("group", group)
//	p.ReplaceQuery(q)
func (p *Params) ReplaceQuery(q url.Values) {
	if p.run == nil {
		return
	}

	// A failed call still counts as the last one.
	if err := checkQuery(q); err != nil {
		p.run.query, p.run.queryReplaced = nil, false
		p.run.fail(tgutil.Errorf("ReplaceQuery: %w", err))
		return
	}

	p.run.query = cloneQuery(q)
	p.run.queryReplaced = true
}

// Navigate opens page with q as its [Params.Query], like a click on a
// tgcomp.PageLink: a new session and state, plus a history entry. page is a
// page name of this app, never a URL.
//
// It takes effect when the run ends, unless a newer event cut it; the last
// call wins, also over [Params.ReplaceQuery]. An unknown page or a q over
// [MaxQuerySize] fails the run and drops an earlier ReplaceQuery.
//
//	sel := tgcomp.DataFrame(p.Main, head, rows, &tgcomp.DataFrameConf{
//		Selection: tgcomp.SelectionModeSingle,
//	})
//	if len(sel) == 1 {
//		p.Navigate("detail", url.Values{"id": {ids[sel[0]]}})
//	}
func (p *Params) Navigate(page string, q url.Values) {
	if p.run == nil {
		return
	}

	// A failed call still counts as the last one, over ReplaceQuery too.
	p.run.navigate = nil

	if p.run.app != nil && !p.run.app.HasPage(page) {
		p.run.query, p.run.queryReplaced = nil, false
		p.run.fail(tgutil.Errorf("Navigate: %w: `%s`", ErrPageNotFound, page))
		return
	}

	if err := checkQuery(q); err != nil {
		p.run.query, p.run.queryReplaced = nil, false
		p.run.fail(tgutil.Errorf("Navigate: %w", err))
		return
	}

	p.run.navigate = &Navigation{Page: page, Query: cloneQuery(q)}
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

	// Hidden keeps the page out of the side nav, e.g. a detail page or an
	// embed target. Its url still works, and the nav shows it while open.
	Hidden bool `json:"hidden,omitzero"`
}

// App is an app
type App struct {
	pageNames []string
	pageConfs map[string]*PageConfig
	pageFuncs map[string]RunFunc

	title string
	icon  string
	about string

	// pluginAssets are the file sets served under [PluginAssetPrefix], by
	// name. Locked so a set can be registered while serving.
	pluginAssets map[string]fs.FS
	pluginLock   sync.RWMutex

	hashPageNameMode bool
	showVersion      bool

	// stickyQuery are the page query keys [App.SetStickyQuery] carries over.
	stickyQuery []string

	// menu is the tree [App.SetMenu] took, and menuIDs its click ids. Nil
	// with no menu, which keeps the menubar out of the frontend.
	menu    *Menu
	menuIDs map[string]bool

	// sessions are the open sessions, for [App.RerunAll].
	sessions     map[*Session]struct{}
	sessionsLock sync.Mutex
}

// AppConf store configs for frontend
type AppConf struct {
	PageNames []string               `json:"page_names"`
	PageConfs map[string]*PageConfig `json:"page_confs"`

	// Title names the app itself, after the page title in the browser tab.
	Title string `json:"title"`

	// Icon is the favicon url, if any.
	Icon string `json:"icon,omitzero"`

	HashPageNameMode bool `json:"hash_page_name_mode"`

	// Version is the toolgui version; ShowVersion says whether the side nav
	// shows it.
	Version     string `json:"version"`
	ShowVersion bool   `json:"show_version"`

	// About is the markdown the About dialog shows, if any.
	About string `json:"about,omitzero"`

	MainContainerID    string `json:"main_container_id"`
	SidebarContainerID string `json:"sidebar_container_id"`

	// Menu is the app's menu tree, if any.
	Menu []*MenuNode `json:"menu,omitzero"`

	// StickyQuery are the page query keys carried to the next page, if any.
	StickyQuery []string `json:"sticky_query,omitzero"`
}

// NewApp return App
func NewApp() *App {
	return &App{
		pageNames: make([]string, 0),
		pageConfs: make(map[string]*PageConfig),
		pageFuncs: make(map[string]RunFunc),

		pluginAssets: make(map[string]fs.FS),
		sessions:     make(map[*Session]struct{}),

		showVersion: true,
	}
}

// SetHashPageMode set value of hash page name mode flag.
func (app *App) SetHashPageNameMode(v bool) {
	app.hashPageNameMode = v
}

// SetTitle sets the app title, shown after the page title in the browser tab
// and used in the web manifest and the desktop window.
//
//	app.SetTitle("My Tool")
func (app *App) SetTitle(v string) {
	app.title = v
}

// SetIcon sets the favicon url, replacing the page emoji in the browser tab.
//
//	app.SetIcon("assets/favicon.svg")
func (app *App) SetIcon(url string) {
	app.icon = url
}

// SetMenu declares the app's menu: a menubar above the app on the web, the
// window menubar on the desktop. Read clicks with [MenuClicked].
//
//	app.SetMenu(tgframe.NewMenu().
//		Submenu("File", func(m *tgframe.Menu) {
//			m.Text("Open", "file_open", &tgframe.MenuTextConf{
//				Accelerator: "CmdOrCtrl+O",
//			})
//			m.Separator()
//			m.Text("Quit", "file_quit")
//		}))
//
// The menu belongs to the app, not a page func, since a native menu can't be
// diffed per run. Items may declare accelerators; see [MenuTextConf].
//
// It panics on an invalid menu (see [ErrMenuItem]) so mistakes surface at
// startup. Passing nil drops the menu.
func (app *App) SetMenu(menu *Menu) {
	if menu == nil {
		app.menu, app.menuIDs = nil, nil
		return
	}

	// Snapshot so later changes to the caller's Menu don't leak in; the check
	// also normalizes accelerators in place.
	nodes := cloneNodes(menu.nodes)

	ids := map[string]bool{}
	if err := checkNodes(nodes, ids, map[string]string{}); err != nil {
		panic(err)
	}

	app.menu, app.menuIDs = &Menu{nodes: nodes}, ids
}

// SetShowVersion set whether the side nav shows the toolgui version.
// It is shown by default.
func (app *App) SetShowVersion(v bool) {
	app.showVersion = v
}

// SetAbout sets the app's introduction in markdown, shown in the About dialog
// opened from the side nav's version line.
//
//	app.SetAbout("# My Tool\nConverts CSV to JSON.")
//
// With [App.SetShowVersion] off, the line reads "About" and the dialog shows
// only this.
func (app *App) SetAbout(markdown string) {
	app.about = markdown
}

// SetStickyQuery declares page query keys shared across pages, e.g. a group
// picked in the sidebar. The side nav and PageLink carry them to the next
// page; a page keeps them in its URL with [Params.ReplaceQuery].
//
//	app.SetStickyQuery("group")
//
// Calling it again replaces the keys; no keys drops them.
func (app *App) SetStickyQuery(keys ...string) {
	if len(keys) == 0 {
		app.stickyQuery = nil
		return
	}
	app.stickyQuery = slices.Clone(keys)
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
		Icon:  app.icon,

		MainContainerID:    realMainContainerID(),
		SidebarContainerID: realSidebarContainerID(),

		HashPageNameMode: app.hashPageNameMode,

		Version:     Version(),
		ShowVersion: app.showVersion,
		About:       app.about,

		Menu: app.menuNodes(),

		StickyQuery: app.stickyQuery,
	}
}

// menuNodes returns the menu tree, nil for an app with no menu.
func (app *App) menuNodes() []*MenuNode {
	if app.menu == nil {
		return nil
	}

	return app.menu.nodes
}

// runEndFunc receives what an uncut run asked of its session: the
// [Params.ReplaceQuery] query or the [Params.Navigate] target. At most one is
// set; navigate wins.
type runEndFunc func(replaced url.Values, navigate *Navigation)

// runContextWithHandlingPanic is runContext that turns a panic into an error
// wrapping [ErrPanic]. The panic's error chain is kept, so errors.Is still
// finds [ErrUpdateInterrupt].
func (app *App) runContextWithHandlingPanic(ctx context.Context,
	name string, query url.Values, state *State,
	notifyFunc SendNotifyPackFunc, onEnd runEndFunc) (err error) {

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

		// An interrupt is how a cut run unwinds, not worth logging.
		if !errors.Is(rErr, ErrUpdateInterrupt) {
			log.Println("Panic", r)
		}
	}()

	err = app.runContext(ctx, name, query, state, notifyFunc, onEnd)
	return
}

// Run runs page `name` once, outside any session, with a background context
// and an empty [Params.Query]. A panic in the page is not recovered. Meant for
// tests; pages are served through a [Session].
func (app *App) Run(name string, state *State, notifyFunc SendNotifyPackFunc) error {
	return app.runContext(context.Background(), name, nil, state, notifyFunc, nil)
}

// runContext runs a page. onEnd, if not nil, gets the last valid
// [Params.ReplaceQuery] or [Params.Navigate] of an uncut run.
func (app *App) runContext(ctx context.Context, name string,
	query url.Values, state *State, notifyFunc SendNotifyPackFunc,
	onEnd runEndFunc) error {
	pageFunc, ok := app.pageFuncs[name]
	if !ok {
		return tgutil.Errorf("%w: `%s`", ErrPageNotFound, name)
	}

	// Menu ids are the app's, so they are set here for [MenuClicked] to check.
	if state != nil {
		state.setMenuIDs(app.menuIDs)
	}

	run := newRunState()
	run.app = app

	newMain := NewContainer(MainContainerID, state, notifyFunc)
	newMain.run = run
	newSidebar := NewContainer(SidebarContainerID, state, notifyFunc)
	newSidebar.run = run

	// Claim the root ids so a page container can't reuse them.
	run.registerID(newMain)
	run.registerID(newSidebar)

	err := pageFunc(&Params{
		Context: ctx,
		State:   state,
		Main:    newMain,
		Sidebar: newSidebar,
		Query:   cloneQuery(query),
		run:     run,
	})

	// A cut run leaves the previous run on screen, so keep its ids.
	if ctx.Err() != nil {
		return NewPageError(err)
	}

	if onEnd != nil {
		switch {
		case run.navigate != nil:
			onEnd(nil, run.navigate)
		case run.queryReplaced:
			onEnd(run.query, nil)
		}
	}

	// Record what is on screen; uploads are checked against these ids.
	if state != nil {
		state.setRunIDs(run.ids, run.indexedFileIDs)
	}

	if err != nil {
		// The page's own error is shown to the user as is.
		return NewPageError(err)
	}

	// Drop state under ids the finished run released and never rewrote.
	if state != nil {
		for id := range run.released {
			state.Delete(id)
		}
	}

	return NewPageError(run.err)
}

// RerunAll reruns every open page, as if its user pressed rerun. Call it when
// data the pages read has changed.
//
// It returns at once and never cuts a run: a running page reruns after it
// ends, and calls during a run fold into one rerun.
func (app *App) RerunAll() {
	app.rerunSessions(func(*Session) bool { return true })
}

// RerunPage is [App.RerunAll] for the open pages named in names only.
func (app *App) RerunPage(names ...string) {
	set := make(map[string]bool, len(names))
	for _, name := range names {
		set[name] = true
	}

	app.rerunSessions(func(s *Session) bool { return set[s.pageName] })
}

func (app *App) rerunSessions(match func(*Session) bool) {
	app.sessionsLock.Lock()
	sessions := make([]*Session, 0, len(app.sessions))
	for s := range app.sessions {
		if match(s) {
			sessions = append(sessions, s)
		}
	}
	app.sessionsLock.Unlock()

	for _, s := range sessions {
		s.requestRerun()
	}
}

func (app *App) addSession(s *Session) {
	app.sessionsLock.Lock()
	defer app.sessionsLock.Unlock()

	app.sessions[s] = struct{}{}
}

func (app *App) removeSession(s *Session) {
	app.sessionsLock.Lock()
	defer app.sessionsLock.Unlock()

	delete(app.sessions, s)
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
