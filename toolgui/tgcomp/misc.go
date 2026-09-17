package tgcomp

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/internal/tcecho"
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcmisc"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// Echo will execute lambda and show the code in the lambda.
// To use Echo, we need to store the code in advance (usually by embedded).
//
//	//go:embed main.go
//	var code string
//	// ...
//	// ok, echo will execute and show `tccontent.Text(c, "hello echo")`
//	tcmisc.Echo(c, code, func() {
//		tccontent.Text(c, "hello echo")
//	})
//
//	// panic, since Echo only parse code line by line
//	tcmisc.Echo(c, code, func() {tccontent.Text(c, "hello echo")})
//
//	// panic, since Echo only parse code that start from caller
//	myFunc := func() {
//		tccontent.Text(c, "hello echo")
//	}
//	tcmisc.Echo(c, code, myFunc)
func Echo(c *tgframe.Container, code string, lambda func()) {
	// Echo reads the line it was called from, so it cannot go through
	// tcmisc.Echo: that would read this file. 2 is this frame, then the
	// caller whose line is being shown.
	tcecho.Echo(c, code, lambda, 2)
}

// Message shows a message to the user.
func Message(c *tgframe.Container, text string, conf ...*MessageConf) {
	tcmisc.Message(c, text, conf...)
}

// MessageInfo is a component that displays a message with info color.
func MessageInfo(c *tgframe.Container, text string, conf ...*MessageConf) {
	tcmisc.MessageInfo(c, text, conf...)
}

// MessageSuccess is a component that displays a message with success color.
func MessageSuccess(c *tgframe.Container, text string, conf ...*MessageConf) {
	tcmisc.MessageSuccess(c, text, conf...)
}

// MessageWarning is a component that displays a message with warning color.
func MessageWarning(c *tgframe.Container, text string, conf ...*MessageConf) {
	tcmisc.MessageWarning(c, text, conf...)
}

// MessageDanger is a component that displays a message with danger color.
func MessageDanger(c *tgframe.Container, text string, conf ...*MessageConf) {
	tcmisc.MessageDanger(c, text, conf...)
}

// MessageConf is the configuration for the Message components.
type MessageConf = tcmisc.MessageConf

// ProgressBar shows a progress bar to the user.
func ProgressBar(
	c *tgframe.Container, value int, label string,
	conf ...*ProgressBarConf) *ProgressBarHandle {

	return tcmisc.ProgressBar(c, value, label, conf...)
}

// ProgressBarConf is the configuration for the ProgressBar component.
type ProgressBarConf = tcmisc.ProgressBarConf

// ProgressBarHandle is what ProgressBar hands out: the bar it drew, to be
// moved along while the page function works.
type ProgressBarHandle = tcmisc.ProgressBarHandle

// Iframe is an experimental component, its feature is not stable.
// Use it with caution.
//
// Iframe shows an iframe. IframeConf.Script allows it to run JavaScript.
func Iframe(c *tgframe.Container, html string, conf ...*IframeConf) {
	tcmisc.Iframe(c, html, conf...)
}

// IframeConf is the configuration for the Iframe component.
type IframeConf = tcmisc.IframeConf

// IframeValue returns the latest value an iframe's guest sent through
// window.toolgui.update, nil when it has none. It reads the value rather than
// drawing the iframe -- see [tcmisc.IframeValue].
func IframeValue[T any](
	c *tgframe.Container, html string, conf ...*IframeConf) *T {

	return tcmisc.IframeValue[T](c, html, conf...)
}

// HTML adds an HTML component to the container.
func HTML(c *tgframe.Container, html string, conf ...*HTMLConf) {
	tcmisc.HTML(c, html, conf...)
}

// HTMLConf is the configuration for the HTML component.
type HTMLConf = tcmisc.HTMLConf

// Plugin runs a script in a sandboxed frame and hands it props. The script is
// served from the app, see [tgframe.App.AddPluginAssets].
func Plugin(c *tgframe.Container, src string, conf ...*PluginConf) {
	tcmisc.Plugin(c, src, conf...)
}

// PluginConf is the configuration for the Plugin component.
type PluginConf = tcmisc.PluginConf

// PluginValue returns the latest value a plugin sent through
// window.toolgui.update, nil when it has none. It reads the value rather than
// drawing the plugin -- see [tcmisc.PluginValue].
func PluginValue[T any](
	c *tgframe.Container, src string, conf ...*PluginConf) *T {

	return tcmisc.PluginValue[T](c, src, conf...)
}

// Spinner shows that the page function is busy, and returns the function that
// takes it down again.
func Spinner(
	c *tgframe.Container, label string, conf ...*SpinnerConf) func() {

	return tcmisc.Spinner(c, label, conf...)
}

// SpinnerConf is the configuration for the Spinner component.
type SpinnerConf = tcmisc.SpinnerConf

// Toast shows a one-off notification that takes itself off the screen again.
// It is fired once per run, so every run that reaches the call fires it again.
func Toast(c *tgframe.Container, text string, conf ...*ToastConf) {
	tcmisc.Toast(c, text, conf...)
}

// ToastConf is the configuration for the Toast component.
type ToastConf = tcmisc.ToastConf

// Status reports a piece of work while the page function does it.
func Status(
	c *tgframe.Container, label string, conf ...*StatusConf) *StatusHandle {

	return tcmisc.Status(c, label, conf...)
}

// StatusConf is the configuration for the Status component.
type StatusConf = tcmisc.StatusConf

// StatusHandle is what Status hands out.
type StatusHandle = tcmisc.StatusHandle

// StatusContainer is the old name of StatusHandle.
//
// Deprecated: use [StatusHandle].
type StatusContainer = tcmisc.StatusHandle
