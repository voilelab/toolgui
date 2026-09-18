package demos

import (
	"strings"
	"testing"
)

func TestDedent(t *testing.T) {
	cases := []struct {
		name string
		code string
		want string
	}{
		{
			name: "function body loses the enclosing tab",
			code: "\ta()\n\tif x {\n\t\tb()\n\t}",
			want: "a()\nif x {\n\tb()\n}",
		},
		{
			name: "a blank line does not hold the indent down",
			code: "\ta()\n\n\tb()",
			want: "a()\n\nb()",
		},
		{
			name: "a line at column 0 leaves the block alone",
			code: "\ts := `first\nsecond`\n\tb()",
			want: "\ts := `first\nsecond`\n\tb()",
		},
		{
			name: "nothing shared, nothing removed",
			code: "a()\nb()",
			want: "a()\nb()",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := dedent(c.code); got != c.want {
				t.Errorf("dedent = %q, want %q", got, c.want)
			}
		})
	}
}

// TestBlockCodeIsDedented is the whole point of it: a code column half a page
// wide cannot spend a tab per line on the function the snippet was cut out of.
func TestBlockCodeIsDedented(t *testing.T) {
	all := []*Demo{Plugin()}
	for _, g := range Groups() {
		all = append(all, g.Demos...)
	}

	for _, d := range all {
		for _, b := range d.Blocks {
			first := strings.SplitN(b.Code, "\n", 2)[0]
			if strings.HasPrefix(first, "\t") {
				t.Errorf("%s: %s starts indented: %q", d.Name, b.ID, first)
			}
		}
	}
}
