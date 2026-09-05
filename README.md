# ToolGUI

[![Go Reference](https://pkg.go.dev/badge/github.com/voilelab/toolgui.svg)](https://pkg.go.dev/github.com/voilelab/toolgui)

This Go package provides a framework for rapidly building interactive data
dashboards and web applications. It aims to offer a similar development
experience to Streamlit for Python users.

> [!WARNING]
> ⚠️ Under Development:
> 
> The API for this package is still under development,
> and may be subject to changes in the future.

## Hello world

```go
package main

import (
	"log"
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgexec"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func Main(p *tgframe.Params) error {
	name := tgcomp.Textbox(p.State, p.Sidebar, "What's your name?")
	if name != "" {
		tgcomp.Text(p.Sidebar, "Hi "+name+"~")
	}

	tgcomp.Text(p.Main, "hello ")
	if tgcomp.Button(p.State, p.Main, "keep going") {
		tgcomp.Text(p.Main, "world")
	}
	return nil
}

func main() {
	app := tgframe.NewApp()
	app.AddPage("index", "Index", Main)
	e := tgexec.NewWebExecutor(app)
	log.Println("Starting service...")
	e.StartService(":3000")
}
```

## For Dev

### Dependency

* [yarn](https://yarnpkg.com/): Frontend
* [cypress](https://www.cypress.io/): E2E Testing
* [taskfile](https://taskfile.dev/): Task runner

### Run demo

```shell
task run_demo
```

### Run E2E Test

```shell
task run_demo
```

```shell
cd toolgui-e2e
cypress e2e:chrome
cypress e2e:firefox
```

### Desktop app

`toolgui-wails` is a separate Go module that runs the same app in a desktop
window with Wails v2. It needs GTK and WebKit, which is why it stays out of the
main module. See [toolgui-wails/README.md](toolgui-wails/README.md).

```shell
task run_wails_hello
```

### Build without the web assets

`toolgui-web/web.go` and `toolgui-wails/assets.go` embed build output that is
not committed on `dev`. To compile Go code without running a frontend build:

```shell
task stub_assets
go build ./...
```

## Release

Releases are cut by the **Release** GitHub Action (`Actions` → `Release` → `Run
workflow`), with a `vX.Y.Z` version as input. It builds the web assets from
`dev`, commits them, tags, and pushes.

`go get` resolves tags, so the tagged commit is what has to carry
`toolgui-web/app/build` — that is the only reason the built assets live in git
at all. `main` is a CI-owned mirror of the latest release commit; do not commit
to it by hand.
