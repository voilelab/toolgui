package tgframe_test

import (
	"strings"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// TestVariadicConfAcceptsNoneOrOne covers the two shapes a caller writes.
func TestVariadicConfAcceptsNoneOrOne(t *testing.T) {
	_, ids, err := keysOf(t, func(p *tgframe.Params) error {
		tgcomp.Button(p.Main, "Save")
		tgcomp.Button(p.Main, "Save", &tgcomp.ButtonConf{ID: "second"})

		// An explicit nil is the same as passing nothing.
		tgcomp.Button(p.Main, "Third", nil)
		return nil
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	want := []string{
		"button_component_Save",
		"button_component_second",
		"button_component_Third",
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Errorf("id %d = %q, want %q", i, ids[i], want[i])
		}
	}
}

// TestTwoConfsPanic pins the rule that replaces a merge: two confs for one
// component have no meaning, so passing two is a caller bug, not a value to
// resolve. The panic names the component, so a line calling several of them
// says which one is wrong.
func TestTwoConfsPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("no panic")
		}

		msg, ok := r.(string)
		if !ok {
			t.Fatalf("panicked with %T, want a string", r)
		}

		if !strings.Contains(msg, "Button") {
			t.Errorf("panic %q does not name the component", msg)
		}

		t.Logf("panicked with: %v", msg)
	}()

	_, _, _ = keysOf(t, func(p *tgframe.Params) error {
		tgcomp.Button(p.Main, "Save",
			&tgcomp.ButtonConf{ID: "a"}, &tgcomp.ButtonConf{ID: "b"})
		return nil
	})
}

// TestConfIDIsReadThroughBase checks the framework gets an id off a conf
// without that conf implementing anything: the only thing ButtonConf declares
// is the embed.
func TestConfIDIsReadThroughBase(t *testing.T) {
	// ButtonConf declares no ID of its own, so the flat literal sets the one
	// the embedded Base carries. SetConfID reads nothing else, which is what
	// lets a single framework helper serve every conf.
	conf := &tgcomp.ButtonConf{ID: "from_base"}

	comp := &tgframe.BaseComponent{Name: "probe_component"}
	tgframe.SetConfID(comp, conf)
	if comp.ID != "probe_component_from_base" {
		t.Errorf("comp.ID = %q, want %q", comp.ID, "probe_component_from_base")
	}

	// A conf carrying no id leaves the component's derived id alone.
	untouched := &tgframe.BaseComponent{Name: "probe_component", ID: "derived"}
	tgframe.SetConfID(untouched, &tgcomp.ButtonConf{})
	if untouched.ID != "derived" {
		t.Errorf("comp.ID = %q, want %q", untouched.ID, "derived")
	}
}

// shadowConf is the mistake the Base doc warns about: a conf that declares its
// own ID. It exists only to pin the behaviour — nothing in tgcomp does this.
type shadowConf struct {
	tgframe.Base
	ID int
}

// TestAConfDeclaringItsOwnIDShadowsBaseSilently records that the compiler does
// not object, and the framework reads the empty Base.ID rather than the field
// the caller thought they set. This is why the convention is that no conf may
// declare a field named ID.
func TestAConfDeclaringItsOwnIDShadowsBaseSilently(t *testing.T) {
	conf := &shadowConf{ID: 7}

	if conf.ID != 7 {
		t.Fatalf("outer ID = %d, want 7", conf.ID)
	}
	if conf.Base.ID != "" {
		t.Fatalf("Base.ID = %q, want empty", conf.Base.ID)
	}

	comp := &tgframe.BaseComponent{Name: "probe_component", ID: "derived"}
	tgframe.SetConfID(comp, conf)
	if comp.ID != "derived" {
		t.Errorf("comp.ID = %q, want the derived id: the shadowed literal"+
			" never reached Base", comp.ID)
	}
}

// TestBothLiteralSpellingsCompile pins the two ways to set an id on a conf.
// The flat one needs the caller's file to be at language version 1.27; the
// nested one works at any version and is the fallback a caller below 1.27
// would need.
//
// toolgui's own go.mod is already at go 1.27.1, so a consumer module must
// declare at least that to depend on it at all — the toolchain refuses an
// older one before it ever type-checks a literal. The fallback is therefore
// unreachable in practice, and is pinned here only so that the claim stays
// true if the go directive is ever lowered.
func TestBothLiteralSpellingsCompile(t *testing.T) {
	flat := &tgcomp.ButtonConf{ID: "x", Disabled: true}
	nested := &tgcomp.ButtonConf{Base: tgframe.Base{ID: "x"}, Disabled: true}

	if flat.ID != nested.ID {
		t.Errorf("flat %q, nested %q", flat.ID, nested.ID)
	}
	if flat.Disabled != nested.Disabled {
		t.Error("Disabled differs between the two spellings")
	}
}
