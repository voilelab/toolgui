package demos

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func formDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	var a, b float64
	var ops []int
	opItems := []string{"sum", "product"}
	tgcomp.Form(p.Main, &tgcomp.FormConf{ID: "form"}).With(func(c *tgframe.Container) {
		a = tgcomp.Number[float64](c, "a")
		b = tgcomp.Number[float64](c, "b")
		ops = tgcomp.MultiSelect(c, "ops", opItems,
			&tgcomp.MultiSelectConf{Placeholder: "pick the operations"})
	})

	// Named rather than numbered, so adding an item to opItems cannot
	// silently turn into one of the operations already here.
	for _, op := range ops {
		switch opItems[op] {
		case "sum":
			tgcomp.Text(p.Main,
				fmt.Sprintf("int(a) + int(b) = %d", int(a)+int(b)))
		case "product":
			tgcomp.Text(p.Main,
				fmt.Sprintf("int(a) * int(b) = %d", int(a)*int(b)))
		}
	}
	// ANCHOR_END: demo
	return nil
}

func widgetFormDemo(p *tgframe.Params) error {
	// ANCHOR: widget
	var threshold int64
	var enabled bool

	tgcomp.Form(p.Main, &tgcomp.FormConf{ID: "widget_form"}).
		With(func(c *tgframe.Container) {
			threshold = tgcomp.Slider(c, "threshold",
				(&tcinput.SliderConf[int64]{}).SetMax(100).SetStep(25))
			enabled = tgcomp.Toggle(c, "enabled")
		})

	tgcomp.Text(p.Main,
		fmt.Sprintf("threshold = %d, enabled = %v", threshold, enabled))
	// ANCHOR_END: widget
	return nil
}

func buttonFormDemo(p *tgframe.Params) error {
	// ANCHOR: button
	var keyword string
	var searched bool

	// No submit button of its own: the Search button inside sends the
	// form, so the page offers one button rather than two.
	tgcomp.Form(p.Main, &tgcomp.FormConf{
		ID: "search_form", HideSubmit: true}).
		With(func(c *tgframe.Container) {
			keyword = tgcomp.Textbox(c, "keyword")
			searched = tgcomp.Button(c, "Search")

			// Only a Button sends the form. A download button reports its
			// press the same way, and handing someone a file is not
			// submitting.
			tgcomp.DownloadButton(c, "Save query", []byte("q"),
				&tgcomp.DownloadButtonConf{Filename: "query.txt"})
		})

	if searched {
		tgcomp.Text(p.Main, "Searched: "+keyword)
	} else {
		tgcomp.Text(p.Main, "Not searched yet")
	}
	// ANCHOR_END: button
	return nil
}

func labelFormDemo(p *tgframe.Params) error {
	// ANCHOR: label
	var city string

	tgcomp.Form(p.Main, &tgcomp.FormConf{
		ID: "label_form", SubmitLabel: "Apply"}).
		With(func(c *tgframe.Container) {
			city = tgcomp.Textbox(c, "city")
		})

	tgcomp.Text(p.Main, "Applied: "+city)
	// ANCHOR_END: label
	return nil
}
