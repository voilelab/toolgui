package tcdata

import (
	"strings"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func TestJSONSerializesAValue(t *testing.T) {
	props := addComponent(t, func(c *tgframe.Container) {
		JSON(c, map[string]int{"a": 1})
	})

	if props["value"] != `{"a":1}` {
		t.Errorf(`value = %v, want {"a":1}`, props["value"])
	}
}

// A string is taken as JSON already, so it reaches the viewer as it was
// written rather than as a quoted string.
func TestJSONPassesAJSONStringThrough(t *testing.T) {
	props := addComponent(t, func(c *tgframe.Container) {
		JSON(c, `{"a": 1}`)
	})

	if props["value"] != `{"a": 1}` {
		t.Errorf(`value = %v, want {"a": 1}`, props["value"])
	}
}

func TestJSONFails(t *testing.T) {
	for _, tc := range []struct {
		name string
		v    any
	}{
		{"a string that is not JSON", "not json at all"},
		{"a value that will not marshal", func() {}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := failMessage(t, func(c *tgframe.Container) {
				JSON(c, tc.v)
			})

			if !strings.Contains(msg, "serialize to JSON") {
				t.Errorf("message = %q, want it to name the serialization", msg)
			}
		})
	}
}
