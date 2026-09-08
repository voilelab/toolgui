package tgframe_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// keysOf runs one page and returns the key of every component it sent, in
// order, paired with the id the component claimed.
func keysOf(t *testing.T, page tgframe.RunFunc) ([]string, []string, error) {
	t.Helper()

	// Read the packs the way the GUI client does, off the wire.
	var pack struct {
		Type      int    `json:"type"`
		Key       string `json:"key"`
		Component struct {
			ID string `json:"id"`
		} `json:"component"`
	}

	var keys, ids []string
	app := tgframe.NewApp()
	app.AddPage("index", "Index", page)

	err := app.Run("index", tgframe.NewState(), func(p tgframe.NotifyPack) {
		bs, mErr := json.Marshal(p)
		if mErr != nil {
			t.Fatalf("marshal pack: %v", mErr)
		}
		if uErr := json.Unmarshal(bs, &pack); uErr != nil {
			t.Fatalf("unmarshal pack: %v", uErr)
		}
		if pack.Type != tgframe.NotifyTypeCreate {
			return
		}

		keys = append(keys, pack.Key)
		ids = append(ids, pack.Component.ID)
	})

	return keys, ids, err
}

func TestIdenticalComponentsGetTheirOwnKey(t *testing.T) {
	keys, ids, err := keysOf(t, func(p *tgframe.Params) error {
		tgcomp.Text(p.Main, "duplicate me")
		tgcomp.Text(p.Main, "duplicate me")
		tgcomp.Markdown(p.Main, "**dup markdown**")
		tgcomp.Divider(p.Main)
		tgcomp.Divider(p.Main)
		return nil
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	want := []string{
		"container_component_container_main/0",
		"container_component_container_main/1",
		"container_component_container_main/2",
		"container_component_container_main/3",
		"container_component_container_main/4",
	}
	if len(keys) != len(want) {
		t.Fatalf("got %d components, want %d: %v", len(keys), len(want), keys)
	}
	for i := range want {
		if keys[i] != want[i] {
			t.Errorf("key %d = %q, want %q", i, keys[i], want[i])
		}
		if ids[i] != "" {
			t.Errorf("component %d claims id %q; components without state should claim none", i, ids[i])
		}
	}
}

func TestNestedContainersKeyUnderTheirComponent(t *testing.T) {
	keys, _, err := keysOf(t, func(p *tgframe.Params) error {
		tgcomp.Text(p.Main, "before")
		left, right := tgcomp.Column2(p.Main)
		tgcomp.Text(left, "in left")
		tgcomp.Text(right, "in right")
		return nil
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	want := []string{
		"container_component_container_main/0", // the text
		"container_component_container_main/1", // the column component
		"container_component_container_main/1/0",
		"container_component_container_main/1/1",
		"container_component_container_main/1/0/0", // the text in the left column
		"container_component_container_main/1/1/0",
	}
	if len(keys) != len(want) {
		t.Fatalf("got %v, want %v", keys, want)
	}
	for i := range want {
		if keys[i] != want[i] {
			t.Errorf("key %d = %q, want %q", i, keys[i], want[i])
		}
	}
}

func TestDuplicatedWidgetIDIsAnError(t *testing.T) {
	_, _, err := keysOf(t, func(p *tgframe.Params) error {
		tgcomp.Button(p.Main, "ok")
		tgcomp.Button(p.Main, "ok")
		return nil
	})

	if !errors.Is(err, tgframe.ErrDuplicatedID) {
		t.Fatalf("err = %v, want ErrDuplicatedID", err)
	}
}

func TestAUserSuppliedIDMakesDuplicatedWidgetsLegal(t *testing.T) {
	_, ids, err := keysOf(t, func(p *tgframe.Params) error {
		tgcomp.Button(p.Main, "ok")
		tgcomp.Button(p.Main, "ok", &tgcomp.ButtonConf{ID: "second"})
		return nil
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	want := []string{"button_component_ok", "button_component_second"}
	for i := range want {
		if ids[i] != want[i] {
			t.Errorf("id %d = %q, want %q", i, ids[i], want[i])
		}
	}
}

func TestComponentsWithClientStateClaimAnID(t *testing.T) {
	_, ids, err := keysOf(t, func(p *tgframe.Params) error {
		tgcomp.Expand(p.Main, "Details", false)
		tgcomp.Tab2(p.Main, "one", "two")
		tgcomp.JSON(p.Main, map[string]int{"a": 1})
		return nil
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	// The client keeps which tab is open, whether an expander is open and
	// which JSON nodes are collapsed, so those components need a name of
	// their own rather than only a position.
	for i, id := range ids {
		if id == "" {
			t.Errorf("component %d claims no id, so its client state would follow its position", i)
		}
	}
}

func TestDuplicatedExpandIsAnError(t *testing.T) {
	_, _, err := keysOf(t, func(p *tgframe.Params) error {
		tgcomp.Expand(p.Main, "Details", false)
		tgcomp.Expand(p.Main, "Details", false)
		return nil
	})

	if !errors.Is(err, tgframe.ErrDuplicatedID) {
		t.Fatalf("err = %v, want ErrDuplicatedID", err)
	}
}

func TestAnExpandConfIDMakesDuplicatesLegal(t *testing.T) {
	_, _, err := keysOf(t, func(p *tgframe.Params) error {
		tgcomp.Expand(p.Main, "Details", false)
		tgcomp.Expand(p.Main, "Details", false,
			&tgcomp.ExpandConf{ID: "second_details"})
		return nil
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestTakingARootContainersIDIsAnError(t *testing.T) {
	_, _, err := keysOf(t, func(p *tgframe.Params) error {
		// The roots are never sent, so nothing else would collide with them
		// unless the run claims their ids up front.
		p.Main.AddContainer(tgframe.MainContainerID)
		return nil
	})

	if !errors.Is(err, tgframe.ErrDuplicatedID) {
		t.Fatalf("err = %v, want ErrDuplicatedID", err)
	}
}

// A run records the ids it drew on the state, so a transport can tell a
// component of the page from a name a client made up.
func TestRunRecordsComponentIDs(t *testing.T) {
	app := tgframe.NewApp()
	app.AddPage("index", "Index", func(p *tgframe.Params) error {
		tgcomp.Text(p.Main, "hello", &tgcomp.TextConf{
			Base: tgframe.Base{ID: "drawn"},
		})
		return nil
	})

	state := tgframe.NewState()

	// The id a component carries is its name and the one the conf gave it.
	const drawnID = "text_component_drawn"

	if state.HasComponentID(drawnID) {
		t.Error("expect no component id before the page runs")
	}

	err := app.Run("index", state, func(p tgframe.NotifyPack) {})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if !state.HasComponentID(drawnID) {
		t.Error("expect the drawn component id to be recorded")
	}

	if state.HasComponentID("never_drawn") {
		t.Error("expect an id the page never drew to be unknown")
	}
}
