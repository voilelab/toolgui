package tgcomp

import "github.com/voilelab/toolgui/toolgui/tgcomp/tcmisc"

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
var Echo = tcmisc.Echo

// Message shows a message to the user.
var Message = tcmisc.Message

// MessageInfo is a component that displays a message with info color.
var MessageInfo = tcmisc.MessageInfo

// MessageSuccess is a component that displays a message with success color.
var MessageSuccess = tcmisc.MessageSuccess

// MessageWarning is a component that displays a message with warning color.
var MessageWarning = tcmisc.MessageWarning

// MessageDanger is a component that displays a message with danger color.
var MessageDanger = tcmisc.MessageDanger

// MessageConf is the configuration for the Message components.
type MessageConf = tcmisc.MessageConf

// ProgressBar shows a progress bar to the user.
var ProgressBar = tcmisc.ProgressBar

// ProgressBarConf is the configuration for the ProgressBar component.
type ProgressBarConf = tcmisc.ProgressBarConf

// Iframe is an experimental component, its feature is not stable.
// Use it with caution.
//
// Iframe show a iframe. IframeConf.Script allows the iframe to run javascript.
var Iframe = tcmisc.Iframe

// IframeConf is the configuration for the Iframe component.
type IframeConf = tcmisc.IframeConf

// IframeValue reads the latest value an iframe sent through window.update.
var IframeValue = tcmisc.IframeValue

// Html adds a html component to the container.
var Html = tcmisc.Html

// HtmlConf is the configuration for the Html component.
type HtmlConf = tcmisc.HtmlConf

// Plugin runs a script in a sandboxed frame and hands it props. The script is
// served from the app, see [tgframe.App.AddPluginAssets].
var Plugin = tcmisc.Plugin

// PluginConf is the configuration for the Plugin component.
type PluginConf = tcmisc.PluginConf

// PluginValue reads the latest value a plugin sent through
// window.toolgui.update.
var PluginValue = tcmisc.PluginValue
