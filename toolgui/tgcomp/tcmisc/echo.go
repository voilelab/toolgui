package tcmisc

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/internal/tcecho"
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
	// 2: this frame, then the caller whose line is being shown.
	tcecho.Echo(c, code, lambda, 2)
}
