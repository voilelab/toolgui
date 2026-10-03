//go:build js && wasm

package tgwasm

import (
	"net/url"
	"strings"
	"syscall/js"
)

// QueryName is the global the page sets its query string on, before the wasm
// program starts.
const QueryName = "toolguiQuery"

// Query returns the query string of the page the app is shown in, such as
// `?lang=en`. A worker's own location is its script, so the page hands this
// over at boot.
//
// It is read once: the query string cannot change without a page load, which
// starts the program again. A host that sets nothing, or something that is not
// a string, answers an empty set.
func Query() url.Values {
	v := js.Global().Get(QueryName)
	if v.Type() != js.TypeString {
		return url.Values{}
	}

	q, _ := url.ParseQuery(strings.TrimPrefix(v.String(), "?"))
	return q
}
