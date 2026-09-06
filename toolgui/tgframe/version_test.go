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

func TestAppConfVersion(t *testing.T) {
	app := NewApp()

	conf := app.AppConf()
	if conf.Version != Version() {
		t.Errorf("AppConf().Version = %q, want %q", conf.Version, Version())
	}
	if !conf.ShowVersion {
		t.Error("AppConf().ShowVersion = false, want true by default")
	}

	app.SetShowVersion(false)
	if app.AppConf().ShowVersion {
		t.Error("AppConf().ShowVersion = true after SetShowVersion(false)")
	}
}
