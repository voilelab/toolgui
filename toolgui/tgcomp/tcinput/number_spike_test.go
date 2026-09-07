package tcinput_test

import (
	"encoding/json"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// Rating is a user-defined named type, the case the `~` in the constraint is
// there for.
type Rating int

func spikeContainer(state *tgframe.State, packs *[]tgframe.NotifyPack) *tgframe.Container {
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
		got := tcinput.Number[float64](spikeContainer(state, &packs), "n")
		if got == nil || *got != 2.5 {
			t.Fatalf("got %v, want 2.5", got)
		}
	})

	t.Run("int64", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set(id, 7.0)
		var packs []tgframe.NotifyPack
		got := tcinput.Number[int64](spikeContainer(state, &packs), "n")
		if got == nil || *got != 7 {
			t.Fatalf("got %v, want 7", got)
		}
	})

	// int is the type S5 was actually missing.
	t.Run("int", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set(id, 7.0)
		var packs []tgframe.NotifyPack
		got := tcinput.Number[int](spikeContainer(state, &packs), "n")
		if got == nil || *got != 7 {
			t.Fatalf("got %v, want 7", got)
		}
	})

	t.Run("named type", func(t *testing.T) {
		state := tgframe.NewState()
		state.Set(id, 4.0)
		var packs []tgframe.NotifyPack
		got := tcinput.Number[Rating](spikeContainer(state, &packs), "n")
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
	got := tcinput.Number[int](spikeContainer(state, &packs), "n")
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
			tc.call(spikeContainer(tgframe.NewState(), &packs))

			if len(packs) != 1 {
				t.Fatalf("got %d packs, want 1", len(packs))
			}

			bs, err := json.Marshal(packs[0])
			if err != nil {
				t.Fatal(err)
			}

			var got struct {
				Component struct {
					Step json.RawMessage `json:"step"`
				} `json:"component"`
			}
			if err := json.Unmarshal(bs, &got); err != nil {
				t.Fatal(err)
			}

			if string(got.Component.Step) != tc.want {
				t.Errorf("step = %q, want %q", got.Component.Step, tc.want)
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
	tcinput.Number(spikeContainer(tgframe.NewState(), &packs), "n", conf)

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
	c := spikeContainer(state, &packs)

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
	if conf.Base.ID != "count" {
		t.Fatalf("Base.ID = %q, want %q", conf.Base.ID, "count")
	}

	state := tgframe.NewState()
	state.Set("number_component_count", 5.0)

	var packs []tgframe.NotifyPack
	got := tcinput.Number(spikeContainer(state, &packs), "n", conf)
	if got == nil || *got != 5 {
		t.Fatalf("got %v, want 5 read under the conf's id", got)
	}
}
