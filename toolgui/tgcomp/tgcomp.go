package tgcomp

//go:generate go run ./internal/gen

// The rest of this package is generated from the tc* declarations marked
// //tgcomp:export. These two are not plain forwards.

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/internal/tcecho"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// Base is embedded in every component conf, and is where a conf's ID comes
// from. A third-party component's conf embeds it the same way the built-in
// ones do.
type Base = tgframe.Base

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
