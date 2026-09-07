package tgframe

import "testing"

// The app title reaches the frontend, which puts it in the browser tab.
func TestAppConfTitle(t *testing.T) {
	app := NewApp()

	if got := app.AppConf().Title; got != "" {
		t.Errorf("AppConf().Title = %q, want empty by default", got)
	}

	app.SetTitle("My Tool")

	if got := app.AppConf().Title; got != "My Tool" {
		t.Errorf("AppConf().Title = %q, want %q", got, "My Tool")
	}
}
