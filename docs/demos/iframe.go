package demos

import (
	"fmt"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// clickedValue is what the interactive iframe sends through
// window.toolgui.update.
//
// ANCHOR: value
type clickedValue struct {
	Clicked bool `json:"clicked"`
}

// ANCHOR_END: value

func iframeSimpleDemo(p *tgframe.Params) error {
	// ANCHOR: simple
	tgcomp.Iframe(
		p.Main,
		"<b>Hello world gen by html</b>",
		&tgcomp.IframeConf{ID: "iframe_with_simple"})
	// ANCHOR_END: simple
	return nil
}

func iframeScriptDemo(p *tgframe.Params) error {
	// ANCHOR: script
	htmlWithScript := `
	<b id="test">Hello world not changed</b>
	<script>
		const element = document.getElementById('test');
		element.innerText = 'Hello world gen by script';
	</script>`
	tgcomp.Iframe(
		p.Main,
		htmlWithScript,
		&tgcomp.IframeConf{Script: true, ID: "iframe_with_script"})
	// ANCHOR_END: script
	return nil
}

func iframeInteractiveDemo(p *tgframe.Params) error {
	// ANCHOR: interactive
	html := `<button id="btn">Click me to update</button>
		<script>
			const btn = document.getElementById('btn');
			btn.addEventListener('click', (event) => {
				window.toolgui.update({clicked: true});
			});
		</script>`
	conf := &tgcomp.IframeConf{
		Script: true,
		Height: "60px",
		ID:     "iframe_with_interactive",
	}

	tgcomp.Iframe(p.Main, html, conf)

	tgcomp.Text(p.Main, time.Now().Format("2006-01-02 15:04:05"))

	// nil until the guest clicks for the first time.
	clicked := false
	if v := tgcomp.IframeValue[clickedValue](p.Main, html, conf); v != nil {
		clicked = v.Clicked
	}

	tgcomp.Text(p.Main, fmt.Sprintf("Status: %v", clicked))
	// ANCHOR_END: interactive
	return nil
}

func iframeRenderDemo(p *tgframe.Params) error {
	// ANCHOR: render
	tgcomp.Iframe(
		p.Main,
		`<div id="out">waiting for render</div>
		<script>
			const out = document.getElementById('out');
			window.toolgui.onRender((props, theme) => {
				out.innerText = 'theme=' + theme + ' id=' + props.id;
			});
			window.toolgui.autoHeight();
		</script>`,
		&tgcomp.IframeConf{
			Script: true,
			Height: "auto",
			ID:     "iframe_with_render",
		})
	// ANCHOR_END: render
	return nil
}
