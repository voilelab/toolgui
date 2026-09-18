package tcinput_test

import (
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var menuItems = []string{"Rename", "Duplicate", "Delete"}

func menu(state *tgframe.State, conf ...*tcinput.MenuConf) *int {
	return tcinput.Menu(defaultContainer(state), "Actions", menuItems, conf...)
}

// clickState is a state carrying the click the frontend would have sent.
func clickState(id string) *tgframe.State {
	state := tgframe.NewState()
	(&tgframe.EventClick{ID: id}).ApplyState(state)
	return state
}

func TestMenuNothingClicked(t *testing.T) {
	if got := menu(tgframe.NewState()); got != nil {
		t.Fatalf("Menu = %v, want nil", *got)
	}
}

func TestMenuReadsClickedItem(t *testing.T) {
	for want, id := range []string{
		"menu_component_Actions_0",
		"menu_component_Actions_1",
		"menu_component_Actions_2",
	} {
		got := menu(clickState(id))
		if got == nil {
			t.Fatalf("Menu on %s = nil, want %d", id, want)
		}

		if *got != want {
			t.Errorf("Menu on %s = %d, want %d", id, *got, want)
		}
	}
}

// The click belongs to the run that handles it, as a [tcinput.Button] press
// does: the next run sees nothing.
func TestMenuClickLastsOneRun(t *testing.T) {
	state := clickState("menu_component_Actions_1")
	if got := menu(state); got == nil || *got != 1 {
		t.Fatalf("Menu = %v, want 1", got)
	}

	// What a session does between runs.
	state.SetClickID("")

	if got := menu(state); got != nil {
		t.Fatalf("Menu on the next run = %d, want nil", *got)
	}
}

// A click on a menu's own id is the trigger button opening the dropdown,
// which the client handles and Go never hears about as a pick.
func TestMenuTriggerIsNotAnItem(t *testing.T) {
	if got := menu(clickState("menu_component_Actions")); got != nil {
		t.Fatalf("Menu = %d, want nil", *got)
	}
}

// Two menus sharing a label are told apart by Conf.ID, and the item ids
// follow it rather than the label.
func TestMenuConfIDSeparatesItems(t *testing.T) {
	state := clickState("menu_component_row_actions_2")

	if got := menu(state, &tcinput.MenuConf{ID: "row_actions"}); got == nil || *got != 2 {
		t.Fatalf("Menu with Conf.ID = %v, want 2", got)
	}

	// The same click leaves the menu that kept the label-derived id alone.
	if got := menu(state); got != nil {
		t.Fatalf("Menu without Conf.ID = %d, want nil", *got)
	}
}

// An index the menu no longer has is not a pick: a shorter list drops the
// click rather than clamping it, the way [tcinput.Select] drops a selection.
func TestMenuClickOutsideItems(t *testing.T) {
	state := clickState("menu_component_Actions_3")

	if got := menu(state); got != nil {
		t.Fatalf("Menu = %d, want nil", *got)
	}
}

func TestMenuNoItems(t *testing.T) {
	state := clickState("menu_component_Actions_0")

	got := tcinput.Menu(defaultContainer(state), "Actions", nil)
	if got != nil {
		t.Fatalf("Menu = %d, want nil", *got)
	}
}
