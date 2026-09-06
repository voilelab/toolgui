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

// MessageConf is the configuration for the Message component.
type MessageConf = tcmisc.MessageConf

// MessageWithConf shows a message to the user with a specific configuration.
var MessageWithConf = tcmisc.MessageWithConf

// ProgressBar shows a progress bar to the user.
var ProgressBar = tcmisc.ProgressBar

// Iframe and IframeWithID is experimental component, their feature is not stable.
// Use them with caution.

// Iframe show a iframe.
// script is used to allow the iframe to run javascript. (notice that this is not secure)
var Iframe = tcmisc.Iframe

// IframeWithID create a iframe component with a user specific id.
// script is used to allow the iframe to run javascript. (notice that this is not secure)
var IframeWithID = tcmisc.IframeWithID

// Html adds a html component to the container.
var Html = tcmisc.Html

// HtmlWithID adds a html component to the container with a specific id.
var HtmlWithID = tcmisc.HtmlWithID
