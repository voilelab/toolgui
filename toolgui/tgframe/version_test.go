package tgframe

import "testing"

func TestVersion(t *testing.T) {
	// Under `go test` toolgui is the main module and its build info version is
	// "(devel)", so this exercises the fallback.
	v := Version()
	if v == "" || v == "(devel)" {
		t.Errorf("Version() = %q, want a real version", v)
	}
}

func TestIsRealVersion(t *testing.T) {
	for _, tt := range []struct {
		v    string
		want bool
	}{
		{"", false},
		{"(devel)", false},
		// Stamped from a checkout with no tag to derive from.
		{"v0.0.0-20260908001841-a570ee0afdf3", false},
		{"v0.4.0", true},
		{"v0.4.1-0.20260908001841-a570ee0afdf3", true},
	} {
		if got := isRealVersion(tt.v); got != tt.want {
			t.Errorf("isRealVersion(%q) = %v, want %v", tt.v, got, tt.want)
		}
	}
}

func TestAppConfShowVersion(t *testing.T) {
	app := NewApp()

	if !app.AppConf().ShowVersion {
		t.Error("AppConf().ShowVersion = false, want true by default")
	}

	app.SetShowVersion(false)
	if app.AppConf().ShowVersion {
		t.Error("AppConf().ShowVersion = true after SetShowVersion(false)")
	}
}
