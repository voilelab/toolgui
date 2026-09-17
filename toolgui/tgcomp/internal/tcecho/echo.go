// Package tcecho holds the implementation of the Echo component.
//
// Echo reads the source line it was called from, so the frame it looks at
// depends on how many calls deep it sits. It lives here rather than in tcmisc
// so that both entry points -- tcmisc.Echo and the tgcomp one forwarding to it
// -- can name that depth for themselves.
package tcecho

import (
	"fmt"
	"go/parser"
	"runtime"
	"strings"
	"sync"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tccontent"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

type echoCodeCache struct {
	data sync.Map
}

func (cc *echoCodeCache) get(filename string, line int) (string, bool) {
	key := fmt.Sprintf("%s\t%d", filename, line)
	val, ok := cc.data.Load(key)
	if ok {
		return val.(string), true
	}
	return "", false
}

func (cc *echoCodeCache) set(filename string, line int, code string) {
	key := fmt.Sprintf("%s\t%d", filename, line)
	cc.data.Store(key, code)
}

var codeCache echoCodeCache

func countIndent(line string) int {
	cnt := 0
	for _, c := range line {
		if c == '\t' {
			cnt++
		} else {
			return cnt
		}
	}
	return 0
}

func removeIndent(lines []string) []string {
	if len(lines) == 0 {
		return lines
	}

	minIndent := -1
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}

		indent := countIndent(l)
		if minIndent == -1 || minIndent > indent {
			minIndent = indent
		}
	}

	if minIndent == -1 {
		minIndent = 0
	}

	ret := make([]string, len(lines))
	for i, l := range lines {
		if strings.TrimSpace(l) == "" {
			ret[i] = ""
		} else {
			ret[i] = l[minIndent:]
		}
	}
	return ret
}

// Echo executes lambda and shows the code in the lambda, which it reads out of
// code at the line skip frames up -- skip is 1 for a direct caller, 2 for one
// that reaches Echo through a forwarder of its own.
func Echo(c *tgframe.Container, code string, lambda func(), skip int) {
	_, filename, line, ok := runtime.Caller(skip)
	if !ok {
		panic("Unable to get caller")
	}

	curCode, ok := codeCache.get(filename, line)
	if !ok {
		codeLines := strings.Split(code, "\n")
		left := line - 1
		right := line
		for i := line; i < len(codeLines); i++ {
			_, err := parser.ParseExpr(strings.Join(codeLines[left:i], "\n"))
			if err == nil {
				right = i
				break
			}
		}

		if left+1 > right-1 {
			panic("Fail to obtain code")
		}

		curCode = strings.Join(removeIndent(codeLines[left+1:right-1]), "\n")
		codeCache.set(filename, line, curCode)
	}

	lambda()
	tccontent.Code(c, curCode)
}
