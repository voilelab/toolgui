// Command hello is a minimal ToolGUI desktop app. The wails CLI builds it,
// driven by the wails.json next to this file:
//
//	task run_wails_hello
package main

import (
	"log"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"

	tgwails "github.com/voilelab/toolgui/toolgui-wails"
)

// greetingKey is where the page keeps what the menu last asked for. The menu
// belongs to the app and the state to the session, so a pick made on one run
// is still there on the next.
const greetingKey = "hello_greeting"

func Main(p *tgframe.Params) error {
	name := tgcomp.Textbox(p.Sidebar, "What's your name?")
	if name != "" {
		tgcomp.Text(p.Sidebar, "Hi "+name+"~")
	}

	// The window's own menubar sends the same click the web menubar does, so
	// this reads exactly as it would in a browser.
	switch {
	case tgframe.MenuClicked(p, "greet_hello"):
		p.State.Set(greetingKey, "hello")
	case tgframe.MenuClicked(p, "greet_howdy"):
		p.State.Set(greetingKey, "howdy")
	}

	greeting, ok := p.State.Get[string](greetingKey)
	if !ok {
		greeting = "hello"
	}

	tgcomp.Text(p.Main, greeting+" ")
	if tgcomp.Button(p.Main, "keep going") {
		tgcomp.Text(p.Main, "world")
	}

	return nil
}

func main() {
	app := tgframe.NewApp()
	app.AddPage("index", "Index", Main)

	// On the desktop this becomes the window's native menubar; the same
	// declaration is a row above the app in a browser.
	app.SetMenu(tgframe.NewMenu().
		Submenu("Greeting", func(m *tgframe.Menu) {
			m.Text("Hello", "greet_hello")
			m.Text("Howdy", "greet_howdy")
		}))

	e := tgwails.NewExecutor(app, &tgwails.Conf{Title: "ToolGUI Hello"})
	err := e.Run()
	if err != nil {
		log.Fatal(err)
	}
}
