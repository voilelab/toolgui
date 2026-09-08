package tcinput_test

import (
	"slices"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var multiselectItems = []string{"dev", "stage", "prod"}

func multiselect(state *tgframe.State, conf ...*tcinput.MultiselectConf) []int {
	return tcinput.Multiselect(defaultContainer(state), "Env", multiselectItems, conf...)
}

func TestMultiselectNothingSelected(t *testing.T) {
	got := multiselect(tgframe.NewState())
	if got == nil {
		t.Fatal("Multiselect = nil, want an empty slice")
	}

	if len(got) != 0 {
		t.Fatalf("Multiselect = %v, want empty", got)
	}
}

func TestMultiselectDefault(t *testing.T) {
	got := multiselect(tgframe.NewState(), &tcinput.MultiselectConf{Default: []int{2, 0}})
	if !slices.Equal(got, []int{0, 2}) {
		t.Fatalf("Multiselect = %v, want [0 2]", got)
	}
}

// The default only stands in until the app user has answered, and an empty
// selection is an answer.
func TestMultiselectEmptySelectionBeatsDefault(t *testing.T) {
	state := tgframe.NewState()
	(&tgframe.EventSelect{
		ID:     "multiselect_component_Env",
		Values: []int{},
	}).ApplyState(state)

	got := multiselect(state, &tcinput.MultiselectConf{Default: []int{1}})
	if len(got) != 0 {
		t.Fatalf("Multiselect = %v, want empty", got)
	}
}

// The state holds float64s once the selection has been through JSON, which is
// how it arrives from a real frontend.
func TestMultiselectReadsSelection(t *testing.T) {
	state := tgframe.NewState()
	state.Set("multiselect_component_Env", []float64{2, 1})

	if got := multiselect(state); !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("Multiselect = %v, want [1 2]", got)
	}
}

// Writing the key from Go, with a Go type, is how a page gives the component
// an initial value without a conf.
func TestMultiselectSetAsInitialValue(t *testing.T) {
	state := tgframe.NewState()
	state.Set("multiselect_component_Env", []int{0, 1})

	if got := multiselect(state); !slices.Equal(got, []int{0, 1}) {
		t.Fatalf("Multiselect = %v, want [0 1]", got)
	}
}

// Anything the items cannot name is dropped rather than handed to the page,
// which would index out of range on it.
func TestMultiselectDropsWhatItemsDoNotName(t *testing.T) {
	state := tgframe.NewState()
	state.Set("multiselect_component_Env", []int{-1, 1, 1, 3, 99})

	if got := multiselect(state); !slices.Equal(got, []int{1}) {
		t.Fatalf("Multiselect = %v, want [1]", got)
	}
}

// The frontend keeps the app user under the cap; a payload that got past it
// anyway is trimmed, so the page never sees more than it asked for.
func TestMultiselectTrimsToMaxSelections(t *testing.T) {
	state := tgframe.NewState()
	state.Set("multiselect_component_Env", []int{0, 1, 2})

	got := multiselect(state, &tcinput.MultiselectConf{MaxSelections: 2})
	if !slices.Equal(got, []int{0, 1}) {
		t.Fatalf("Multiselect = %v, want [0 1]", got)
	}
}

// The caller owns what it gets back, so writing to it must not reach into the
// component that was just sent.
func TestMultiselectDefaultIsNotShared(t *testing.T) {
	conf := &tcinput.MultiselectConf{Default: []int{0, 1}}

	got := multiselect(tgframe.NewState(), conf)
	got[0] = 2

	if !slices.Equal(conf.Default, []int{0, 1}) {
		t.Fatalf("conf.Default = %v, want [0 1]", conf.Default)
	}

	if again := multiselect(tgframe.NewState(), conf); !slices.Equal(again, []int{0, 1}) {
		t.Fatalf("Multiselect = %v, want [0 1]", again)
	}
}
