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
