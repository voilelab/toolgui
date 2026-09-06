package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// runner drives the page like a client does: it feeds events to a session and
// collects the checkboxes of the last run.
type runner struct {
	session *tgframe.Session
	packs   chan any
}

func newRunner(t *testing.T) *runner {
	t.Helper()

	app := tgframe.NewApp()
	app.AddPage("main", "Main", Main)

	r := &runner{packs: make(chan any, 256)}
	session, err := tgframe.NewSession(app, "main", tgframe.NewState(), func(pack any) error {
		r.packs <- pack
		return nil
	})
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	r.session = session
	t.Cleanup(session.Close)
	return r
}

// run applies an event and returns the labels of the checkboxes it rendered.
func (r *runner) run(t *testing.T, event any) []string {
	t.Helper()

	bs, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	if err := r.session.HandleRawEvent(bs); err != nil {
		t.Fatalf("handle event: %v", err)
	}

	labels := []string{}
	for {
		select {
		case pack := <-r.packs:
			if result, ok := pack.(*tgframe.ResultPack); ok {
				if !result.Success {
					t.Fatalf("run failed: %s", result.Error)
				}
				return labels
			}

			if label, ok := checkboxLabel(t, pack); ok {
				labels = append(labels, label)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for the run to finish")
		}
	}
}

// checkboxLabel returns the label of the checkbox a pack creates, if any.
func checkboxLabel(t *testing.T, pack any) (string, bool) {
	t.Helper()

	bs, err := json.Marshal(pack)
	if err != nil {
		t.Fatalf("marshal pack: %v", err)
	}

	var parsed struct {
		Component struct {
			Name  string `json:"name"`
			Label string `json:"label"`
		} `json:"component"`
	}
	if err := json.Unmarshal(bs, &parsed); err != nil {
		t.Fatalf("unmarshal pack: %v", err)
	}

	return parsed.Component.Label, parsed.Component.Name == "checkbox_component"
}

func input(id string, value any) any {
	return map[string]any{"type": "input", "id": id, "value": value}
}

func click(id string) any {
	return map[string]any{"type": "click", "id": id}
}

func equalLabels(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}

	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}

	return true
}

func TestTodoFlow(t *testing.T) {
	r := newRunner(t)

	r.run(t, input("textbox_component_Add todo", "buy milk"))
	if labels := r.run(t, click("button_component_Add")); !equalLabels(labels, []string{"buy milk"}) {
		t.Fatalf("after add: %v", labels)
	}

	r.run(t, input("textbox_component_Add todo", "walk dog"))
	labels := r.run(t, click("button_component_Add"))
	if !equalLabels(labels, []string{"buy milk", "walk dog"}) {
		t.Fatalf("after second add: %v", labels)
	}

	// Check the first todo, then remove the done ones: the removed item is
	// gone from the same run, not one run later.
	r.run(t, input("checkbox_component_todo_1", true))
	if labels := r.run(t, click("button_component_Remove done")); !equalLabels(labels, []string{"walk dog"}) {
		t.Fatalf("after remove done: %v", labels)
	}

	// The remaining todo keeps its own checkbox state, it doesn't inherit the
	// one of the removed item that used to sit at its index.
	if labels := r.run(t, click("button_component_Remove done")); !equalLabels(labels, []string{"walk dog"}) {
		t.Fatalf("second remove done: %v", labels)
	}
}

func TestAddEmptyTodo(t *testing.T) {
	r := newRunner(t)

	if labels := r.run(t, click("button_component_Add")); len(labels) != 0 {
		t.Fatalf("empty input should add nothing, got %v", labels)
	}
}
