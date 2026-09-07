package tgframe

import "testing"

const testStateKey = "number_component_Age"

// TestGetFloatNumericSources covers the two routes a number reaches the state
// by: the frontend, which sends JSON and so lands a float64, and user code
// writing a default with Set, which writes whatever Go type it had at hand.
func TestGetFloatNumericSources(t *testing.T) {
	cases := []struct {
		name string
		val  any
	}{
		{"int", 30},
		{"int8", int8(30)},
		{"int16", int16(30)},
		{"int32", int32(30)},
		{"int64", int64(30)},
		{"uint", uint(30)},
		{"uint8", uint8(30)},
		{"uint16", uint16(30)},
		{"uint32", uint32(30)},
		{"uint64", uint64(30)},
		{"float32", float32(30)},
		{"float64", 30.0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			state := NewState()
			state.Set(testStateKey, c.val)

			got := state.GetFloat(testStateKey)
			if got == nil {
				t.Fatalf("GetFloat = nil, want 30")
			}
			if *got != 30 {
				t.Errorf("GetFloat = %v, want 30", *got)
			}

			gotInt := state.GetInt(testStateKey)
			if gotInt == nil {
				t.Fatalf("GetInt = nil, want 30")
			}
			if *gotInt != 30 {
				t.Errorf("GetInt = %v, want 30", *gotInt)
			}
		})
	}
}

// TestGettersOnWrongType pins the point of the whole exercise: a value of the
// wrong type reads back as absent, and never panics the page it was read from.
func TestGettersOnWrongType(t *testing.T) {
	cases := []struct {
		name string
		val  any
	}{
		{"string", "abc"},
		{"numeric string", "30"},
		{"bool", true},
		{"nil", nil},
		{"struct", struct{}{}},
		{"slice", []int{1}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			state := NewState()
			state.Set(testStateKey, c.val)

			if got := state.GetFloat(testStateKey); got != nil {
				t.Errorf("GetFloat = %v, want nil", *got)
			}
			if got := state.GetInt(testStateKey); got != nil {
				t.Errorf("GetInt = %v, want nil", *got)
			}
			// Each getter is skipped for the one case that is its own type.
			if _, isBool := c.val.(bool); !isBool {
				if got := state.GetBool(testStateKey); got {
					t.Errorf("GetBool = %v, want false", got)
				}
			}
			if _, isString := c.val.(string); !isString {
				if got := state.GetString(testStateKey); got != nil {
					t.Errorf("GetString = %v, want nil", *got)
				}
			}
		})
	}
}

func TestGettersOnMissingKey(t *testing.T) {
	state := NewState()

	if got := state.GetString("missing"); got != nil {
		t.Errorf("GetString = %v, want nil", *got)
	}
	if got := state.GetFloat("missing"); got != nil {
		t.Errorf("GetFloat = %v, want nil", *got)
	}
	if got := state.GetInt("missing"); got != nil {
		t.Errorf("GetInt = %v, want nil", *got)
	}
	if got := state.GetBool("missing"); got {
		t.Errorf("GetBool = %v, want false", got)
	}
}

func TestGettersOnRightType(t *testing.T) {
	state := NewState()
	state.Set("s", "abc")
	state.Set("b", true)
	state.Set("f", 1.5)

	if got := state.GetString("s"); got == nil || *got != "abc" {
		t.Errorf("GetString = %v, want abc", got)
	}
	if got := state.GetBool("b"); !got {
		t.Errorf("GetBool = %v, want true", got)
	}
	if got := state.GetFloat("f"); got == nil || *got != 1.5 {
		t.Errorf("GetFloat = %v, want 1.5", got)
	}
	// An int reads a fractional value truncated, the way the number
	// component does when its T is integral.
	if got := state.GetInt("f"); got == nil || *got != 1 {
		t.Errorf("GetInt = %v, want 1", got)
	}
}

func TestGet(t *testing.T) {
	state := NewState()
	state.Set("s", "abc")

	if got, ok := Get[string](state, "s"); !ok || got != "abc" {
		t.Errorf("Get[string] = %v, %v, want abc, true", got, ok)
	}
	// Get is a plain type assertion: unlike GetFloat it does not read one
	// numeric type as another.
	if got, ok := Get[float64](state, "s"); ok || got != 0 {
		t.Errorf("Get[float64] = %v, %v, want 0, false", got, ok)
	}
	if got, ok := Get[string](state, "missing"); ok || got != "" {
		t.Errorf("Get[string] = %q, %v, want \"\", false", got, ok)
	}
}

type testTODOList struct {
	Items []string
}

func TestDefault(t *testing.T) {
	state := NewState()

	list := Default(state, "todoList", testTODOList{})
	if len(list.Items) != 0 {
		t.Fatalf("Items = %v, want empty", list.Items)
	}

	list.Items = append(list.Items, "buy milk")

	// The state keeps the pointer, so the next run reads the write back.
	again := Default(state, "todoList", testTODOList{})
	if len(again.Items) != 1 || again.Items[0] != "buy milk" {
		t.Errorf("Items = %v, want [buy milk]", again.Items)
	}
}

func TestDefaultOnWrongType(t *testing.T) {
	state := NewState()
	state.Set("todoList", "not a list")

	list := Default(state, "todoList", testTODOList{Items: []string{"a"}})
	if len(list.Items) != 1 || list.Items[0] != "a" {
		t.Errorf("Items = %v, want [a]", list.Items)
	}
}
