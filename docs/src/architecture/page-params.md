# Page Parameters

> Design note (TG-88). It records the decisions that the implementation tickets
> build on. `Params.Query` and `PageLink` (TG-96), `ReplaceQuery` (TG-97) and
> `Navigate` (TG-103) are implemented. `SetStickyQuery` (TG-98) is not yet.

An analysis tool needs three things that pages cannot do today:

* **Shareable links**: a URL such as `detail?group=a&name=x` that opens a page
  showing that group and that row.
* **Links with parameters**: a link on one page that opens another page with
  arguments.
* **Selections that survive a page switch**: a group picked on one page is
  still picked on the next one.

## Before this design

* Each page load opens its own session. The web frontend keeps the `state_id`
  only in JavaScript, so navigating to another page starts with an empty state
  and every input resets. See [Session Cache](session-cache.md).
* A page function cannot read the URL. The only query string read anywhere is
  in wasm, where `tgwasm.Query()` and `?embed` are read once when the program
  boots.
* No component links to another page of the app.

## Decision summary

| Question | Decision |
| --- | --- |
| URL shape | Page parameters always follow the page name: `/detail?group=a` in path mode, `#/detail?group=a` in hash mode and wasm. The real query string stays app-level. |
| Layer | Page level. `Params.Query` reads parameters, `tgcomp.PageLink` writes them into a link, and `Params.ReplaceQuery` updates the address bar. Components get no URL option. |
| Shared selection | Use URL parameters. The app declares sticky keys that carry over when the user switches pages. No cross-page state. |
| Wails | Links and parameters work. Nothing can be shared because there is no address bar. |
| Security | The framework never writes a parameter into `State`. `Params.Query` is untrusted, size-capped input that the page validates like any other request input. |

## URL shape

**Page parameters sit right after the page name, wherever the page name is.**

| Mode | URL |
| --- | --- |
| Web, path mode | `/detail?group=a&name=x` |
| Web, hash mode | `/#/detail?group=a&name=x` |
| Wasm (always hash) | `/?embed=1#/detail?group=a&name=x` |

The real query string is left alone in hash mode for two reasons:

* In wasm it already means something else. It is app-level and read once at
  boot (`tgwasm.Query()`, `?embed`). Changing it reloads the page, and with it
  the wasm binary. A hash change does not reload anything: `WasmApp` already
  opens a new session on `hashchange`.
* It keeps the page parameters scoped to the page. An embed iframe can keep
  `?embed=1` and still switch between `#/a?x=1` and `#/b?y=2`.

So the single rule for every frontend is: split the page part (the pathname,
or the hash after `#/`) at the first `?`. The left side is the page name and
the right side is the page query. The frontend currently reads the hash with
`hash.substring(2)` in `App.tsx`, `wsapp.tsx` and `WasmApp.tsx`. That would
turn `#/detail?group=a` into a page named `detail?group=a`. These three call
sites move to one shared parser in `@toolgui-web/lib`, with a matching
`pageHref(name, query)` builder.

### Transport

| Executor | How the query reaches the session |
| --- | --- |
| Web | On the update socket URL: `/api/update/{name}?group=a`. A reconnect reads the current location again, so it picks up any `ReplaceQuery`. |
| Wasm | `start(pageName, query)` |
| Wails | `Start(pageName, query string)` |

`NewSession` takes the query, and `Session` holds it as `url.Values`. Each run
receives it as `Params.Query`.

## Layer: page, not component

Making each component opt in (for example `SelectConf.URLParam: "group"`) was
rejected for these reasons:

* The rules about which value wins, and whether a default is written to the
  URL, would be implemented separately in each input component.
* The state lives on the server and is never sent to the client. A widget
  draws from its `Conf.Default` (see
  [Setting an input's initial value](state-storage.md#setting-an-inputs-initial-value)).
  A component-level option would therefore have to write the state and patch
  the default for every input type.
* A parameter often isn't an input at all. A row name on a detail page selects
  data; nothing on the screen edits it.

The page-level design uses three pieces:

```go
// Read: the query the page was opened with.
func Detail(p *tgframe.Params) error {
	group := p.Query.Get("group")

	// The URL seeds the input. Once the user picks something, the state wins.
	def := slices.Index(groups, group) // -1 when unknown: no default
	conf := &tgcomp.SelectConf{}
	if def >= 0 {
		conf.SetDefault(def)
	}
	idx := tgcomp.Select(p.Sidebar, "Group", groups, conf)
	...
}

// Write a link: open another page with arguments.
tgcomp.PageLink(p.Main, "Open", "detail", url.Values{
	"group": {g}, "name": {row.Name},
})

// Write the address bar: keep the current view shareable.
// idx is nil when nothing is selected, so the key is dropped.
q := url.Values{}
if idx != nil {
	q.Set("group", groups[*idx])
}
p.ReplaceQuery(q)
```

Precedence follows from how `Conf.Default` already behaves. The URL decides the
first run. After that, the state decides, because a default applies only while
the state key is absent. On a reload or a shared link the state is empty, so
the URL decides again. The page decides whether a default value goes into the
URL; the framework has no rule for it.

* `PageLink` renders a real `<a href>`, so middle-click and "copy link" work.
  It takes a page name and a query, not a URL. The frontend builds the href with
  the shared `pageHref`, so a link cannot point off the app. An unknown page
  name is a page error (`run.err`).
* `ReplaceQuery` sends a pack. The frontend calls `history.replaceState`: no
  reload, no new session, no history entry per selection. The session stores
  the new query, so later runs and reconnects see it. It replaces the whole
  page query, and the page passes every key it wants to keep.
* The pack is sent when the run ends, just before its result, and only for a
  run no newer event cut. So when runs are interrupted, the value of the last
  finished run is the one that stays. A query over the 8 KiB cap fails the run
  (`run.err`) and nothing is sent.

### Example: a shareable selection

The URL seeds the Select in the sidebar. Each run writes the selection back,
so the address bar always opens this view:

```go
var groups = []string{"a", "b", "c"}

func Detail(p *tgframe.Params) error {
	// Untrusted: only a known group seeds the input.
	conf := &tgcomp.SelectConf{}
	if def := slices.Index(groups, p.Query.Get("group")); def >= 0 {
		conf.SetDefault(def)
	}
	idx := tgcomp.Select(p.Sidebar, "Group", groups, conf)

	// The whole query: keys not set here are dropped.
	q := url.Values{}
	if idx != nil {
		q.Set("group", groups[*idx])
	}
	p.ReplaceQuery(q)

	if idx == nil {
		tgcomp.Text(p.Main, "Pick a group.")
		return nil
	}
	tgcomp.Text(p.Main, "Group "+groups[*idx])
	return nil
}
```

Open `/detail?group=b` and `b` is selected. Pick `c` and the address bar reads
`/detail?group=c`, with no reload and no new history entry. A reload or the
copied URL opens `c`. Unchanged runs send nothing, because the session only
sends a query that differs from the one it holds.

## Navigating from code

`p.Navigate(page, query)` opens another page of the app from a run, the way a
click on a `PageLink` would. It is for a jump with no link to click: a
DataFrame row picked, a button pressed.

* It sends a pack when the run ends, just before the result, and only for a
  run no newer event cut. The last call of a run wins, and it wins over
  `ReplaceQuery`.
* The frontend opens the page the same way as a `PageLink`: path mode moves the
  browser (`/detail?id=0004`), hash mode changes the hash and reloads, wasm
  changes the hash. Each is a history entry, so Back returns to the list. The
  new page gets a new session and a new state.
* It takes a page name, not a URL, the same as `PageLink`. An unknown page or
  a query over 8 KiB fails the run (`run.err`), and the client stays. The
  frontend also drops a pack naming a page it does not know.
* On Wails it opens the page with the query in memory, the same as a
  `PageLink`.

### Example: list to detail

The list page jumps to the detail page when a row is picked. The detail page
reads the id from the URL, so a shared link opens it directly:

```go
func Problems(p *tgframe.Params) error {
	sel := tgcomp.DataFrame(p.Main, []string{"ID", "Title"}, rows,
		&tgcomp.DataFrameConf{Selection: tgcomp.SelectionModeSingle})
	if len(sel) == 1 {
		p.Navigate("detail", url.Values{"id": {ids[sel[0]]}})
	}
	return nil
}

func Detail(p *tgframe.Params) error {
	// Untrusted: only a known id opens a problem.
	prob, ok := problems[p.Query.Get("id")]
	if !ok {
		tgcomp.Text(p.Main, "No such problem.")
		return nil
	}

	tgcomp.PageLink(p.Main, "Back to list", "problems", nil)
	tgcomp.Title(p.Main, prob.Title)
	return nil
}
```

Pick row `0004` and the browser opens `/detail?id=0004`. Back returns to the
list, which opens in a new session, so nothing is selected and no `Slot` or
manual clearing is needed.

## Shared selection: sticky URL keys, not shared state

A cross-page state key was rejected:

* A state belongs to one page load. Sharing one across pages means keeping the
  `state_id` across navigations and running several pages' ids in one state.
  That breaks [TG-36](exposure.md)'s rule that a state only takes ids the last
  run drew, because that set belongs to one page.
* It makes state live longer than any page, which needs its own expiry story.

Instead, the app declares which keys carry over:

```go
app.SetStickyQuery("group")
```

When the user switches pages from the side nav, the frontend copies these keys
from the current page query into the new page URL. `PageLink` adds them too,
unless the link sets them itself. The selection is in the URL because the page
called `ReplaceQuery`, so the URL alone carries it. The server is not involved.

This does not affect TG-90 / TG-95 (`App.RerunAll`). There is no shared state
to watch, so the rerun mechanism stays as designed.

## Wails

There is no address bar, so nothing can be shared. Everything else works:

* `PageLink` and the side nav go through `onNavigate`, which carries the query
  to `Start(pageName, query)`.
* `ReplaceQuery` updates the query that `WailsApp` keeps in memory, so sticky
  keys still carry over.

The docs state this limitation plainly. Wails is not left as unsupported.

## Security

A URL is input that anyone can construct and send to a user.

* **Nothing is written to `State`.** `Params.Query` is a read-only value given
  to the page function. TG-36's id allowlist is untouched, and a query cannot
  grow the state.
* **The page validates what it reads.** It treats the query like any request
  parameter. It looks a value up in a known set (as `slices.Index` does
  above), and never uses it as a path, a command or SQL without checking it.
  The doc for `Params.Query` says so.
* **Size cap.** The query is capped at 8 KiB encoded. Above the cap the
  session is refused with a fatal result, the same as an unknown page.
* **No open redirect.** `PageLink`, `ReplaceQuery` and `Navigate` take a page
  name and values, never a URL. The frontend encodes the values with `URLSearchParams`,
  so neither can change the origin or the scheme.
* Query values are not logged. A row name can be data.

## Out of scope

* **Typed parameters** (`QueryInt`, struct binding). `url.Values` and the
  standard library are enough until repeated parsing code shows up.
* **Web embed mode.** `?embed` is wasm-only today, and nothing here changes it.
