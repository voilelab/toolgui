package tcecho

import (
	"strings"
	"testing"
)

func TestRemoveIndent(t *testing.T) {
	cases := []struct {
		name  string
		lines []string
		want  []string
	}{
		{
			name:  "the shared tab goes, the deeper one stays",
			lines: []string{"\t\ta()", "\t\tif x {", "\t\t\tb()", "\t\t}"},
			want:  []string{"a()", "if x {", "\tb()", "}"},
		},
		{
			name:  "a whitespace only line is emptied, not measured",
			lines: []string{"\t\ta()", "\t\t", "\t\tb()"},
			want:  []string{"a()", "", "b()"},
		},
		{
			name:  "a line at column 0 leaves the lines alone",
			lines: []string{"\t\ts := `first", "second`", "\t\tb()"},
			want:  []string{"\t\ts := `first", "second`", "\t\tb()"},
		},
		{
			name:  "nothing to remove",
			lines: []string{},
			want:  []string{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := removeIndent(c.lines)
			if strings.Join(got, "\n") != strings.Join(c.want, "\n") {
				t.Errorf("removeIndent = %q, want %q", got, c.want)
			}
		})
	}
}

func TestCountIndent(t *testing.T) {
	cases := map[string]int{
		"":         0,
		"a":        0,
		"\ta":      1,
		"\t\ta":    2,
		"\t\t":     2,
		"  a":      0,
		"\t\t // ": 2,
	}

	for line, want := range cases {
		if got := countIndent(line); got != want {
			t.Errorf("countIndent(%q) = %d, want %d", line, got, want)
		}
	}
}
