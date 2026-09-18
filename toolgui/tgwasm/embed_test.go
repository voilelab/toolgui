//go:build js && wasm

package tgwasm

import (
	"syscall/js"
	"testing"
)

func TestEmbedded(t *testing.T) {
	cases := []struct {
		name string
		set  any
		want bool
	}{
		{name: "unset", set: nil, want: false},
		{name: "true", set: true, want: true},
		{name: "false", set: false, want: false},
		// A host that writes something else says nothing: the flag is a
		// boolean, and guessing at a string would read "false" as embedded.
		{name: "string", set: "true", want: false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.set == nil {
				js.Global().Delete(EmbedName)
			} else {
				js.Global().Set(EmbedName, c.set)
			}
			defer js.Global().Delete(EmbedName)

			if got := Embedded(); got != c.want {
				t.Errorf("Embedded() = %v, want %v", got, c.want)
			}
		})
	}
}
