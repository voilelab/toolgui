// Command hello is a minimal ToolGUI desktop app.
//
//	task asset_wails
//	go run ./example/hello
package main

import (
	"log"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"

	tgwails "github.com/voilelab/toolgui/toolgui-wails"
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

	e := tgwails.NewExecutor(app, &tgwails.Conf{Title: "ToolGUI Hello"})
	err := e.Run()
	if err != nil {
		log.Fatal(err)
	}
}
