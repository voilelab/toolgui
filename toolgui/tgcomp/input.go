package tgcomp

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
)

// Button create a button and return true if it's clicked.
var Button = tcinput.Button

// ButtonConf store optional conf for Button
type ButtonConf = tcinput.ButtonConf

// ButtonWithConf create a button and return true if it's clicked.
var ButtonWithConf = tcinput.ButtonWithConf

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

// TextboxWithConf create a textbox and return its value.
var TextboxWithConf = tcinput.TextboxWithConf

// Number create a number input and return its value.
var NumberFloat64 = tcinput.Number[float64]

// NumberInt64 create a number input and return its value.
var NumberInt64 = tcinput.Number[int64]

// NumberWithConf create a number input and return its value.
var NumberWithConfFloat64 = tcinput.NumberWithConf[float64]

// NumberWithConfInt64 create a number input and return its value.
var NumberWithConfInt64 = tcinput.NumberWithConf[int64]

// Form create a form component.
var Form = tcinput.Form
