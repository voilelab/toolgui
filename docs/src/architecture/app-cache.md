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
loading screen until it is ready:

```go
type App struct {
	mu   sync.RWMutex
	data *Data // nil until loaded
	err  error
}

func NewApp() *App {
	app := &App{}
	go app.load()
	return app
}

func (app *App) load() {
	data, err := loadData() // slow

	app.mu.Lock()
	app.data, app.err = data, err
	app.mu.Unlock()
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
		tgcomp.Button(p.Main, "Refresh")
		return nil
	}

	tgcomp.Text(p.Main, data.Summary())
	return nil
}

func main() {
	app := NewApp()

	tgApp := tgframe.NewApp()
	tgApp.AddPage("page1", "Page1", app.Page1)

	// Serves at once; load() keeps running in the background.
	tgexec.NewWebExecutor(tgApp).StartService("127.0.0.1:3000")
}
```

The page func must not wait for the load: draw the loading screen and return.
A page blocked on the load keeps its session busy, and the user gets no answer
until it finishes.

The page reruns only on a user event, so a page opened before the load
finished keeps showing the loading screen until the user does something. The
`Refresh` button above is that something; the side nav's `Rerun` button does
the same, but a button on the page is easier to find. Store the loaded data with one
assignment, as `load()` does, so a run sees either nothing or all of it.

Additional Considerations:

- Cache Invalidation: As the application runs, the underlying data sources might change. It's crucial to have a strategy to invalidate cached data when necessary to ensure consistency. This could involve periodically refreshing the cache or implementing mechanisms to detect changes in the source data.

- Memory Usage: App-level caches can consume memory. It's essential to choose appropriate data structures and cache eviction policies to balance performance gains with memory constraints.
