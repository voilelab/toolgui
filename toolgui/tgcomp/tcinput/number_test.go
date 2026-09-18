package tcinput_test

import (
	"encoding/json/jsontext"
	"math"
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
// number, so [tgframe.State.GetNumber] reads a float64 whatever T is.
func TestNumberAcceptsEveryTypeInTheSet(t *testing.T) {
	const id = "number_component_n"

	t.Run("float64", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set(id, 2.5)
		var packs []tgframe.NotifyPack
		got, _ := tcinput.Number[float64](testContainer(state, &packs), "n")
		if got != 2.5 {
			t.Fatalf("got %v, want 2.5", got)
		}
	})

	t.Run("int64", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set(id, 7.0)
		var packs []tgframe.NotifyPack
		got, _ := tcinput.Number[int64](testContainer(state, &packs), "n")
		if got != 7 {
			t.Fatalf("got %v, want 7", got)
		}
	})

	// int is the type S5 was actually missing.
	t.Run("int", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set(id, 7.0)
		var packs []tgframe.NotifyPack
		got, _ := tcinput.Number[int](testContainer(state, &packs), "n")
		if got != 7 {
			t.Fatalf("got %v, want 7", got)
		}
	})

	t.Run("named type", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set(id, 4.0)
		var packs []tgframe.NotifyPack
		got, _ := tcinput.Number[Rating](testContainer(state, &packs), "n")
		if got != Rating(4) {
			t.Fatalf("got %v, want Rating(4)", got)
		}
	})
}

// TestNumberTruncatesTowardsTheIntegralType records what an integral T does
// with the fractional float the state can hold: it truncates and it does not
// round. What it no longer does is keep quiet about it -- nothing on the wire
// says T is integral, so the client's number box takes a decimal whatever T
// is, and a page that saved the 2 would be saving a number nobody typed.
func TestNumberTruncatesTowardsTheIntegralType(t *testing.T) {
	state := tgframe.NewState()
	state.Set("number_component_n", 2.9)

	var packs []tgframe.NotifyPack
	got, ok := tcinput.Number[int](testContainer(state, &packs), "n")
	if got != 2 {
		t.Fatalf("got %v, want 2", got)
	}
	if ok {
		t.Error("ok = true for a truncated 2.9, want false")
	}

	// A float64 T holds it as it is, so there is nothing to report.
	state = tgframe.NewState()
	state.Set("number_component_n", 2.9)

	packs = nil
	if got, ok := tcinput.Number[float64](
		testContainer(state, &packs), "n"); got != 2.9 || !ok {

		t.Errorf("float64 T = %v, %v, want 2.9, true", got, ok)
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
	if got, _ := tcinput.Number[int](c, "n"); got != 3 {
		t.Errorf("Number[int] = %v, want 3", got)
	}
	if got, _ := tcinput.Number[Rating](c, "n"); got != Rating(3) {
		t.Errorf("Number[Rating] = %v, want Rating(3)", got)
	}

	// And with a conf, T can instead be inferred from the conf alone.
	got, _ := tcinput.Number(c, "n", &tcinput.NumberConf[int64]{})
	if got != 3 {
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
	got, _ := tcinput.Number(testContainer(state, &packs), "n", conf)
	if got != 5 {
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

// TestNumberAppliesTheRangeToWhatArrives pins the answer to the bug this is
// for: the client sends whatever the app user typed, so a value outside
// Min/Max arrives and is pulled to the bound. Leaving the last value that was
// inside the range is what let a page save a number nobody had entered.
//
// Pulling it to the bound is only half of that: 20 reads the same whether it
// was typed or clamped, so wantOK pins the signal that tells them apart. It is
// false exactly when the value returned is not the one that arrived -- pulled
// to a bound, or a float T cannot hold.
func TestNumberAppliesTheRangeToWhatArrives(t *testing.T) {
	const id = "number_component_n"

	bounded := func() *tcinput.NumberConf[int] {
		return (&tcinput.NumberConf[int]{}).SetMin(10).SetMax(20)
	}

	for _, tc := range []struct {
		name   string
		sent   float64
		conf   *tcinput.NumberConf[int]
		want   int
		wantOK bool
	}{
		{"below min", 9, bounded(), 10, false},
		{"above max", 21, bounded(), 20, false},
		{"at min", 10, bounded(), 10, true},
		{"at max", 20, bounded(), 20, true},
		{"within", 15, bounded(), 15, true},

		// The Default is not where an out-of-range value falls back to: it is
		// as stale as the last value inside the range, and the box is not
		// showing it either.
		{"out of range with a default", 99, &tcinput.NumberConf[int]{Default: 12, Min: ptr(10), Max: ptr(20)}, 20, false},

		// Only the bound that is set applies, and only a bound that applies
		// can make the value anything but the one that arrived.
		{"min only", 99, (&tcinput.NumberConf[int]{}).SetMin(10), 99, true},
		{"max only", -99, (&tcinput.NumberConf[int]{}).SetMax(20), -99, true},

		// No bounds, nothing to apply: this is the Number that behaves as it
		// always did, and nothing an int holds can be out of a range that was
		// never set.
		{"no bounds", 9999, &tcinput.NumberConf[int]{}, 9999, true},

		// The bounds are compared before the truncation, on the number that
		// was typed: 20.9 is over a Max of 20, so it comes back as the bound
		// rather than as the 20 it would have truncated to.
		{"truncated at the bound", 20.9, bounded(), 20, false},

		// And under a Max of 21 the same 20.9 is in range, so it truncates --
		// and the truncation is reported, because 20 is not what was typed
		// any more than a clamped 20 would have been. The value handed over
		// is still the 20, not the Default.
		{"truncated inside the range", 20.9, (&tcinput.NumberConf[int]{}).SetMax(21), 20, false},
		{"truncated, no bounds", 20.9, &tcinput.NumberConf[int]{Default: 7}, 20, false},

		// A whole number that arrived as a float is not truncated at all.
		{"whole float", 20.0, (&tcinput.NumberConf[int]{}).SetMax(21), 20, true},

		// A float an integral T cannot hold converts to an
		// implementation-defined value -- on amd64, math.MinInt. Comparing
		// the bounds first is what keeps that number out of the result: it is
		// below every Min and above no Max, so converting first would slip it
		// past a conf with only a Max.
		{"beyond the type", 1e20, bounded(), 20, false},
		{"beyond the type, max only", 1e20, (&tcinput.NumberConf[int]{}).SetMax(20), 20, false},
		{"beyond the type, min only", -1e20, (&tcinput.NumberConf[int]{}).SetMin(10), 10, false},

		// With no bound to be pulled to there is nothing to report, so the
		// Default stands -- and a Default nobody typed is exactly what the
		// signal is for, bounds or no bounds.
		{"beyond the type, no bounds", 1e20, &tcinput.NumberConf[int]{Default: 7}, 7, false},

		// The far edge of what it can hold is still a value.
		{"at the type's edge", math.MinInt, &tcinput.NumberConf[int]{}, math.MinInt, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := tgframe.NewState()
			state.Set(id, tc.sent)

			var packs []tgframe.NotifyPack
			got, ok := tcinput.Number(testContainer(state, &packs), "n", tc.conf)
			if got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			if ok != tc.wantOK {
				t.Errorf("ok = %v, want %v", ok, tc.wantOK)
			}
		})
	}
}

// TestNumberFallsBackToTheDefaultWithNothingSent keeps the untouched box
// apart from the clamped one: with no value in the state at all the Default
// stands, bounds or no bounds.
func TestNumberFallsBackToTheDefaultWithNothingSent(t *testing.T) {
	conf := (&tcinput.NumberConf[int]{Default: 12}).SetMin(10).SetMax(20)

	var packs []tgframe.NotifyPack
	got, ok := tcinput.Number(testContainer(tgframe.NewState(), &packs), "n", conf)
	if got != 12 {
		t.Fatalf("got %v, want 12", got)
	}

	// An untouched box is not an invalid one: the Default is the answer the
	// app user is looking at, so there is nothing for a page to refuse.
	if !ok {
		t.Error("ok = false for an untouched box, want true")
	}
}

// TestNumberWithAFloatTheTypeCannotHold covers what the table above cannot: a
// float64 T holds everything a float64 can, and an int64 has a different edge
// from the int the table uses.
func TestNumberWithAFloatTheTypeCannotHold(t *testing.T) {
	const id = "number_component_n"

	read := func(
		sent float64,
		call func(*tgframe.Container) (float64, bool)) (float64, bool) {

		state := tgframe.NewState()
		state.Set(id, sent)

		var packs []tgframe.NotifyPack
		return call(testContainer(state, &packs))
	}

	// Default 7, so falling back to it is visible rather than being the zero
	// everything else could also be.
	asFloat := func(c *tgframe.Container) (float64, bool) {
		return tcinput.Number(c, "n", &tcinput.NumberConf[float64]{Default: 7})
	}
	asInt64 := func(c *tgframe.Container) (float64, bool) {
		v, ok := tcinput.Number(c, "n", &tcinput.NumberConf[int64]{Default: 7})
		return float64(v), ok
	}

	// NaN is nobody's value, whatever T is. It cannot arrive as JSON, but the
	// state is not only fed by the client.
	// Falling back to the Default is reported: 7 is not the number that was
	// sent, and a page that stores it stores something nobody entered.
	if got, ok := read(math.NaN(), asFloat); got != 7 || ok {
		t.Errorf("NaN for a float64 T = %v, %v, want the default 7, false", got, ok)
	}
	if got, ok := read(math.NaN(), asInt64); got != 7 || ok {
		t.Errorf("NaN for an int64 T = %v, %v, want the default 7, false", got, ok)
	}

	// An infinity is a float64, so only the integral T has to refuse it.
	if got, ok := read(math.Inf(1), asFloat); !math.IsInf(got, 1) || !ok {
		t.Errorf("+Inf for an unbounded float64 T = %v, %v, want +Inf, true", got, ok)
	}
	if got, ok := read(math.Inf(1), asInt64); got != 7 || ok {
		t.Errorf("+Inf for an int64 T = %v, %v, want the default 7, false", got, ok)
	}

	// 1e20 is a perfectly good float64 and no int64 at all.
	if got, ok := read(1e20, asFloat); got != 1e20 || !ok {
		t.Errorf("1e20 for a float64 T = %v, %v, want 1e20, true", got, ok)
	}
	if got, ok := read(1e20, asInt64); got != 7 || ok {
		t.Errorf("1e20 for an int64 T = %v, %v, want the default 7, false", got, ok)
	}
}

// TestNumberSignalsOnlyWhatTheAppUserCouldNotEnter is the shape TG-72 settled
// on, stated on its own rather than as a column: the second return is there so
// a page can refuse to act, and a page that refuses too often is as wrong as
// one that never does.
func TestNumberSignalsOnlyWhatTheAppUserCouldNotEnter(t *testing.T) {
	const id = "number_component_n"

	read := func(sent float64, conf *tcinput.NumberConf[int]) (int, bool) {
		state := tgframe.NewState()
		state.Set(id, sent)

		var packs []tgframe.NotifyPack
		return tcinput.Number(testContainer(state, &packs), "n", conf)
	}

	// With no Min or Max there is no range to be outside of, so every whole
	// number an int holds is the app user's own -- including the ones a
	// bounded Number would have refused. math.MaxInt is not in the list: as a
	// float64 it rounds up to 2^63, which is one past what an int holds, so it
	// belongs with the values below rather than here.
	for _, sent := range []float64{0, -9999, 9999, math.MinInt} {
		if got, ok := read(sent, &tcinput.NumberConf[int]{}); !ok {
			t.Errorf("unbounded %v = %v, false; want true", sent, got)
		}
	}

	// The value handed over stays inside the range even when the signal is
	// false: the signal is the extra, not a replacement. Returning the 999 as
	// it arrived would put the implementation-defined conversion back.
	conf := (&tcinput.NumberConf[int]{}).SetMin(0).SetMax(24)
	for _, sent := range []float64{-1, 999, 1e20, -1e20, 12.5} {
		got, ok := read(sent, conf)
		if ok {
			t.Errorf("%v against [0, 24] = true, want false", sent)
		}
		if got < 0 || got > 24 {
			t.Errorf("%v against [0, 24] = %v, want it inside the range", sent, got)
		}
	}
}
