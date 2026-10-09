# App Cache

An app-level cache stores data for the entire duration of the application's execution,
from launch to termination.

This cache is not provided by ToolGUI and needs to be implemented by the developer.

For example:

```go
type App struct {
	data sync.Map
}

func (app *App) QueryData(key string) any {
	if v, ok := app.data.Load(key); ok {
		return v
	}

	// calculate the value, then keep the first one stored
	calVal := calculate(key)
	v, _ := app.data.LoadOrStore(key, calVal)
	return v
}

func (app *App) Page1(p *tgframe.Params) error {
	tgcomp.Text(p.Main, app.QueryData("key").(string))
	return nil
}
```

## Loading in the background

A cache that takes long to fill should not hold up the service. Start the
service right away, fill the cache in a goroutine, and let the page draw a
loading screen until it is ready. When the load finishes, `RerunAll` reruns
the pages already open, so they show the data without the user doing
anything:

```go
type App struct {
	tgApp *tgframe.App

	mu   sync.RWMutex
	data *Data // nil until loaded
	err  error
}

func (app *App) load() {
	data, err := loadData() // slow

	app.mu.Lock()
	app.data, app.err = data, err
	app.mu.Unlock()

	// Pages opened during the load still show the loading screen.
	app.tgApp.RerunAll()
}

// snapshot returns what is loaded so far; data is nil while loading.
func (app *App) snapshot() (*Data, error) {
	app.mu.RLock()
	defer app.mu.RUnlock()
	return app.data, app.err
}

func (app *App) Page1(p *tgframe.Params) error {
	data, err := app.snapshot()
	if err != nil {
		return err
	}

	if data == nil {
		tgcomp.Spinner(p.Main, "Loading data...")
		return nil
	}

	tgcomp.Text(p.Main, data.Summary())
	return nil
}

func main() {
	tgApp := tgframe.NewApp()
	app := &App{tgApp: tgApp}
	tgApp.AddPage("page1", "Page1", app.Page1)

	go app.load()

	// Serves at once; load() keeps running in the background.
	tgexec.NewWebExecutor(tgApp).StartService("127.0.0.1:3000")
}
```

The page func must not wait for the load: draw the loading screen and return.
A page blocked on the load keeps its session busy, and the user gets no answer
until it finishes. Store the loaded data with one assignment, as `load()`
does, so a run sees either nothing or all of it.

`RerunAll` returns at once. It never cuts a run in flight, since that would
drop the button click the run is handling: a page in the middle of a run is
rerun once that run ends, and calls made meanwhile fold into that one rerun.
`RerunPage("page1")` reruns only the pages named.

## A reload button

The same call makes a "reload data" button that updates every open page, not
only the one it was pressed on:

```go
func (app *App) Page1(p *tgframe.Params) error {
	if tgcomp.Button(p.Main, "Reload data") {
		go app.load()
	}

	// ...draw from app.snapshot() as above
	return nil
}
```

Load in a goroutine rather than in the page func, so the run that handled the
click ends at once. Pages, the one with the button included, rerun with the new
data when `load()` calls `RerunAll`. To show the loading screen while it runs,
clear `app.data` before starting the load.

Additional Considerations:

- Cache Invalidation: As the application runs, the underlying data sources might change. It's crucial to have a strategy to invalidate cached data when necessary to ensure consistency. This could involve periodically refreshing the cache or implementing mechanisms to detect changes in the source data.

- Memory Usage: App-level caches can consume memory. It's essential to choose appropriate data structures and cache eviction policies to balance performance gains with memory constraints.
