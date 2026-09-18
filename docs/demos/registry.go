package demos

// groups is every component that has an example, in the order the coarse
// pages have always shown them. The demo app reads it twice: once for the
// coarse pages, once to give every component a page of its own.
var groups = []*Group{
	{
		Name: "content", Title: "Content",
		Demos: []*Demo{
			demo("title", "Title", show("show_title", titleDemo, "demo")),
			demo("subtitle", "Subtitle",
				show("show_subtitle", subtitleDemo, "demo")),
			demo("text", "Text",
				show("show_text", textDemo, "demo"),
				show("show_duplicate", textDuplicateDemo, "duplicate")),
			demo("caption", "Caption", show("show_caption", captionDemo, "demo")),
			demo("metric", "Metric", show("show_metric", metricDemo, "demo")),
			demo("badge", "Badge", show("show_badge", badgeDemo, "demo")),
			demo("image", "Image", show("show_image", imageDemo, "demo")),
			// The column id keeps its typo: it is what the end to end tests
			// have always named this row.
			demo("divider", "Divider", show("show_divier", dividerDemo, "demo")),
			demo("link", "Link", show("show_link", linkDemo, "demo")),
			demo("link_button", "Link Button",
				show("show_link_button", linkButtonDemo, "demo")),
			demo("latex", "Latex", show("show_latex", latexDemo, "demo")),
			demo("emoji", "Emoji", show("show_emoji", emojiDemo, "text", "markdown")),
		},
	},
	{
		Name: "data", Title: "Data",
		Demos: []*Demo{
			demo("json", "JSON", show("show_json", jsonDemo, "demo")),
			demo("table", "Table", show("show_table", tableDemo, "demo")),
			demo("dataframe", "DataFrame",
				show("show_dataframe", dataFrameDemo, "demo"),
				show("show_dataframe_multi", dataFrameMultiDemo, "multi"),
				show("show_dataframe_single", dataFrameSingleDemo, "single")),
			demo("chart", "Chart",
				show("show_line_chart", lineChartDemo, "line"),
				show("show_bar_chart", barChartDemo, "bar"),
				show("show_area_chart", areaChartDemo, "area")),
			demo("scatter_chart", "Scatter Chart",
				show("show_scatter_chart", scatterChartDemo, "demo")),
		},
	},
	{
		Name: "input", Title: "Input", NumberDividers: true,
		Demos: []*Demo{
			demo("textarea", "Textarea",
				show("show_textarea", textareaDemo, "demo")),
			demo("textbox", "Textbox", show("show_textbox", textboxDemo, "demo")),
			demo("fileupload", "FileUpload",
				show("show_fileupload", fileuploadDemo, "demo")),
			demo("checkbox", "Checkbox",
				show("show_checkbox", checkboxDemo, "demo")),
			demo("button", "Button", show("show_button", buttonDemo, "demo")),
			demo("menu", "Menu", show("show_menu", menuDemo, "demo")),
			demo("select", "Select", show("show_select", selectDemo, "demo")),
			demo("multiselect", "MultiSelect",
				show("show_multiselect", multiSelectDemo, "demo")),
			demo("radio", "Radio", show("show_radio", radioDemo, "demo")),
			demo("datepicker", "DatePicker",
				show("show_datepicker", datePickerDemo, "demo")),
			demo("timepicker", "TimePicker",
				show("show_timepicker", timePickerDemo, "demo")),
			demo("datetimepicker", "DateTimePicker",
				show("show_datetimepicker", dateTimePickerDemo, "demo")),
			demo("number_input", "Number Input",
				show("show_number", numberDemo, "demo")),
			demo("form", "Form",
				show("show_form", formDemo, "demo"),
				show("show_widget_form", widgetFormDemo, "widget"),
				show("show_button_form", buttonFormDemo, "button"),
				show("show_label_form", labelFormDemo, "label")),
			demo("download_button", "Download Button",
				show("show_download_button", downloadButtonDemo, "demo")),
			demo("download_file", "Download File",
				show("show_download_file", downloadFileDemo, "demo")),
			demo("slider", "Slider", show("show_slider", sliderDemo, "demo")),
			demo("select_slider", "Select Slider",
				show("show_select_slider", selectSliderDemo, "demo")),
			demo("toggle", "Toggle", show("show_toggle", toggleDemo, "demo")),
			demo("color_picker", "Color Picker",
				show("show_color_picker", colorPickerDemo, "demo")),
		},
	},
	{
		Name: "layout", Title: "Layout",
		Demos: []*Demo{
			demo("column", "Column", show("show_col", columnDemo, "demo")),
			demo("box", "Box", show("show_box", boxDemo, "demo")),
			demo("tab", "Tab", show("show_tab", tabDemo, "demo")),
			demo("expand", "Expand", show("show_expand", expandDemo, "demo")),
			demo("popover", "Popover", show("show_popover", popoverDemo, "demo")),
			demo("dialog", "Dialog", show("show_dialog", dialogDemo, "demo")).
				withSidebar(dialogSidebarDemo),
			demo("empty", "Empty", show("show_empty", emptyDemo, "demo")),
		},
	},
	{
		Name: "misc", Title: "Misc", LeadDivider: true,
		Demos: []*Demo{
			demo("echo", "Echo", show("show_echo", echoDemo, "demo")),
			demo("message", "Message", show("show_msg", messageDemo, "demo")),
			demo("progress_bar", "Progress Bar",
				show("show_progress_bar", progressBarDemo, "demo")),
			demo("spinner", "Spinner", show("show_spinner", spinnerDemo, "demo")),
			demo("status", "Status", show("show_status", statusDemo, "demo")),
			demo("toast", "Toast", show("show_toast", toastDemo, "demo")),
			demo("error", "Error",
				show("show_error", errorDemo, "demo"),
				show("show_panic", panicDemo, "panic")),
			demo("iframe", "Iframe",
				show("show_iframe_simple", iframeSimpleDemo, "simple"),
				show("show_iframe_script", iframeScriptDemo, "script"),
				show("show_iframe_interactive", iframeInteractiveDemo, "interactive"),
				show("show_iframe_render", iframeRenderDemo, "render")),
			demo("html", "HTML", show("show_html", htmlDemo, "demo")),
		},
	},
}

// pluginComponent is the one component that is not in a group: a plugin is
// loaded over a url, and only an executor serving one can show it. The demo
// app adds it where there is a server.
var pluginComponent = demo("plugin", "Plugin",
	show("show_plugin", pluginDemo, "demo"))

// Groups returns the coarse pages, in nav order.
func Groups() []*Group { return groups }

// Plugin returns the plugin component's demo.
func Plugin() *Demo { return pluginComponent }
