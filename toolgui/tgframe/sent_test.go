package tgframe

import (
	"errors"
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgjson"
)

type sentTestComp struct {
	*BaseComponent
	Text string `json:"text"`
}

func newSentTestComp(id, text string) *sentTestComp {
	return &sentTestComp{
		BaseComponent: &BaseComponent{Name: "sent_test_component", ID: id},
		Text:          text,
	}
}

// wireTypes is the type of every notify pack recorded, by key, in order.
func wireTypes(t *testing.T, packs []any) [][2]any {
	t.Helper()

	var out [][2]any
	for _, p := range packs {
		if _, ok := p.(NotifyPack); !ok {
			continue
		}

		bs, err := tgjson.Marshal(p)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		var head struct {
			Type int    `json:"type"`
			Key  string `json:"key"`
		}
		if err := tgjson.Unmarshal(bs, &head); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		out = append(out, [2]any{head.Key, head.Type})
	}

	return out
}

// runPacks runs the session once and returns the notify packs of that run.
func runPacks(t *testing.T, s *Session, r *packRecorder) [][2]any {
	t.Helper()

	before := r.count()
	s.HandleEvent(&EventEmpty{})
	waitResult(t, r)

	r.lock.Lock()
	defer r.lock.Unlock()
	return wireTypes(t, r.packs[before:])
}

func expectPacks(t *testing.T, got [][2]any, want ...[2]any) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("packs = %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("packs = %v, want %v", got, want)
		}
	}
}

const sentMain = "container_component_container_main"

func TestSessionKeepsUnchangedComponents(t *testing.T) {
	text := "a"
	s, r := newTestSession(t, func(p *Params) error {
		p.Main.AddComponent(newSentTestComp("", "fixed"))
		p.Main.AddComponent(newSentTestComp("", text))
		return nil
	})
	t.Cleanup(s.Close)

	k0, k1 := sentMain+"/0", sentMain+"/1"

	expectPacks(t, runPacks(t, s, r),
		[2]any{k0, NotifyTypeCreate}, [2]any{k1, NotifyTypeCreate})

	expectPacks(t, runPacks(t, s, r),
		[2]any{k0, NotifyTypeKeep}, [2]any{k1, NotifyTypeKeep})

	text = "b"
	expectPacks(t, runPacks(t, s, r),
		[2]any{k0, NotifyTypeKeep}, [2]any{k1, NotifyTypeCreate})
}

// A keep pack carries the place, and a create the same bytes as before.
func TestSessionPackWire(t *testing.T) {
	s, r := newTestSession(t, func(p *Params) error {
		p.Main.AddComponent(newSentTestComp("x", "hi"))
		return nil
	})
	t.Cleanup(s.Close)

	runPacks(t, s, r)
	runPacks(t, s, r)

	r.lock.Lock()
	defer r.lock.Unlock()

	var wires []string
	for _, p := range r.packs {
		if _, ok := p.(NotifyPack); ok {
			bs, _ := tgjson.Marshal(p)
			wires = append(wires, string(bs))
		}
	}

	want := []string{
		`{"type":1,"parent_key":"container_component_container_main","index":0,` +
			`"key":"container_component_container_main/0",` +
			`"component":{"name":"sent_test_component","id":"x","text":"hi"}}`,
		`{"type":4,"parent_key":"container_component_container_main","index":0,` +
			`"key":"container_component_container_main/0"}`,
	}
	if len(wires) != len(want) {
		t.Fatalf("wires = %v", wires)
	}
	for i := range want {
		if wires[i] != want[i] {
			t.Errorf("wire %d = %s, want %s", i, wires[i], want[i])
		}
	}
}

// A new session is a new page or a reconnect: the client may hold nothing.
func TestNewSessionSendsAll(t *testing.T) {
	run := func(p *Params) error {
		p.Main.AddComponent(newSentTestComp("", "x"))
		return nil
	}

	for range 2 {
		s, r := newTestSession(t, run)
		expectPacks(t, runPacks(t, s, r), [2]any{sentMain + "/0", NotifyTypeCreate})
		s.Close()
	}
}

// What a run did not send is gone once it succeeds, so writing it again is a
// create.
func TestSessionSendsAgainAfterGone(t *testing.T) {
	show := true
	s, r := newTestSession(t, func(p *Params) error {
		if show {
			p.Main.AddComponent(newSentTestComp("", "x"))
		}
		return nil
	})
	t.Cleanup(s.Close)

	k := sentMain + "/0"
	expectPacks(t, runPacks(t, s, r), [2]any{k, NotifyTypeCreate})

	show = false
	expectPacks(t, runPacks(t, s, r))

	show = true
	expectPacks(t, runPacks(t, s, r), [2]any{k, NotifyTypeCreate})
}

// A failed run leaves the client tree alone, so what it skipped is still kept.
func TestSessionFailedRunKeepsTree(t *testing.T) {
	fail := false
	s, r := newTestSession(t, func(p *Params) error {
		if fail {
			return errors.New("boom")
		}
		p.Main.AddComponent(newSentTestComp("", "x"))
		return nil
	})
	t.Cleanup(s.Close)

	k := sentMain + "/0"
	runPacks(t, s, r)

	fail = true
	expectPacks(t, runPacks(t, s, r))

	fail = false
	expectPacks(t, runPacks(t, s, r), [2]any{k, NotifyTypeKeep})
}

// A cleared slot is gone from the client, contents and all.
func TestSessionSlotSendsAgain(t *testing.T) {
	s, r := newTestSession(t, func(p *Params) error {
		box := p.Main.AddComponent(newSentTestComp("", "box"))
		slot := p.Main.AddSlotTo(box, "slot", 0)
		slot.With(func(c *Container) {
			c.AddComponent(newSentTestComp("", "x"))
		})
		return nil
	})
	t.Cleanup(s.Close)

	runPacks(t, s, r)
	got := runPacks(t, s, r)

	box := sentMain + "/0"
	expectPacks(t, got,
		[2]any{box, NotifyTypeKeep},
		[2]any{box + "/0", NotifyTypeDelete},
		[2]any{box + "/0", NotifyTypeCreate},
		[2]any{box + "/0/0", NotifyTypeCreate})
}

func TestSentCacheUpdateDropsHash(t *testing.T) {
	c := newSentCache()
	comp := newSentTestComp("", "x")
	comp.setKey("k")

	c.beginRun()
	c.create("k", "", []byte(`{}`))
	c.filter(NewNotifyPackUpdate(comp))

	c.beginRun()
	if c.create("k", "", []byte(`{}`)) {
		t.Error("kept a node an update changed")
	}
}

func TestSentCacheDeleteDropsSubtree(t *testing.T) {
	c := newSentCache()
	c.beginRun()
	for _, k := range []string{"a", "a/0", "a/0/1", "ab"} {
		c.create(k, "", []byte(`{}`))
	}

	c.remove("a")

	if len(c.nodes) != 1 || c.nodes["ab"] == nil {
		t.Errorf("nodes = %v, want only ab", c.nodes)
	}
}

// The client retires the old place of a named component that moved, with its
// subtree, so what sat under the old place has to be sent again.
func TestSentCacheMovedIDDropsOldPlace(t *testing.T) {
	c := newSentCache()
	c.beginRun()
	c.create("m/1", "box", []byte(`{"id":"box"}`))
	c.create("m/1/0", "", []byte(`{"t":1}`))
	c.endRun()

	c.beginRun()
	c.create("m/0", "box", []byte(`{"id":"box"}`))

	if c.nodes["m/1"] != nil || c.nodes["m/1/0"] != nil {
		t.Fatalf("old place kept: %v", c.nodes)
	}
	if c.create("m/1/0", "", []byte(`{"t":1}`)) {
		t.Error("kept a node the client retired")
	}
	if len(c.byID["box"]) != 1 {
		t.Errorf("byID = %v", c.byID)
	}
}

// A create the transport failed to send is not kept next run.
func TestSessionFailedSendIsNotKept(t *testing.T) {
	app := NewApp()
	app.AddPage(testPageName, "Test", func(p *Params) error {
		p.Main.AddComponent(newSentTestComp("", "x"))
		return nil
	})

	fail := true
	r := newPackRecorder()
	s, err := NewSession(app, testPageName, nil, NewState(), func(pack any) error {
		if _, ok := pack.(*notifyPackCreateRaw); ok && fail {
			fail = false
			r.send(&ResultPack{Error: "send"})
			return errors.New("send")
		}
		return r.send(pack)
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)

	runPacks(t, s, r)
	waitResult(t, r)
	expectPacks(t, runPacks(t, s, r), [2]any{sentMain + "/0", NotifyTypeCreate})
}
