package tcinput_test

import (
	"encoding/json/jsontext"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgjson"
)

// Rating is a user-defined named type, the case the `~` in the constraint is
// there for.
type Rating int

func testContainer(state *tgframe.State, packs *[]tgframe.NotifyPack) *tgframe.Container {
	return tgframe.NewContainer("test", state, func(pack tgframe.NotifyPack) {
		*packs = append(*packs, pack)
	})
}

// TestNumberAcceptsEveryTypeInTheSet checks the constraint widened from
// `float64 | int64` to `~int | ~int64 | ~float64` really instantiates, and
// that the round trip through the state survives it: the client sends a JSON
// number, so [tgframe.State.GetFloat] hands back a float64 whatever T is.
func TestNumberAcceptsEveryTypeInTheSet(t *testing.T) {
	const id = "number_component_n"

	t.Run("float64", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set(id, 2.5)
		var packs []tgframe.NotifyPack
		got := tcinput.Number[float64](testContainer(state, &packs), "n")
		if got == nil || *got != 2.5 {
			t.Fatalf("got %v, want 2.5", got)
		}
	})

	t.Run("int64", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set(id, 7.0)
		var packs []tgframe.NotifyPack
		got := tcinput.Number[int64](testContainer(state, &packs), "n")
		if got == nil || *got != 7 {
			t.Fatalf("got %v, want 7", got)
		}
	})

	// int is the type S5 was actually missing.
	t.Run("int", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set(id, 7.0)
		var packs []tgframe.NotifyPack
		got := tcinput.Number[int](testContainer(state, &packs), "n")
		if got == nil || *got != 7 {
			t.Fatalf("got %v, want 7", got)
		}
	})

	t.Run("named type", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set(id, 4.0)
		var packs []tgframe.NotifyPack
		got := tcinput.Number[Rating](testContainer(state, &packs), "n")
		if got == nil || *got != Rating(4) {
			t.Fatalf("got %v, want Rating(4)", got)
		}
	})
}

// TestNumberTruncatesTowardsTheIntegralType records what an integral T does
// with the fractional float the state can hold: it truncates, it does not
// round and it does not report the loss.
func TestNumberTruncatesTowardsTheIntegralType(t *testing.T) {
	state := tgframe.NewState()
	state.Set("number_component_n", 2.9)

	var packs []tgframe.NotifyPack
	got := tcinput.Number[int](testContainer(state, &packs), "n")
	if got == nil || *got != 2 {
		t.Fatalf("got %v, want 2", got)
	}
}

// zeroStep is a conf whose step is explicitly zero, the case Number rewrites.
// A function literal cannot take type parameters, so it lives out here.
func zeroStep[T tcinput.Numeric]() *tcinput.NumberConf[T] {
	return (&tcinput.NumberConf[T]{}).SetStep(0)
}

// TestNumberStepsByOneForIntegralTypes checks the int64 special case survived
// being rewritten as arithmetic rather than a type switch — a type switch
// cannot see a named type, so `Rating` would have slipped past one.
func TestNumberStepsByOneForIntegralTypes(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func(*tgframe.Container)
		want string
	}{
		{"int", func(c *tgframe.Container) { tcinput.Number(c, "n", zeroStep[int]()) }, "1"},
		{"int64", func(c *tgframe.Container) { tcinput.Number(c, "n", zeroStep[int64]()) }, "1"},
		{"named", func(c *tgframe.Container) { tcinput.Number(c, "n", zeroStep[Rating]()) }, "1"},
		{"float64", func(c *tgframe.Container) { tcinput.Number(c, "n", zeroStep[float64]()) }, "0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var packs []tgframe.NotifyPack
			tc.call(testContainer(tgframe.NewState(), &packs))

			if len(packs) != 1 {
				t.Fatalf("got %d packs, want 1", len(packs))
			}

			bs, err := tgjson.Marshal(packs[0])
			if err != nil {
				t.Fatal(err)
			}

			var got struct {
				Component struct {
					Step jsontext.Value `json:"step"`
				} `json:"component"`
			}
			if err := tgjson.Unmarshal(bs, &got); err != nil {
				t.Fatal(err)
			}

			// A semantic compare, so the step is pinned to the number it
			// is rather than to the digits the encoder wrote.
			step, want := got.Component.Step, jsontext.Value(tc.want)
			if err := step.Canonicalize(); err != nil {
				t.Fatalf("canonicalize step: %v", err)
			}
			if err := want.Canonicalize(); err != nil {
				t.Fatalf("canonicalize want: %v", err)
			}

			if string(step) != string(want) {
				t.Errorf("step = %q, want %q", step, want)
			}
		})
	}
}

// TestNumberDoesNotWriteBackToTheCallersConf pins a bug the old
// NumberWithConf had: it defaulted Step by assigning into the caller's conf,
// so a conf reused across runs was mutated behind the caller's back.
func TestNumberDoesNotWriteBackToTheCallersConf(t *testing.T) {
	conf := (&tcinput.NumberConf[int]{}).SetStep(0)

	var packs []tgframe.NotifyPack
	tcinput.Number(testContainer(tgframe.NewState(), &packs), "n", conf)

	if *conf.Step != 0 {
		t.Errorf("conf.Step = %v, want 0", *conf.Step)
	}
}

// TestNumberInfersTFromExplicitInstantiation is the acceptance condition the
// variadic conf put at risk: with no conf argument there is nothing for
// inference to work from, so T has to come from the explicit instantiation
// alone. It does — a type parameter list is never inferred from a variadic
// that was not passed.
func TestNumberInfersTFromExplicitInstantiation(t *testing.T) {
	state := tgframe.NewState()
	state.Set("number_component_n", 3.0)

	var packs []tgframe.NotifyPack
	c := testContainer(state, &packs)

	// No conf, T from the instantiation.
	if got := tcinput.Number[int](c, "n"); got == nil || *got != 3 {
		t.Errorf("Number[int] = %v, want 3", got)
	}
	if got := tcinput.Number[Rating](c, "n"); got == nil || *got != Rating(3) {
		t.Errorf("Number[Rating] = %v, want Rating(3)", got)
	}

	// And with a conf, T can instead be inferred from the conf alone.
	got := tcinput.Number(c, "n", &tcinput.NumberConf[int64]{})
	if got == nil || *got != 3 {
		t.Errorf("Number(conf) = %v, want 3", got)
	}
}

// TestNumberConfEmbedsBase checks a generic conf embeds Base like any other,
// and that the flat literal reaches it.
func TestNumberConfEmbedsBase(t *testing.T) {
	conf := &tcinput.NumberConf[int]{ID: "count"}

	// The state is keyed by the id the component ends up with, so reading the
	// value back under "count" is what proves the flat literal reached the
	// embedded Base: nothing else feeds that id.
	state := tgframe.NewState()
	state.Set("number_component_count", 5.0)

	var packs []tgframe.NotifyPack
	got := tcinput.Number(testContainer(state, &packs), "n", conf)
	if got == nil || *got != 5 {
		t.Fatalf("got %v, want 5 read under the conf's id", got)
	}
}

// TestNumberOmitsUnsetBounds pins that an unset bound stays off the wire.
// The fields are `*T`, so `omitzero` drops a nil pointer the same way
// `omitempty` did — but an explicit zero bound is a value, not an absence,
// and still has to be sent.
func TestNumberOmitsUnsetBounds(t *testing.T) {
	var packs []tgframe.NotifyPack
	tcinput.Number[int](testContainer(tgframe.NewState(), &packs), "n")

	if len(packs) != 1 {
		t.Fatalf("got %d packs, want 1", len(packs))
	}

	bs, err := tgjson.Marshal(packs[0])
	if err != nil {
		t.Fatal(err)
	}

	var got struct {
		Component map[string]jsontext.Value `json:"component"`
	}
	if err := tgjson.Unmarshal(bs, &got); err != nil {
		t.Fatal(err)
	}

	// Step is not in the list: Number defaults an integral T's step to 1.
	for _, name := range []string{"default", "min", "max"} {
		if v, ok := got.Component[name]; ok {
			t.Errorf("%s = %s, want it left out", name, v)
		}
	}

	// An explicitly zero bound is a bound, so it is written.
	packs = nil
	conf := (&tcinput.NumberConf[int]{}).SetMin(0)
	tcinput.Number(testContainer(tgframe.NewState(), &packs), "n", conf)

	bs, err = tgjson.Marshal(packs[0])
	if err != nil {
		t.Fatal(err)
	}

	got.Component = nil
	if err := tgjson.Unmarshal(bs, &got); err != nil {
		t.Fatal(err)
	}

	if v, ok := got.Component["min"]; !ok || string(v) != "0" {
		t.Errorf("min = %s (present %v), want 0", v, ok)
	}
}

// TestNumberRefusesAValueOutOfRange pins the answer to the bug this range
// check is for: the client sends whatever the user typed, so a value outside
// Min/Max has to be reported as no value. Handing back the last legal one --
// or the Default -- is what let a page save a number nobody had entered.
func TestNumberRefusesAValueOutOfRange(t *testing.T) {
	const id = "number_component_n"

	bounded := func() *tcinput.NumberConf[int] {
		return (&tcinput.NumberConf[int]{}).SetMin(10).SetMax(20)
	}

	for _, tc := range []struct {
		name string
		sent float64
		conf *tcinput.NumberConf[int]
		want *int
	}{
		{"below min", 9, bounded(), nil},
		{"above max", 21, bounded(), nil},
		{"at min", 10, bounded(), ptr(10)},
		{"at max", 20, bounded(), ptr(20)},
		{"within", 15, bounded(), ptr(15)},

		// The Default is not a fallback for a refused value: it is as stale
		// as the last legal one, and the box is not showing it.
		{"out of range with a default", 99, bounded().SetDefault(12), nil},

		// Only the bound that is set is judged.
		{"min only", 99, (&tcinput.NumberConf[int]{}).SetMin(10), ptr(99)},
		{"max only", -99, (&tcinput.NumberConf[int]{}).SetMax(20), ptr(-99)},

		// No bounds, no check: this is the Number that behaves as it always
		// did.
		{"no bounds", 9999, &tcinput.NumberConf[int]{}, ptr(9999)},

		// The range is judged on what the page is handed, so an integral T is
		// truncated first: 20.9 is the 20 that is in range, not the 20.9 that
		// is not.
		{"truncated into range", 20.9, bounded(), ptr(20)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := tgframe.NewState()
			state.Set(id, tc.sent)

			var packs []tgframe.NotifyPack
			got := tcinput.Number(testContainer(state, &packs), "n", tc.conf)

			switch {
			case tc.want == nil && got != nil:
				t.Fatalf("got %v, want nil", *got)
			case tc.want != nil && got == nil:
				t.Fatalf("got nil, want %v", *tc.want)
			case tc.want != nil && *got != *tc.want:
				t.Fatalf("got %v, want %v", *got, *tc.want)
			}
		})
	}
}

// TestNumberFallsBackToTheDefaultWithNothingSent keeps the untouched box
// apart from the refused one: with no value in the state at all the Default
// still stands, bounds or no bounds.
func TestNumberFallsBackToTheDefaultWithNothingSent(t *testing.T) {
	conf := (&tcinput.NumberConf[int]{}).SetMin(10).SetMax(20).SetDefault(12)

	var packs []tgframe.NotifyPack
	got := tcinput.Number(testContainer(tgframe.NewState(), &packs), "n", conf)
	if got == nil || *got != 12 {
		t.Fatalf("got %v, want 12", got)
	}
}

func ptr[T any](v T) *T {
	return &v
}
