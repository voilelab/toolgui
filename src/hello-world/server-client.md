# Server-Client

## Step by step Hello World

1. Create `main.go`:

```go
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

	tgexec.NewWebExecutor(app).StartService(":3001")
}
```

2. Create go.mod and download toolgui:

```bash
go mod init toolgui-helloworld
go mod tidy
```

3. Run helloworld

```bash
go run main.go
```

## Explain

* Create a ToolGUI App: The `App` instance includes the info that app needs.

```go
app := tgframe.NewApp()
```

* Register a **page** in App: Tell `App` instance, we will have a page in the App.
  * `index` is the name.
  * `Index` is the title.

```go
app.AddPage("index", "Index", ...)
```

* The Page Func: Draw a text component in the **Main** container.

```go
func(p *tgframe.Params) error {
	tgcomp.Text(p.Main, "Hello world")
	return nil
}
```

* WebExecutor: The `App` only includes the logic of app, but not the GUI.
The web executor provides a web server GUI interface for `App`.

```go
tgexec.NewWebExecutor(app).StartService(":3001")
```
