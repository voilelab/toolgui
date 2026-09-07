package tgcomp

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// Button create a button and return true if it's clicked.
var Button = tcinput.Button

// ButtonConf store optional conf for Button
type ButtonConf = tcinput.ButtonConf

// DownloadButton create a download button component.
var DownloadButton = tcinput.DownloadButton

type DownloadButtonConf = tcinput.DownloadButtonConf

// DownloadButtonWithConf create a download button component with a user specific configuration.
var DownloadButtonWithConf = tcinput.DownloadButtonWithConf

// Checkbox create a checkbox and return true if it's clicked.
var Checkbox = tcinput.Checkbox

// CheckboxConf store optional conf for Checkbox
type CheckboxConf = tcinput.CheckboxConf

// CheckboxWithConf create a checkbox and return true if it's clicked.
var CheckboxWithConf = tcinput.CheckboxWithConf

// Datepicker create a datepicker and return its selected date.
var Datepicker = tcinput.Datepicker

// Timepicker create a timepicker and return its selected time.
var Timepicker = tcinput.Timepicker

// Datetimepicker create a datetimepicker and return its selected datetime.
var Datetimepicker = tcinput.Datetimepicker

// Fileupload create a fileupload and return its selected file.
var Fileupload = tcinput.Fileupload

// Radio create a group of radio items and return its selected value.
var Radio = tcinput.Radio

// Select create a select dropdown list and return its selected value.
var Select = tcinput.Select

// SelectConf store optional conf for Select
type SelectConf = tcinput.SelectConf

// Textarea create a textarea and return its value.
var Textarea = tcinput.Textarea

// TextareaConf store optional conf for Textarea
type TextareaConf = tcinput.TextareaConf

// TextareaWithConf create a textarea and return its value.
var TextareaWithConf = tcinput.TextareaWithConf

// Textbox create a textbox and return its value.
var Textbox = tcinput.Textbox

// TextboxConf store optional conf for Textbox
type TextboxConf = tcinput.TextboxConf

// Numeric is the value type a [Number] can hold.
type Numeric = tcinput.Numeric

// NumberConf store optional conf for Number
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
