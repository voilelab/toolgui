//go:build js && wasm

// Command hello is the wasm example: a ToolGUI app that runs entirely in the
// browser. Build it with `task build_wasm_hello`.
package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgwasm"
)

func Index(p *tgframe.Params) error {
	name := tgcomp.Textbox(p.Sidebar, "What's your name?")
	if name != "" {
		tgcomp.Text(p.Sidebar, "Hi "+name+"~")
	}

	tgcomp.Title(p.Main, "Hello, WebAssembly")
	tgcomp.Text(p.Main, "This page function is Go, compiled to wasm and run "+
		"by your browser. There is no server behind it.")

	if tgcomp.Button(p.Main, "keep going") {
		tgcomp.Text(p.Main, "world")
	}

	return nil
}

func Echo(p *tgframe.Params) error {
	tgcomp.Title(p.Main, "Echo")

	text := tgcomp.Textarea(p.Main, "Say something")
	tgcomp.Text(p.Main, strings.ToUpper(text))

	return nil
}

// Upload shows the browser-local data story: the file is read by the page
// function without ever leaving the tab.
func Upload(p *tgframe.Params) error {
	tgcomp.Title(p.Main, "Count a file")

	file := tcinput.Fileupload(p.Main, "Pick a text file", "")
	if file == nil {
		return nil
	}

	fp, err := file.Open()
	if err != nil {
		return err
	}
	defer fp.Close()

	// Counting through the reader is what the other transports do off disk;
	// in the tab it reads the bytes the upload already put in memory.
	lines := 1
	buf := make([]byte, 32*1024)
	for {
		n, err := fp.Read(buf)
		lines += bytes.Count(buf[:n], []byte("\n"))

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}
	}

	tgcomp.Text(p.Main, fmt.Sprintf("%s: %d lines, %d bytes",
		file.Name, lines, file.Size))

	return nil
}

func main() {
	app := tgframe.NewApp()
	app.AddPage("index", "Index", Index)
	app.AddPage("echo", "Echo", Echo)
	app.AddPage("upload", "Upload", Upload)

	// A static host cannot route paths, so pages live in the hash.
	app.SetHashPageNameMode(true)

	tgwasm.NewExecutor(app).Run()
}
