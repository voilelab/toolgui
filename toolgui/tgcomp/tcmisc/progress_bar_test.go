package tcmisc

import (
	"errors"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// runPage runs page as the only page of an app, on state.
func runPage(t *testing.T, state *tgframe.State, page tgframe.RunFunc) error {
	t.Helper()

	app := tgframe.NewApp()
	app.AddPage("index", "Index", page)

	return app.RunWithHandlingPanic("index", state, func(tgframe.NotifyPack) {})
}

// stateKeyOf is the key a bar declared with Conf.ID stores its state under:
// [tgframe.BaseComponent.SetID] prefixes the conf id with the component name.
func stateKeyOf(confID string) string {
	return progressBarComponentName + "_" + confID
}

func TestProgressBarRemoveFreesTheIDForReuse(t *testing.T) {
	err := runPage(t, tgframe.NewState(), func(p *tgframe.Params) error {
		bar := ProgressBar(p.Main, 0, "loading", &ProgressBarConf{ID: "bar"})
		bar.Remove()

		ProgressBar(p.Main, 100, "done", &ProgressBarConf{ID: "bar"})
		return nil
	})

	if errors.Is(err, tgframe.ErrDuplicatedID) {
		t.Fatalf("reusing the id after Remove reported %v", err)
	}
	if err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestProgressBarRemoveReleasesTheState(t *testing.T) {
	state := tgframe.NewState()
	state.Set(stateKeyOf("bar"), 42)

	err := runPage(t, state, func(p *tgframe.Params) error {
		ProgressBar(p.Main, 0, "loading", &ProgressBarConf{ID: "bar"}).Remove()
		return nil
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if got := state.GetInt(stateKeyOf("bar")); got != nil {
		t.Errorf("state under the removed id = %v, want it dropped", *got)
	}
}

func TestProgressBarIDClaimedAgainKeepsItsState(t *testing.T) {
	state := tgframe.NewState()
	state.Set(stateKeyOf("bar"), 42)

	err := runPage(t, state, func(p *tgframe.Params) error {
		ProgressBar(p.Main, 0, "loading", &ProgressBarConf{ID: "bar"}).Remove()
		ProgressBar(p.Main, 100, "done", &ProgressBarConf{ID: "bar"})
		return nil
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	got := state.GetInt(stateKeyOf("bar"))
	if got == nil || *got != 42 {
		t.Errorf("state under the reclaimed id = %v, want 42", got)
	}
}

func TestProgressBarWithoutIDRemoves(t *testing.T) {
	err := runPage(t, tgframe.NewState(), func(p *tgframe.Params) error {
		ProgressBar(p.Main, 0, "loading").Remove()
		return nil
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
}
