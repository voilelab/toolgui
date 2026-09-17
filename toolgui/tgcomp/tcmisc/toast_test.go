package tcmisc_test

import (
	"testing"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcmisc"
	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgjson"
)

// toastsOf runs page once and returns the props of every toast it wrote, in
// the order it wrote them, each paired with the node key it was sent to.
func toastsOf(t *testing.T, app *tgframe.App, state *tgframe.State) []toastPack {
	t.Helper()

	var toasts []toastPack
	err := app.Run("index", state, func(p tgframe.NotifyPack) {
		bs, mErr := tgjson.Marshal(p)
		if mErr != nil {
			t.Fatalf("marshal pack: %v", mErr)
		}

		var one toastPack
		if uErr := tgjson.Unmarshal(bs, &one); uErr != nil {
			t.Fatalf("unmarshal pack: %v", uErr)
		}

		if one.Component.Name == "toast_component" {
			toasts = append(toasts, one)
		}
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	return toasts
}

// toastPack is a notify pack carrying a toast, cut down to what a test
// asserts on.
type toastPack struct {
	Key       string `json:"key"`
	Component struct {
		Name       string `json:"name"`
		ID         string `json:"id"`
		Text       string `json:"text"`
		Icon       string `json:"icon"`
		DurationMS int64  `json:"duration_ms"`
		Seq        uint64 `json:"seq"`
	} `json:"component"`
}

func appOf(page tgframe.RunFunc) *tgframe.App {
	app := tgframe.NewApp()
	app.AddPage("index", "Index", page)
	return app
}

// The one the ticket is about: a toast is an event, and the forest only holds
// nodes. Two runs write the same call in the same place, so the node and its
// key are the same both times; the serial is what has to differ, or the client
// sees nothing new and the second toast never fires.
func TestToastFiresAgainOnTheNextRun(t *testing.T) {
	app := appOf(func(p *tgframe.Params) error {
		tgcomp.Toast(p.Main, "Saved")
		return nil
	})

	state := tgframe.NewState()

	first := toastsOf(t, app, state)
	second := toastsOf(t, app, state)

	if len(first) != 1 || len(second) != 1 {
		t.Fatalf("got %d then %d toasts, want 1 then 1", len(first), len(second))
	}

	if first[0].Key != second[0].Key {
		t.Fatalf("key = %q then %q, want the same node both runs",
			first[0].Key, second[0].Key)
	}

	if first[0].Component.Seq == second[0].Component.Seq {
		t.Errorf("seq = %d both runs, want the second run to differ",
			first[0].Component.Seq)
	}
}

// Two calls in one run are two nodes, and the client shows two toasts. They
// share the run's serial, which is what "fired once per run" means: the serial
// says which run, not which toast.
func TestToastTwiceInOneRunAreTwoNodes(t *testing.T) {
	app := appOf(func(p *tgframe.Params) error {
		tgcomp.Toast(p.Main, "First")
		tgcomp.Toast(p.Main, "Second")
		return nil
	})

	toasts := toastsOf(t, app, tgframe.NewState())
	if len(toasts) != 2 {
		t.Fatalf("got %d toasts, want 2", len(toasts))
	}

	if toasts[0].Key == toasts[1].Key {
		t.Errorf("both toasts sent to %q, want two nodes", toasts[0].Key)
	}

	if toasts[0].Component.Seq != toasts[1].Component.Seq {
		t.Errorf("seq = %d and %d, want one serial for the run",
			toasts[0].Component.Seq, toasts[1].Component.Seq)
	}
}

func TestToastConfDrivesTheProps(t *testing.T) {
	app := appOf(func(p *tgframe.Params) error {
		tgcomp.Toast(p.Main, "Saved", &tgcomp.ToastConf{
			ID:       "saved",
			Icon:     "✅",
			Duration: 2 * time.Second,
		})
		return nil
	})

	toasts := toastsOf(t, app, tgframe.NewState())
	if len(toasts) != 1 {
		t.Fatalf("got %d toasts, want 1", len(toasts))
	}

	props := toasts[0].Component
	if props.Text != "Saved" {
		t.Errorf("text = %q, want Saved", props.Text)
	}

	if props.Icon != "✅" {
		t.Errorf("icon = %q, want ✅", props.Icon)
	}

	if props.DurationMS != 2000 {
		t.Errorf("duration_ms = %d, want 2000", props.DurationMS)
	}

	if props.ID != "toast_component_saved" {
		t.Errorf("id = %q, want toast_component_saved", props.ID)
	}
}

// Zero means the client's default, so it has to arrive as zero rather than as
// a duration of no time at all.
func TestToastDefaultDurationIsZero(t *testing.T) {
	app := appOf(func(p *tgframe.Params) error {
		tgcomp.Toast(p.Main, "Saved")
		return nil
	})

	toasts := toastsOf(t, app, tgframe.NewState())
	if len(toasts) != 1 {
		t.Fatalf("got %d toasts, want 1", len(toasts))
	}

	if toasts[0].Component.DurationMS != 0 {
		t.Errorf("duration_ms = %d, want 0", toasts[0].Component.DurationMS)
	}
}

// A duration shorter than the millisecond the client counts in was still a
// duration the caller asked for, and must not round away into "default".
func TestToastSubMillisecondDurationStaysSet(t *testing.T) {
	app := appOf(func(p *tgframe.Params) error {
		tgcomp.Toast(p.Main, "Saved", &tgcomp.ToastConf{Duration: time.Microsecond})
		return nil
	})

	toasts := toastsOf(t, app, tgframe.NewState())
	if len(toasts) != 1 {
		t.Fatalf("got %d toasts, want 1", len(toasts))
	}

	if toasts[0].Component.DurationMS != 1 {
		t.Errorf("duration_ms = %d, want 1", toasts[0].Component.DurationMS)
	}
}

func TestToastNegativeDurationPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("no panic, want one for a negative duration")
		}
	}()

	c := tgframe.NewContainer("test", tgframe.NewState(), func(tgframe.NotifyPack) {})
	tcmisc.Toast(c, "Saved", &tcmisc.ToastConf{Duration: -time.Second})
}
