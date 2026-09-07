package main

import (
	"archive/zip"
	"crypto/md5"
	"embed"
	"errors"
	"fmt"
	"image/jpeg"
	"io"
	"io/fs"
	"log/slog"
	"strings"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

//go:embed main.go
var code string

// The demo plugin ships as the files it is made of, served under
// /plugin/colorpicker/.
//
//go:embed plugins/colorpicker
var colorPickerAssets embed.FS

var pickerColors = []string{"#ff3860", "#ffdd57", "#23d160", "#3273dc", "#b86bff"}

const readme = `
# [ToolGUI](https://github.com/voilelab/toolgui)

This Go package provides a framework for rapidly building interactive data
dashboards and web applications. It aims to offer a similar development
experience to Streamlit for Python users.

> [!WARNING]
> ⚠️ Under Development:
> 
> The API for this package is still under development,
> and may be subject to changes in the future.

## Example

` + "```go" + `
package main

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgexec"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func main() {
	app := tgframe.NewApp()
	app.AddPage("index", "Index", func(p *tgframe.Params) error {
		tgcomp.Text(p.Main, "Hello world", nil)
		return nil
	})

	tgexec.NewWebExecutor(app).StartService(":3001")
}
` + "```"

func SourceCodePage(p *tgframe.Params) error {
	tgcomp.Title(p.Main, "Example for ToolGUI", nil)
	tgcomp.Code(p.Main, code)
	return nil
}

func MainPage(p *tgframe.Params) error {
	tgcomp.Markdown(p.Main, readme)
	return nil
}

func SidebarPage(p *tgframe.Params) error {
	if tgcomp.Checkbox(p.State, p.Main, "Show sidebar") {
		tgcomp.Text(p.Sidebar, "Sidebar is here", nil)
	}

	tgcomp.Code(p.Main, code)
	return nil
}

func ContentPage(p *tgframe.Params) error {
	headerCompCol, headerCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "header_of_rows"})
	tgcomp.Subtitle(headerCompCol, "Component")
	tgcomp.Subtitle(headerCodeCol, "Code")

	titleCompCol, titleCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_title"})
	tgcomp.Echo(titleCodeCol, code, func() {
		tgcomp.Title(titleCompCol, "Title", nil)
	})

	tgcomp.Divider(p.Main)

	subtitleCompCol, subtitleCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_subtitle"})
	tgcomp.Echo(subtitleCodeCol, code, func() {
		tgcomp.Subtitle(subtitleCompCol, "Subtitle")
	})

	tgcomp.Divider(p.Main)

	textCompCol, textCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_text"})
	tgcomp.Echo(textCodeCol, code, func() {
		tgcomp.Text(textCompCol, "Text", nil)
	})

	tgcomp.Divider(p.Main)

	imageCompCol, imageCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_image"})
	tgcomp.Echo(imageCodeCol, code, func() {
		tgcomp.ImageWithConf(imageCompCol, "https://http.cat/100",
			&tgcomp.ImageConf{
				Width: "200px",
			})
	})

	tgcomp.Divider(p.Main)

	dividerCompCol, dividerCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_divier"})
	tgcomp.Echo(dividerCodeCol, code, func() {
		tgcomp.Divider(dividerCompCol)
	})

	tgcomp.Divider(p.Main)

	linkCompCol, linkCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_link"})
	tgcomp.Echo(linkCodeCol, code, func() {
		tgcomp.Link(linkCompCol, "Link", "https://www.example.com/")
	})

	tgcomp.Divider(p.Main)

	latexCompCol, latexCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_latex"})
	tgcomp.Echo(latexCodeCol, code, func() {
		tgcomp.Latex(latexCompCol, "E = mc^2")
	})

	tgcomp.Divider(p.Main)

	// A shortcode expands wherever text is decoration, and stays literal
	// wherever it is the thing being shown.
	emojiCompCol, emojiCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_emoji"})
	tgcomp.Echo(emojiCodeCol, code, func() {
		tgcomp.Text(emojiCompCol, "Shipped it :tada:", nil)
		tgcomp.Markdown(emojiCompCol, "A `:tada:` in code stays as written.")
	})

	tgcomp.Divider(p.Main)

	// Components are placed by position, so writing the same thing twice
	// shows it twice.
	dupCompCol, dupCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_duplicate"})
	tgcomp.Echo(dupCodeCol, code, func() {
		tgcomp.Text(dupCompCol, "written twice", nil)
		tgcomp.Text(dupCompCol, "written twice", nil)
	})

	return nil
}

func DataPage(p *tgframe.Params) error {
	headerCompCol, headerCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "header_of_rows"})
	tgcomp.Subtitle(headerCompCol, "Component")
	tgcomp.Subtitle(headerCodeCol, "Code")

	jsonCompCol, jsonCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_json"})

	tgcomp.Echo(jsonCodeCol, code, func() {
		type DemoJSONHeader struct {
			Type int
		}

		type DemoJSON struct {
			Header   DemoJSONHeader
			IntValue int
			URL      string
			IsOk     bool
		}

		tgcomp.JSON(jsonCompCol, &DemoJSON{})
	})

	tgcomp.Divider(p.Main)

	tableCompCol, tableCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_table"})
	tgcomp.Echo(tableCodeCol, code, func() {
		tgcomp.Table(tableCompCol, []string{"a", "b"},
			[][]string{{"1", "2"}, {"3", "4"}})
	})

	tgcomp.Divider(p.Main)

	lineCompCol, lineCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_line_chart"})
	tgcomp.Echo(lineCodeCol, code, func() {
		tgcomp.LineChart(lineCompCol, "demo_line",
			[]string{"Mon", "Tue", "Wed", "Thu", "Fri"},
			[]tgcomp.ChartSeries{
				{Name: "visits", Values: []float64{12, 19, 9, 24, 17}},
				{Name: "signups", Values: []float64{3, 7, 4, 9, 6}},
			})
	})

	tgcomp.Divider(p.Main)

	barCompCol, barCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_bar_chart"})
	tgcomp.Echo(barCodeCol, code, func() {
		tgcomp.BarChart(barCompCol, "demo_bar",
			[]string{"Go", "Rust", "Python"},
			[]tgcomp.ChartSeries{
				{Name: "stars", Values: []float64{31, 24, 47}},
			})
	})

	tgcomp.Divider(p.Main)

	areaCompCol, areaCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_area_chart"})
	tgcomp.Echo(areaCodeCol, code, func() {
		tgcomp.ChartWithConf(areaCompCol, "demo_area", &tgcomp.ChartConf{
			Kind:   tgcomp.ChartKindArea,
			Labels: []string{"Q1", "Q2", "Q3", "Q4"},
			Series: []tgcomp.ChartSeries{
				{Name: "cloud", Values: []float64{4, 6, 5, 9}},
				{Name: "desktop", Values: []float64{2, 3, 4, 4}},
			},
			Stacked: true,
			YLabel:  "revenue",
		})
	})

	return nil
}

func LayoutPage(p *tgframe.Params) error {
	headerCompCol, headerCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "header_of_rows"})
	tgcomp.Subtitle(headerCompCol, "Component")
	tgcomp.Subtitle(headerCodeCol, "Code")

	colCompCol, colCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_col"})
	tgcomp.Echo(colCodeCol, code, func() {
		cols := tgcomp.Column(colCompCol, "cols", 3)
		for i, col := range cols {
			tgcomp.Text(col, fmt.Sprintf("col-%d", i), nil)
		}
	})

	tgcomp.Divider(p.Main)

	boxCompCol, boxCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_box"})
	tgcomp.Echo(boxCodeCol, code, func() {
		box := tgcomp.Box(boxCompCol, "box")
		tgcomp.Text(box, "A box!", nil)
	})

	tgcomp.Divider(p.Main)

	tabCompCol, tabCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_tab"})
	tgcomp.Echo(tabCodeCol, code, func() {
		tab1, tab2 := tgcomp.Tab2(tabCompCol, "tab1", "tab2")
		tgcomp.Text(tab1, "A tab!", nil)
		tgcomp.Text(tab2, "B tab!", nil)
	})

	tgcomp.Divider(p.Main)

	expandCompCol, expandCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_expand"})
	tgcomp.Echo(expandCodeCol, code, func() {
		expand := tgcomp.Expand(expandCompCol, "Expand", true)
		tgcomp.Text(expand, "A expand!", nil)
	})

	return nil
}

func InputPage(p *tgframe.Params) error {
	headerCompCol, headerCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "header_of_rows"})
	tgcomp.Subtitle(headerCompCol, "Component")
	tgcomp.Subtitle(headerCodeCol, "Code")

	textareaCompCol, textareaCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_textarea"})
	tgcomp.Echo(textareaCodeCol, code, func() {
		textareaValue := tgcomp.TextareaWithConf(
			p.State, textareaCompCol, "Textarea",
			&tgcomp.TextareaConf{
				Height: 5,
				Color:  tcutil.ColorWarning,
			})
		tgcomp.Text(textareaCompCol, "Value: "+textareaValue, &tgcomp.TextConf{ID: "textarea_result"})
	})

	tgcomp.DividerWithID(p.Main, "1")

	textboxCompCol, textboxCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_textbox"})
	tgcomp.Echo(textboxCodeCol, code, func() {
		textboxValue := tgcomp.Textbox(textboxCompCol, "Textbox", &tgcomp.TextboxConf{
			Placeholder: "input the value here",
			Color:       tcutil.ColorInfo,
		})
		tgcomp.Text(textboxCompCol, "Value: "+textboxValue, &tgcomp.TextConf{ID: "textbox_result"})
	})

	tgcomp.DividerWithID(p.Main, "2")

	fileuploadCompCol, fileuploadCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_fileupload"})
	tgcomp.Echo(fileuploadCodeCol, code, func() {
		fileObj := tgcomp.Fileupload(p.State, fileuploadCompCol,
			"Fileupload", ".jpg,.png")
		if fileObj == nil {
			return
		}

		tgcomp.Text(fileuploadCompCol, "Fileupload filename: "+fileObj.Name, nil)
		tgcomp.Text(fileuploadCompCol, fmt.Sprintf("Fileupload bytes length: %d", fileObj.Size), nil)
		if strings.HasSuffix(fileObj.Name, ".jpg") {
			// Decoding reads the upload off disk, so the image never has to
			// be held twice.
			fp, err := fileObj.Open()
			if err != nil {
				return
			}
			defer fp.Close()

			img, err := jpeg.Decode(fp)
			if err == nil {
				tgcomp.Image(fileuploadCompCol, img)
			}
		}
	})

	tgcomp.DividerWithID(p.Main, "3")

	checkboxCompCol, checkboxCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_checkbox"})
	tgcomp.Echo(checkboxCodeCol, code, func() {
		checkboxValue := tgcomp.Checkbox(p.State, checkboxCompCol, "Checkbox")
		tgcomp.Text(checkboxCompCol, fmt.Sprint("Value: ", checkboxValue), &tgcomp.TextConf{ID: "checkbox_result"})
	})

	tgcomp.DividerWithID(p.Main, "4")

	buttonCompCol, buttonCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_button"})
	tgcomp.Echo(buttonCodeCol, code, func() {
		btnClicked := tgcomp.Button(buttonCompCol, "button", nil)
		tgcomp.Text(buttonCompCol, fmt.Sprint("Value: ", btnClicked), &tgcomp.TextConf{ID: "button_result"})
	})

	tgcomp.DividerWithID(p.Main, "5")

	selectCompCol, selectCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_select"})
	tgcomp.Echo(selectCodeCol, code, func() {
		selIdx := tgcomp.Select(selectCompCol, "Select", []string{"Value1", "Value2"}, nil)

		selItem := ""
		if selIdx != nil {
			selItem = fmt.Sprintf("Value%d", (*selIdx)+1)
		}

		tgcomp.Text(selectCompCol, "Value: "+selItem, &tgcomp.TextConf{ID: "select_result"})
	})

	tgcomp.DividerWithID(p.Main, "6")

	radioCompCol, radioCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_radio"})
	tgcomp.Echo(radioCodeCol, code, func() {
		selIdx := tgcomp.Radio(p.State, radioCompCol,
			"Radio", []string{"Value3", "Value4"})

		selItem := ""
		if selIdx != nil {
			selItem = fmt.Sprintf("Value%d", (*selIdx)+3)
		}

		tgcomp.Text(radioCompCol, "Value: "+selItem, &tgcomp.TextConf{ID: "radio_result"})
	})

	tgcomp.DividerWithID(p.Main, "7")

	datepickerCompCol, datepickerCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_datepicker"})
	tgcomp.Echo(datepickerCodeCol, code, func() {
		dateValue := tgcomp.Datepicker(p.State, datepickerCompCol, "Datepicker")
		val := ""
		if dateValue != nil {
			val = fmt.Sprintf("%04d-%02d-%02d", dateValue.Year, dateValue.Month, dateValue.Day)
		}

		tgcomp.Text(datepickerCompCol, "Value: "+val, &tgcomp.TextConf{ID: "datepicker_result"})
	})

	tgcomp.DividerWithID(p.Main, "8")

	timepickerCompCol, timepickerCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_timepicker"})
	tgcomp.Echo(timepickerCodeCol, code, func() {
		timeValue := tgcomp.Timepicker(p.State, timepickerCompCol, "Timepicker")
		val := ""
		if timeValue != nil {
			val = fmt.Sprintf("%02d:%02d", timeValue.Hour, timeValue.Min)
		}

		tgcomp.Text(timepickerCompCol, "Value: "+val, &tgcomp.TextConf{ID: "timepicker_result"})
	})

	tgcomp.DividerWithID(p.Main, "9")

	datetimepickerCompCol, datetimepickerCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_datetimepicker"})
	tgcomp.Echo(datetimepickerCodeCol, code, func() {
		datetimeValue := tgcomp.Datetimepicker(p.State, datetimepickerCompCol, "Datetimepicker")
		val := ""
		if datetimeValue != nil {
			val = datetimeValue.Format("2006-01-02 15:04")
		}

		tgcomp.Text(datetimepickerCompCol, "Value: "+val, &tgcomp.TextConf{ID: "datetimepicker_result"})
	})

	tgcomp.DividerWithID(p.Main, "10")

	numberCompCol, numberCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_number"})
	tgcomp.Echo(numberCodeCol, code, func() {
		numberValue := tgcomp.Number[float64](numberCompCol, "Number", (&tcinput.NumberConf[float64]{
			Placeholder: "input the value here",
			Color:       tcutil.ColorSuccess,
		}).SetDefault(10).SetMin(10).SetMax(20).SetStep(2))

		valStr := ""
		if numberValue != nil {
			valStr = fmt.Sprint(*numberValue)
		}

		tgcomp.Text(numberCompCol, "Value: "+valStr, &tgcomp.TextConf{ID: "number_result"})
	})

	tgcomp.DividerWithID(p.Main, "11")

	formCompCol, formCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_form"})
	tgcomp.Echo(formCodeCol, code, func() {
		var a, b *float64
		tgcomp.Form(formCompCol, "form").With(func(c *tgframe.Container) {
			a = tgcomp.Number[float64](c, "a", nil)
			b = tgcomp.Number[float64](c, "b", nil)
		})

		if a != nil && b != nil {
			tgcomp.Text(formCompCol, fmt.Sprintf("int(a) + int(b) = %d", int(*a)+int(*b)), nil)
		}
	})

	tgcomp.DividerWithID(p.Main, "12")

	downloadButtonCompCol, downloadButtonCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_download_button"})
	tgcomp.Echo(downloadButtonCodeCol, code, func() {
		if tgcomp.DownloadButtonWithConf(
			p.State, downloadButtonCompCol, "Download", []byte("123"),
			&tgcomp.DownloadButtonConf{
				Filename: "123.txt",
				Color:    tcutil.ColorInfo,
			}) {
			tgcomp.Text(downloadButtonCompCol, "Downloaded!", nil)
		}
	})

	return nil
}

func PluginPage(p *tgframe.Params) error {
	tgcomp.Title(p.Main, "Plugin", nil)
	tgcomp.Text(p.Main, "A plugin is a script the app serves, running in a sandboxed frame.", nil)

	tgcomp.Divider(p.Main)

	pluginCompCol, pluginCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_plugin"})
	tgcomp.Echo(pluginCodeCol, code, func() {
		var value struct {
			Color string `json:"color"`
		}

		// Nothing is selected until the plugin sends its first value, which
		// is not an error to read.
		_ = tgcomp.PluginValue(p.State, "color_picker", &value)

		tgcomp.PluginWithConf(pluginCompCol, "color_picker",
			tgframe.PluginAssetURL("colorpicker", "colorpicker.js"),
			&tgcomp.PluginConf{
				Style: tgframe.PluginAssetURL("colorpicker", "colorpicker.css"),
				Props: map[string]any{
					"colors":   pickerColors,
					"selected": value.Color,
				},
				Height: "auto",
			})

		tgcomp.Text(pluginCompCol, "Selected: "+value.Color, nil)
	})

	return nil
}

func MiscPage(p *tgframe.Params) error {
	headerCompCol, headerCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "header_of_rows"})
	tgcomp.Subtitle(headerCompCol, "Component")
	tgcomp.Subtitle(headerCodeCol, "Code")

	tgcomp.Divider(p.Main)

	echoCompCol, echoCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_echo"})
	tgcomp.Echo(echoCodeCol, code, func() {
		tgcomp.Echo(echoCompCol, code, func() {
			tgcomp.Text(echoCompCol, "hello echo", nil)
		})
	})

	tgcomp.Divider(p.Main)

	msgCompCol, msgCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_msg"})
	tgcomp.Echo(msgCodeCol, code, func() {
		tgcomp.Message(msgCompCol, "body of msg")
	})

	tgcomp.Echo(msgCodeCol, code, func() {
		tgcomp.MessageWithConf(msgCompCol, "body of msg2",
			&tgcomp.MessageConf{
				Title: "danger!",
				Color: tcutil.ColorDanger,
			})
	})

	tgcomp.Divider(p.Main)

	prgbarCompCol, prgbarCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_progress_bar"})
	tgcomp.Echo(prgbarCodeCol, code, func() {
		tgcomp.ProgressBar(prgbarCompCol, 30, "progress_bar")
	})

	tgcomp.Divider(p.Main)

	errorCompCol, errorCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_error"})
	if tgcomp.Button(errorCompCol, "Show error", nil) {
		return errors.New("new error")
	}
	tgcomp.Code(errorCodeCol, `if tgcomp.Button(errorCompCol, "Show error", nil) {
	return errors.New("New error")
}`)

	tgcomp.Divider(p.Main)

	panicCompCol, panicCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_panic"})
	tgcomp.Echo(panicCodeCol, code, func() {
		if tgcomp.Button(panicCompCol, "Show panic", nil) {
			panic("show panic")
		}
	})

	tgcomp.Divider(p.Main)

	iframeSimpleCompCol, iframeSimpleCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_iframe_simple"})
	tgcomp.Echo(iframeSimpleCodeCol, code, func() {
		tgcomp.IframeWithID(
			iframeSimpleCompCol,
			"<b>Hello world gen by html</b>",
			false,
			"iframe_with_simple")
	})

	tgcomp.Divider(p.Main)

	iframeScriptCompCol, iframeScriptCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_iframe_script"})
	tgcomp.Echo(iframeScriptCodeCol, code, func() {
		htmlWithScript := `
		<b id="test">Hello world not changed</b>
		<script>
			const element = document.getElementById('test');
			element.innerText = 'Hello world gen by script';
		</script>`
		tgcomp.IframeWithID(
			iframeScriptCompCol,
			htmlWithScript,
			true,
			"iframe_with_script")
	})

	tgcomp.Divider(p.Main)

	iframeInteractiveCompCol, iframeInteractiveCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_iframe_interactive"})
	tgcomp.Echo(iframeInteractiveCodeCol, code, func() {
		tgcomp.IframeWithConf(
			iframeInteractiveCompCol,
			`<button id="btn">Click me to update</button>
			<script>
				const btn = document.getElementById('btn');
				btn.addEventListener('click', (event) => {
					window.toolgui.update({clicked: true});
				});
			</script>`,
			&tgcomp.IframeConf{
				Script: true,
				Height: "60px",
				ID:     "iframe_with_interactive",
			})

		tgcomp.Text(iframeInteractiveCompCol, time.Now().Format("2006-01-02 15:04:05"), nil)

		var value struct {
			Clicked bool `json:"clicked"`
		}
		err := tgcomp.IframeValue(p.State, "iframe_with_interactive", &value)
		if err != nil {
			return
		}

		tgcomp.Text(iframeInteractiveCompCol, fmt.Sprintf("Status: %v", value.Clicked), nil)
	})

	tgcomp.Divider(p.Main)

	iframeRenderCompCol, iframeRenderCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_iframe_render"})
	tgcomp.Echo(iframeRenderCodeCol, code, func() {
		tgcomp.IframeWithConf(
			iframeRenderCompCol,
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
	})

	tgcomp.Divider(p.Main)

	htmlCompCol, htmlCodeCol := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "show_html"})
	tgcomp.Echo(htmlCodeCol, code, func() {
		tgcomp.Html(htmlCompCol,
			"<b>Hello world gen by html component</b>")
	})

	return nil
}

func getFiles(p *tgframe.Params, f *tcinput.FileObject) ([]string, error) {
	fp, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	// The key covers the content, and hashing it through a stream keeps the
	// file off the heap.
	hash := md5.New()
	if _, err := io.Copy(hash, fp); err != nil {
		return nil, err
	}

	key := fmt.Sprintf("%s_%s_%x", f.Name, f.Type, hash.Sum(nil))

	v := p.State.GetFuncCache(key)
	if v != nil {
		slog.Info("cache found")
		return v.([]string), nil
	}

	// zip reads at an offset, which it can do straight against the file.
	cbzFp, err := zip.NewReader(fp, int64(f.Size))
	if err != nil {
		return nil, err
	}

	ret := []string{}
	for _, f := range cbzFp.File {
		ret = append(ret, f.Name)
	}

	p.State.SetFuncCache(key, ret)
	return ret, nil
}

func FuncCachePage(p *tgframe.Params) error {
	cbzfile := tgcomp.Fileupload(p.State, p.Sidebar, "CBZ File", "application/x-cbz")

	if cbzfile == nil {
		return nil
	}

	files, err := getFiles(p, cbzfile)
	if err != nil {
		return err
	}

	for i, f := range files {
		tgcomp.Text(p.Main, fmt.Sprintf("%d: %s", i, f), nil)
	}

	return nil
}

// newApp build the demo. main lives in main_server.go and main_wasm.go: the
// pages are the same either way, only the executor differs.
func newApp() *tgframe.App {
	app := tgframe.NewApp()

	// The title trails the page title in the browser tab, and names the app
	// in the manifest main_server.go sets unless that gives its own name.
	app.SetTitle("ToolGUI Demo")

	app.AddPage("index", "Index", MainPage)
	app.AddPage("content", "Content", ContentPage)
	app.AddPage("data", "Data", DataPage)
	app.AddPage("input", "Input", InputPage)
	app.AddPage("layout", "Layout", LayoutPage)
	app.AddPage("misc", "Misc", MiscPage)
	app.AddPage("sidebar", "Sidebar", SidebarPage)
	app.AddPage("function_cache", "Function Cache", FuncCachePage)
	app.AddPage("code", "Source Code", SourceCodePage)

	return app
}

// addPluginDemo adds the plugin page and the files its plugin is made of.
// It's the server build's to call: a plugin is loaded over a url, and the
// browser build has no executor serving one.
func addPluginDemo(app *tgframe.App) error {
	assets, err := fs.Sub(colorPickerAssets, "plugins/colorpicker")
	if err != nil {
		return err
	}

	if err := app.AddPluginAssets("colorpicker", assets); err != nil {
		return err
	}

	app.AddPage("plugin", "Plugin", PluginPage)
	return nil
}
