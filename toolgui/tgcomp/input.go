package tgcomp

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// Button create a button and return true if it's clicked.
var Button = tcinput.Button

// ButtonConf is the configuration for the Button component.
type ButtonConf = tcinput.ButtonConf

// DownloadButton create a download button component.
var DownloadButton = tcinput.DownloadButton

// DownloadButtonConf is the configuration for the DownloadButton component.
type DownloadButtonConf = tcinput.DownloadButtonConf

// Checkbox create a checkbox and return true if it's checked.
var Checkbox = tcinput.Checkbox

// CheckboxConf is the configuration for the Checkbox component.
type CheckboxConf = tcinput.CheckboxConf

// Datepicker create a datepicker and return its selected date.
var Datepicker = tcinput.Datepicker

// DatepickerConf is the configuration for the Datepicker component.
type DatepickerConf = tcinput.DatepickerConf

// Timepicker create a timepicker and return its selected time.
var Timepicker = tcinput.Timepicker

// TimepickerConf is the configuration for the Timepicker component.
type TimepickerConf = tcinput.TimepickerConf

// Datetimepicker create a datetimepicker and return its selected datetime.
var Datetimepicker = tcinput.Datetimepicker

// DatetimepickerConf is the configuration for the Datetimepicker component.
type DatetimepickerConf = tcinput.DatetimepickerConf

// Fileupload create a fileupload and return its selected file.
var Fileupload = tcinput.Fileupload

// FileuploadConf is the configuration for the Fileupload component.
type FileuploadConf = tcinput.FileuploadConf

// Radio create a group of radio items and return its selected value.
var Radio = tcinput.Radio

// RadioConf is the configuration for the Radio component.
type RadioConf = tcinput.RadioConf

// Select create a select dropdown list and return its selected value.
var Select = tcinput.Select

// SelectConf is the configuration for the Select component.
type SelectConf = tcinput.SelectConf

// Textarea create a textarea and return its value.
var Textarea = tcinput.Textarea

// TextareaConf is the configuration for the Textarea component.
type TextareaConf = tcinput.TextareaConf

// Textbox create a textbox and return its value.
var Textbox = tcinput.Textbox

// TextboxConf is the configuration for the Textbox component.
type TextboxConf = tcinput.TextboxConf

// Numeric is the value type a [Number] can hold.
type Numeric = tcinput.Numeric

// NumberConf is the configuration for the Number component.
type NumberConf[T tcinput.Numeric] = tcinput.NumberConf[T]

// Number create a number input and return its value.
//
// A generic function cannot be forwarded by a var, so this is a wrapper rather
// than an alias like its neighbours.
func Number[T tcinput.Numeric](
	c *tgframe.Container, label string, conf ...*NumberConf[T]) *T {

	return tcinput.Number[T](c, label, conf...)
}

// Form create a form component.
var Form = tcinput.Form

// FormConf is the configuration for the Form component.
type FormConf = tcinput.FormConf
