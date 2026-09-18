package tgcomp

import (
	"time"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// Button create a button and return true if it's clicked.
func Button(c *tgframe.Container, label string, conf ...*ButtonConf) bool {
	return tcinput.Button(c, label, conf...)
}

// ButtonConf is the configuration for the Button component.
type ButtonConf = tcinput.ButtonConf

// ButtonClicked reports whether the click this run is handling is the one on
// the button the same label and conf would draw, and can be asked before the
// button is drawn.
func ButtonClicked(c *tgframe.Container, label string, conf ...*ButtonConf) bool {
	return tcinput.ButtonClicked(c, label, conf...)
}

// DownloadButton create a download button component.
func DownloadButton(
	c *tgframe.Container, text string, body []byte,
	conf ...*DownloadButtonConf) bool {

	return tcinput.DownloadButton(c, text, body, conf...)
}

// DownloadButtonConf is the configuration for the DownloadButton component.
type DownloadButtonConf = tcinput.DownloadButtonConf

// DownloadButtonClicked reports whether the click this run is handling is the
// one on the download button the same text and conf would draw, and can be
// asked before the button is drawn. It needs no body.
func DownloadButtonClicked(
	c *tgframe.Container, text string,
	conf ...*DownloadButtonConf) bool {

	return tcinput.DownloadButtonClicked(c, text, conf...)
}

// DownloadFile create a button that hands the app user a file to download,
// fetched by token rather than carried in the pack.
func DownloadFile(
	c *tgframe.Container, text string, body []byte,
	conf ...*DownloadFileConf) bool {

	return tcinput.DownloadFile(c, text, body, conf...)
}

// DownloadFileConf is the configuration for the DownloadFile component.
type DownloadFileConf = tcinput.DownloadFileConf

// DownloadFileClicked reports whether the click this run is handling is the
// one on the download file button the same text and conf would draw, and can
// be asked before the button is drawn. It needs no body.
func DownloadFileClicked(
	c *tgframe.Container, text string,
	conf ...*DownloadFileConf) bool {

	return tcinput.DownloadFileClicked(c, text, conf...)
}

// Checkbox create a checkbox and return true if it's checked.
func Checkbox(c *tgframe.Container, label string, conf ...*CheckboxConf) bool {
	return tcinput.Checkbox(c, label, conf...)
}

// CheckboxConf is the configuration for the Checkbox component.
type CheckboxConf = tcinput.CheckboxConf

// Toggle create a switch and return true if it's on.
func Toggle(c *tgframe.Container, label string, conf ...*ToggleConf) bool {
	return tcinput.Toggle(c, label, conf...)
}

// ToggleConf is the configuration for the Toggle component.
type ToggleConf = tcinput.ToggleConf

// ColorPicker create a color picker and return the picked color.
func ColorPicker(
	c *tgframe.Container, label string, conf ...*ColorPickerConf) string {

	return tcinput.ColorPicker(c, label, conf...)
}

// ColorPickerConf is the configuration for the ColorPicker component.
type ColorPickerConf = tcinput.ColorPickerConf

// DatePicker create a datepicker and return its selected date.
func DatePicker(
	c *tgframe.Container, label string, conf ...*DatePickerConf) *time.Time {

	return tcinput.DatePicker(c, label, conf...)
}

// DatePickerConf is the configuration for the DatePicker component.
type DatePickerConf = tcinput.DatePickerConf

// TimePicker create a timepicker and return its selected time.
func TimePicker(
	c *tgframe.Container, label string, conf ...*TimePickerConf) *time.Time {

	return tcinput.TimePicker(c, label, conf...)
}

// TimePickerConf is the configuration for the TimePicker component.
type TimePickerConf = tcinput.TimePickerConf

// DateTimePicker create a datetimepicker and return its selected datetime.
func DateTimePicker(
	c *tgframe.Container, label string,
	conf ...*DateTimePickerConf) *time.Time {

	return tcinput.DateTimePicker(c, label, conf...)
}

// DateTimePickerConf is the configuration for the DateTimePicker component.
type DateTimePickerConf = tcinput.DateTimePickerConf

// FileUpload create a fileupload and return its selected file.
func FileUpload(
	c *tgframe.Container, label, accept string,
	conf ...*FileUploadConf) *FileObject {

	return tcinput.FileUpload(c, label, accept, conf...)
}

// FileUploadConf is the configuration for the FileUpload component.
type FileUploadConf = tcinput.FileUploadConf

// FileObject is what FileUpload hands back: the file the app user picked.
type FileObject = tcinput.FileObject

// Radio create a group of radio items and return its selected value.
func Radio(
	c *tgframe.Container, label string, items []string,
	conf ...*RadioConf) *int {

	return tcinput.Radio(c, label, items, conf...)
}

// RadioConf is the configuration for the Radio component.
type RadioConf = tcinput.RadioConf

// Select create a select dropdown list and return its selected value.
func Select(
	c *tgframe.Container, label string, items []string,
	conf ...*SelectConf) *int {

	return tcinput.Select(c, label, items, conf...)
}

// SelectConf is the configuration for the Select component.
type SelectConf = tcinput.SelectConf

// MultiSelect create a dropdown list that takes more than one item and return
// the indices of the selected ones.
func MultiSelect(
	c *tgframe.Container, label string, items []string,
	conf ...*MultiSelectConf) []int {

	return tcinput.MultiSelect(c, label, items, conf...)
}

// MultiSelectConf is the configuration for the MultiSelect component.
type MultiSelectConf = tcinput.MultiSelectConf

// Menu create a button with a list of actions behind it, and return the index
// of the item clicked.
func Menu(
	c *tgframe.Container, label string, items []string,
	conf ...*MenuConf) *int {

	return tcinput.Menu(c, label, items, conf...)
}

// MenuConf is the configuration for the Menu component.
type MenuConf = tcinput.MenuConf

// SelectSlider create a slider over a list of items and return the index of
// the selected one.
func SelectSlider(
	c *tgframe.Container, label string, items []string,
	conf ...*SelectSliderConf) int {

	return tcinput.SelectSlider(c, label, items, conf...)
}

// SelectSliderConf is the configuration for the SelectSlider component.
type SelectSliderConf = tcinput.SelectSliderConf

// Textarea create a textarea and return its value.
func Textarea(
	c *tgframe.Container, label string, conf ...*TextareaConf) string {

	return tcinput.Textarea(c, label, conf...)
}

// TextareaConf is the configuration for the Textarea component.
type TextareaConf = tcinput.TextareaConf

// Textbox create a textbox and return its value.
func Textbox(
	c *tgframe.Container, label string, conf ...*TextboxConf) string {

	return tcinput.Textbox(c, label, conf...)
}

// TextboxConf is the configuration for the Textbox component.
type TextboxConf = tcinput.TextboxConf

// Numeric is the value type a [Number] can hold.
type Numeric = tcinput.Numeric

// NumberConf is the configuration for the Number component.
type NumberConf[T tcinput.Numeric] = tcinput.NumberConf[T]

// Number create a number input and return its value, always within Conf.Min
// and Conf.Max, and whether that value is the one the app user entered -- see
// [tcinput.Number].
func Number[T tcinput.Numeric](
	c *tgframe.Container, label string, conf ...*NumberConf[T]) (T, bool) {

	return tcinput.Number[T](c, label, conf...)
}

// SliderConf is the configuration for the Slider component.
type SliderConf[T tcinput.Numeric] = tcinput.SliderConf[T]

// Slider create a slider over a numeric range and return its value.
func Slider[T tcinput.Numeric](
	c *tgframe.Container, label string, conf ...*SliderConf[T]) T {

	return tcinput.Slider[T](c, label, conf...)
}

// Form create a form component.
func Form(c *tgframe.Container, conf ...*FormConf) *tgframe.Container {
	return tcinput.Form(c, conf...)
}

// FormConf is the configuration for the Form component.
type FormConf = tcinput.FormConf
