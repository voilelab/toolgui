package main

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	_ "embed"
	"errors"
	"fmt"
	"image/jpeg"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/VoileLab/toolgui/toolgui/tgcomp"
	"github.com/VoileLab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/VoileLab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/VoileLab/toolgui/toolgui/tgexec"
	"github.com/VoileLab/toolgui/toolgui/tgframe"
)

//go:embed main.go
var code string

const readme = `
# [ToolGUI](https://github.com/VoileLab/toolgui)

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
	"github.com/VoileLab/toolgui/toolgui/tgcomp"
	"github.com/VoileLab/toolgui/toolgui/tgexec"
	"github.com/VoileLab/toolgui/toolgui/tgframe"
)

func main() {
	app := tgframe.NewApp()
	app.AddPage("index", "Index", func(p *tgframe.Params) error {
		tgcomp.Text(p.Main, "Hello world")
		return nil
	})

	tgexec.NewWebExecutor(app).StartService(":3001")
}
` + "```"

func SourceCodePage(p *tgframe.Params) error {
	tgcomp.Title(p.Main, "Example for ToolGUI")
	tgcomp.Code(p.Main, code)
	return nil
}

func MainPage(p *tgframe.Params) error {
	tgcomp.Markdown(p.Main, readme)
	return nil
}

func SidebarPage(p *tgframe.Params) error {
	if tgcomp.Checkbox(p.State, p.Main, "Show sidebar") {
		tgcomp.Text(p.Sidebar, "Sidebar is here")
	}

	tgcomp.Code(p.Main, code)
	return nil
}

func ContentPage(p *tgframe.Params) error {
	headerCompCol, headerCodeCol := tgcomp.EqColumn2(p.Main, "header_of_rows")
	tgcomp.Subtitle(headerCompCol, "Component")
	tgcomp.Subtitle(headerCodeCol, "Code")

	titleCompCol, titleCodeCol := tgcomp.EqColumn2(p.Main, "show_title")
	tgcomp.Echo(titleCodeCol, code, func() {
		tgcomp.Title(titleCompCol, "Title")
	})

	tgcomp.Divider(p.Main)

	subtitleCompCol, subtitleCodeCol := tgcomp.EqColumn2(p.Main, "show_subtitle")
	tgcomp.Echo(subtitleCodeCol, code, func() {
		tgcomp.Subtitle(subtitleCompCol, "Subtitle")
	})

	tgcomp.Divider(p.Main)

	textCompCol, textCodeCol := tgcomp.EqColumn2(p.Main, "show_text")
	tgcomp.Echo(textCodeCol, code, func() {
		tgcomp.Text(textCompCol, "Text")
	})

	tgcomp.Divider(p.Main)

	imageCompCol, imageCodeCol := tgcomp.EqColumn2(p.Main, "show_image")
	tgcomp.Echo(imageCodeCol, code, func() {
		tgcomp.ImageWithConf(imageCompCol, "https://http.cat/100",
			&tgcomp.ImageConf{
				Width: "200px",
			})
	})

	tgcomp.Divider(p.Main)

	dividerCompCol, dividerCodeCol := tgcomp.EqColumn2(p.Main, "show_divier")
	tgcomp.Echo(dividerCodeCol, code, func() {
		tgcomp.Divider(dividerCompCol)
	})

	tgcomp.Divider(p.Main)

	linkCompCol, linkCodeCol := tgcomp.EqColumn2(p.Main, "show_link")
	tgcomp.Echo(linkCodeCol, code, func() {
		tgcomp.Link(linkCompCol, "Link", "https://www.example.com/")
	})

	tgcomp.Divider(p.Main)

	latexCompCol, latexCodeCol := tgcomp.EqColumn2(p.Main, "show_latex")
	tgcomp.Echo(latexCodeCol, code, func() {
		tgcomp.Latex(latexCompCol, "E = mc^2")
	})

	return nil
}

func DataPage(p *tgframe.Params) error {
	headerCompCol, headerCodeCol := tgcomp.EqColumn2(p.Main, "header_of_rows")
	tgcomp.Subtitle(headerCompCol, "Component")
	tgcomp.Subtitle(headerCodeCol, "Code")

	jsonCompCol, jsonCodeCol := tgcomp.EqColumn2(p.Main, "show_json")

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

	tableCompCol, tableCodeCol := tgcomp.EqColumn2(p.Main, "show_table")
	tgcomp.Echo(tableCodeCol, code, func() {
		tgcomp.Table(tableCompCol, []string{"a", "b"},
			[][]string{{"1", "2"}, {"3", "4"}})
	})

	return nil
}

func LayoutPage(p *tgframe.Params) error {
	headerCompCol, headerCodeCol := tgcomp.EqColumn2(p.Main, "header_of_rows")
	tgcomp.Subtitle(headerCompCol, "Component")
	tgcomp.Subtitle(headerCodeCol, "Code")

	colCompCol, colCodeCol := tgcomp.EqColumn2(p.Main, "show_col")
	tgcomp.Echo(colCodeCol, code, func() {
		cols := tgcomp.Column(colCompCol, "cols", 3)
		for i, col := range cols {
			tgcomp.Text(col, fmt.Sprintf("col-%d", i))
		}
	})

	tgcomp.Divider(p.Main)

	boxCompCol, boxCodeCol := tgcomp.EqColumn2(p.Main, "show_box")
	tgcomp.Echo(boxCodeCol, code, func() {
		box := tgcomp.Box(boxCompCol, "box")
		tgcomp.Text(box, "A box!")
	})

	tgcomp.Divider(p.Main)

	tabCompCol, tabCodeCol := tgcomp.EqColumn2(p.Main, "show_tab")
	tgcomp.Echo(tabCodeCol, code, func() {
		tab1, tab2 := tgcomp.Tab2(tabCompCol, "tab1", "tab2")
		tgcomp.Text(tab1, "A tab!")
		tgcomp.Text(tab2, "B tab!")
	})

	tgcomp.Divider(p.Main)

	expandCompCol, expandCodeCol := tgcomp.EqColumn2(p.Main, "show_expand")
	tgcomp.Echo(expandCodeCol, code, func() {
		expand := tgcomp.Expand(expandCompCol, "Expand", true)
		tgcomp.Text(expand, "A expand!")
	})

	return nil
}

func InputPage(p *tgframe.Params) error {
	headerCompCol, headerCodeCol := tgcomp.EqColumn2(p.Main, "header_of_rows")
	tgcomp.Subtitle(headerCompCol, "Component")
	tgcomp.Subtitle(headerCodeCol, "Code")

	textareaCompCol, textareaCodeCol := tgcomp.EqColumn2(p.Main, "show_textarea")
	tgcomp.Echo(textareaCodeCol, code, func() {
		textareaValue := tgcomp.TextareaWithConf(
			p.State, textareaCompCol, "Textarea",
			&tgcomp.TextareaConf{
				Height: 5,
				Color:  tcutil.ColorWarning,
			})
		tgcomp.TextWithID(textareaCompCol,
			"Value: "+textareaValue, "textarea_result")
	})

	tgcomp.DividerWithID(p.Main, "1")

	textboxCompCol, textboxCodeCol := tgcomp.EqColumn2(p.Main, "show_textbox")
	tgcomp.Echo(textboxCodeCol, code, func() {
		textboxValue := tgcomp.TextboxWithConf(p.State, textboxCompCol, "Textbox",
			&tgcomp.TextboxConf{
				Placeholder: "input the value here",
				Color:       tcutil.ColorInfo,
			})
		tgcomp.TextWithID(textboxCompCol, "Value: "+textboxValue, "textbox_result")
	})

	tgcomp.DividerWithID(p.Main, "2")

	fileuploadCompCol, fileuploadCodeCol := tgcomp.EqColumn2(p.Main, "show_fileupload")
	tgcomp.Echo(fileuploadCodeCol, code, func() {
		fileObj := tgcomp.Fileupload(p.State, fileuploadCompCol,
			"Fileupload", ".jpg,.png")
		if fileObj == nil {
			return
		}

		tgcomp.Text(fileuploadCompCol, "Fileupload filename: "+fileObj.Name)
		tgcomp.Text(fileuploadCompCol,
			fmt.Sprintf("Fileupload bytes length: %d", len(fileObj.Bytes)))
		if strings.HasSuffix(fileObj.Name, ".jpg") {
			img, err := jpeg.Decode(bytes.NewReader(fileObj.Bytes))
			if err == nil {
				tgcomp.Image(fileuploadCompCol, img)
			}
		}
	})

	tgcomp.DividerWithID(p.Main, "3")

	checkboxCompCol, checkboxCodeCol := tgcomp.EqColumn2(p.Main, "show_checkbox")
	tgcomp.Echo(checkboxCodeCol, code, func() {
		checkboxValue := tgcomp.Checkbox(p.State, checkboxCompCol, "Checkbox")
		tgcomp.TextWithID(checkboxCompCol,
			fmt.Sprint("Value: ", checkboxValue), "checkbox_result")
	})

	tgcomp.DividerWithID(p.Main, "4")

	buttonCompCol, buttonCodeCol := tgcomp.EqColumn2(p.Main, "show_button")
	tgcomp.Echo(buttonCodeCol, code, func() {
		btnClicked := tgcomp.Button(p.State, buttonCompCol, "button")
		tgcomp.TextWithID(buttonCompCol,
			fmt.Sprint("Value: ", btnClicked), "button_result")
	})

	tgcomp.DividerWithID(p.Main, "5")

	selectCompCol, selectCodeCol := tgcomp.EqColumn2(p.Main, "show_select")
	tgcomp.Echo(selectCodeCol, code, func() {
		selIdx := tgcomp.Select(p.State, selectCompCol,
			"Select", []string{"Value1", "Value2"})

		selItem := ""
		if selIdx != nil {
			selItem = fmt.Sprintf("Value%d", (*selIdx)+1)
		}

		tgcomp.TextWithID(selectCompCol, "Value: "+selItem, "select_result")
	})

	tgcomp.DividerWithID(p.Main, "6")

	radioCompCol, radioCodeCol := tgcomp.EqColumn2(p.Main, "show_radio")
	tgcomp.Echo(radioCodeCol, code, func() {
		selIdx := tgcomp.Radio(p.State, radioCompCol,
			"Radio", []string{"Value3", "Value4"})

		selItem := ""
		if selIdx != nil {
			selItem = fmt.Sprintf("Value%d", (*selIdx)+3)
		}

		tgcomp.TextWithID(radioCompCol, "Value: "+selItem, "radio_result")
	})

	tgcomp.DividerWithID(p.Main, "7")

	datepickerCompCol, datepickerCodeCol := tgcomp.EqColumn2(p.Main, "show_datepicker")
	tgcomp.Echo(datepickerCodeCol, code, func() {
		dateValue := tgcomp.Datepicker(p.State, datepickerCompCol, "Datepicker")
		val := ""
		if dateValue != nil {
			val = fmt.Sprintf("%04d-%02d-%02d", dateValue.Year, dateValue.Month, dateValue.Day)
		}

		tgcomp.TextWithID(datepickerCompCol, "Value: "+val, "datepicker_result")
	})

	tgcomp.DividerWithID(p.Main, "8")

	timepickerCompCol, timepickerCodeCol := tgcomp.EqColumn2(p.Main, "show_timepicker")
	tgcomp.Echo(timepickerCodeCol, code, func() {
		timeValue := tgcomp.Timepicker(p.State, timepickerCompCol, "Timepicker")
		val := ""
		if timeValue != nil {
			val = fmt.Sprintf("%02d:%02d", timeValue.Hour, timeValue.Min)
		}

		tgcomp.TextWithID(timepickerCompCol, "Value: "+val, "timepicker_result")
	})

	tgcomp.DividerWithID(p.Main, "9")

	datetimepickerCompCol, datetimepickerCodeCol := tgcomp.EqColumn2(p.Main, "show_datetimepicker")
	tgcomp.Echo(datetimepickerCodeCol, code, func() {
		datetimeValue := tgcomp.Datetimepicker(p.State, datetimepickerCompCol, "Datetimepicker")
		val := ""
		if datetimeValue != nil {
			val = datetimeValue.Format("2006-01-02 15:04")
		}

		tgcomp.TextWithID(datetimepickerCompCol, "Value: "+val, "datetimepicker_result")
	})

	tgcomp.DividerWithID(p.Main, "10")

	numberCompCol, numberCodeCol := tgcomp.EqColumn2(p.Main, "show_number")
	tgcomp.Echo(numberCodeCol, code, func() {
		numberValue := tgcomp.NumberWithConfFloat64(p.State, numberCompCol, "Number",
			(&tcinput.NumberConf[float64]{
				Placeholder: "input the value here",
				Color:       tcutil.ColorSuccess,
			}).SetDefault(10).SetMin(10).SetMax(20).SetStep(2))

		valStr := ""
		if numberValue != nil {
			valStr = fmt.Sprint(*numberValue)
		}

		tgcomp.TextWithID(numberCompCol, "Value: "+valStr, "number_result")
	})

	tgcomp.DividerWithID(p.Main, "11")

	formCompCol, formCodeCol := tgcomp.EqColumn2(p.Main, "show_form")
	tgcomp.Echo(formCodeCol, code, func() {
		var a, b *float64
		tgcomp.Form(formCompCol, "form").With(func(c *tgframe.Container) {
			a = tgcomp.NumberFloat64(p.State, c, "a")
			b = tgcomp.NumberFloat64(p.State, c, "b")
		})

		if a != nil && b != nil {
			tgcomp.Text(formCompCol, fmt.Sprintf("int(a) + int(b) = %d", int(*a)+int(*b)))
		}
	})

	tgcomp.DividerWithID(p.Main, "12")

	downloadButtonCompCol, downloadButtonCodeCol := tgcomp.EqColumn2(p.Main, "show_download_button")
	tgcomp.Echo(downloadButtonCodeCol, code, func() {
		if tgcomp.DownloadButtonWithConf(
			p.State, downloadButtonCompCol, "Download", []byte("123"),
			&tgcomp.DownloadButtonConf{
				Filename: "123.txt",
				Color:    tcutil.ColorInfo,
			}) {
			tgcomp.Text(downloadButtonCompCol, "Downloaded!")
		}
	})

	return nil
}

func MiscPage(p *tgframe.Params) error {
	headerCompCol, headerCodeCol := tgcomp.EqColumn2(p.Main, "header_of_rows")
	tgcomp.Subtitle(headerCompCol, "Component")
	tgcomp.Subtitle(headerCodeCol, "Code")

	tgcomp.Divider(p.Main)

	echoCompCol, echoCodeCol := tgcomp.EqColumn2(p.Main, "show_echo")
	tgcomp.Echo(echoCodeCol, code, func() {
		tgcomp.Echo(echoCompCol, code, func() {
			tgcomp.Text(echoCompCol, "hello echo")
		})
	})

	tgcomp.Divider(p.Main)

	msgCompCol, msgCodeCol := tgcomp.EqColumn2(p.Main, "show_msg")
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

	prgbarCompCol, prgbarCodeCol := tgcomp.EqColumn2(p.Main, "show_progress_bar")
	tgcomp.Echo(prgbarCodeCol, code, func() {
		tgcomp.ProgressBar(prgbarCompCol, 30, "progress_bar")
	})

	tgcomp.Divider(p.Main)

	errorCompCol, errorCodeCol := tgcomp.EqColumn2(p.Main, "show_error")
	if tgcomp.Button(p.State, errorCompCol, "Show error") {
		return errors.New("new error")
	}
	tgcomp.Code(errorCodeCol, `if tgcomp.Button(p.State, errorCompCol, "Show error") {
	return errors.New("New error")
}`)

	tgcomp.Divider(p.Main)

	panicCompCol, panicCodeCol := tgcomp.EqColumn2(p.Main, "show_panic")
	tgcomp.Echo(panicCodeCol, code, func() {
		if tgcomp.Button(p.State, panicCompCol, "Show panic") {
			panic("show panic")
		}
	})

	tgcomp.Divider(p.Main)

	iframeSimpleCompCol, iframeSimpleCodeCol := tgcomp.EqColumn2(p.Main, "show_iframe_simple")
	tgcomp.Echo(iframeSimpleCodeCol, code, func() {
		tgcomp.IframeWithID(
			iframeSimpleCompCol,
			"<b>Hello world gen by html</b>",
			false,
			"iframe_with_simple")
	})

	tgcomp.Divider(p.Main)

	iframeScriptCompCol, iframeScriptCodeCol := tgcomp.EqColumn2(p.Main, "show_iframe_script")
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

	iframeInteractiveCompCol, iframeInteractiveCodeCol := tgcomp.EqColumn2(p.Main, "show_iframe_interactive")
	tgcomp.Echo(iframeInteractiveCodeCol, code, func() {
		tgcomp.IframeWithID(
			iframeInteractiveCompCol,
			`<button id="btn">Click me to update</button>
			<script>
				const btn = document.getElementById('btn');
				btn.addEventListener('click', (event) => {
					window.update({type: "click", id: "iframe_with_interactive_btn"});
				});
			</script>`,
			true,
			"iframe_with_interactive")

		tgcomp.Text(iframeInteractiveCompCol, time.Now().Format("2006-01-02 15:04:05"))
		clickStatus := p.State.GetClickID() == "iframe_with_interactive_btn"
		tgcomp.Text(iframeInteractiveCompCol, fmt.Sprintf("Status: %v", clickStatus))
	})

	tgcomp.Divider(p.Main)

	htmlCompCol, htmlCodeCol := tgcomp.EqColumn2(p.Main, "show_html")
	tgcomp.Echo(htmlCodeCol, code, func() {
		tgcomp.Html(htmlCompCol,
			"<b>Hello world gen by html component</b>")
	})

	return nil
}

func getFiles(p *tgframe.Params, f *tcinput.FileObject) ([]string, error) {
	key := fmt.Sprintf("%s_%s_%x", f.Name, f.Type, md5.Sum(f.Bytes))

	v := p.State.GetFuncCache(key)
	if v != nil {
		slog.Info("cache found")
		return v.([]string), nil
	}

	buf := bytes.NewReader(f.Bytes)

	cbzFp, err := zip.NewReader(buf, buf.Size())
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
		tgcomp.Text(p.Main, fmt.Sprintf("%d: %s", i, f))
	}

	return nil
}

func main() {
	app := tgframe.NewApp()

	app.AddPageByConfig(&tgframe.PageConfig{
		Name:  "index",
		Title: "Index",
		Emoji: "🏠",
	}, MainPage)

	app.AddPage("content", "Content", ContentPage)
	app.AddPage("data", "Data", DataPage)
	app.AddPage("input", "Input", InputPage)
	app.AddPage("layout", "Layout", LayoutPage)
	app.AddPage("misc", "Misc", MiscPage)
	app.AddPage("sidebar", "Sidebar", SidebarPage)
	app.AddPage("function_cache", "Function Cache", FuncCachePage)
	app.AddPage("code", "Source Code", SourceCodePage)

	e := tgexec.NewWebExecutor(app)
	log.Println("Starting service...")
	err := e.StartService(":3000")
	if err != nil {
		log.Println(err)
	}
}
