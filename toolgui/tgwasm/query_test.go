//go:build js && wasm

package tgwasm

import (
	"syscall/js"
	"testing"
)

func TestQuery(t *testing.T) {
	cases := []struct {
		name string
		set  any
		want string
	}{
		{name: "unset", set: nil, want: ""},
		{name: "empty", set: "", want: ""},
		{name: "search", set: "?lang=en&embed", want: "en"},
		{name: "no question mark", set: "lang=zh-TW", want: "zh-TW"},
		{name: "escaped", set: "?lang=a%20b", want: "a b"},
		{name: "not a string", set: true, want: ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.set == nil {
				js.Global().Delete(QueryName)
			} else {
				js.Global().Set(QueryName, c.set)
			}
			defer js.Global().Delete(QueryName)

			if got := Query().Get("lang"); got != c.want {
				t.Errorf("Query().Get(%q) = %q, want %q", "lang", got, c.want)
			}
		})
	}
}
