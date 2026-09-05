package tgframe

import (
	"sync"
	"sync/atomic"
	"testing"
)

const testPageName = "test"

// packRecorder is a SendPackFunc that records packs and reports whether it
// was ever called by two goroutines at once.
type packRecorder struct {
	lock    sync.Mutex
	packs   []any
	results chan *ResultPack

	inSend     atomic.Bool
	concurrent atomic.Bool
}

func newPackRecorder() *packRecorder {
	return &packRecorder{
		results: make(chan *ResultPack, 16),
	}
}

func (r *packRecorder) send(pack any) error {
	if !r.inSend.CompareAndSwap(false, true) {
		r.concurrent.Store(true)
	}
	defer r.inSend.Store(false)

	r.lock.Lock()
	r.packs = append(r.packs, pack)
	r.lock.Unlock()

	if result, ok := pack.(*ResultPack); ok {
		// Never block: a blocked send would stop the run from seeing the
		// interrupt signal.
		select {
		case r.results <- result:
		default:
		}
	}

	return nil
}

func (r *packRecorder) count() int {
	r.lock.Lock()
	defer r.lock.Unlock()
	return len(r.packs)
}

func newTestSession(t *testing.T, runFunc RunFunc) (*Session, *packRecorder) {
	t.Helper()

	app := NewApp()
	app.AddPage(testPageName, "Test", runFunc)

	recorder := newPackRecorder()
	session, err := NewSession(app, testPageName, NewState(), recorder.send)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	return session, recorder
}

func addTestComponent(p *Params, id string) {
	p.Main.AddComponent(&BaseComponent{Name: "test_component", ID: id})
}

func TestNewSessionUnknownPage(t *testing.T) {
	app := NewApp()
	_, err := NewSession(app, "nope", NewState(), func(any) error { return nil })
	if err == nil {
		t.Fatal("expect an error for an unknown page")
	}
}

func TestSessionHandleEventRunsPage(t *testing.T) {
	session, recorder := newTestSession(t, func(p *Params) error {
		addTestComponent(p, "comp")
		return nil
	})
	defer session.Close()

	session.HandleEvent(&EventEmpty{})

	result := <-recorder.results
	if !result.Success {
		t.Fatalf("expect a successful run, got %q", result.Error)
	}

	// ready pack, one create pack, result pack
	if got := recorder.count(); got != 3 {
		t.Fatalf("expect 3 packs, got %d", got)
	}
}

// TestSessionInterruptsRunningPage checks a new event cuts the running page
// func instead of waiting for it to finish.
func TestSessionInterruptsRunningPage(t *testing.T) {
	const spinCount = 100000

	var runs atomic.Int32
	started := make(chan struct{}, 1)

	session, recorder := newTestSession(t, func(p *Params) error {
		if runs.Add(1) > 1 {
			return nil
		}

		// Keep sending until an event interrupts us.
		started <- struct{}{}
		for i := 0; i < spinCount; i++ {
			addTestComponent(p, "spin")
		}
		return nil
	})
	defer session.Close()

	session.HandleEvent(&EventEmpty{})
	<-started

	// Blocks until the first run is interrupted and unwinds.
	session.HandleEvent(&EventEmpty{})

	first := <-recorder.results
	if first.Success {
		t.Fatal("expect the interrupted run to report an error")
	}

	second := <-recorder.results
	if !second.Success {
		t.Fatalf("expect the second run to succeed, got %q", second.Error)
	}

	if got := recorder.count(); got >= spinCount {
		t.Fatalf("expect the first run to be cut short, got %d packs", got)
	}
}

// TestSessionSerializesSends checks a transport is never called by the runner
// goroutine and an event handler at the same time.
func TestSessionSerializesSends(t *testing.T) {
	const spinCount = 100000

	started := make(chan struct{}, 1)

	session, recorder := newTestSession(t, func(p *Params) error {
		started <- struct{}{}
		for i := 0; i < spinCount; i++ {
			addTestComponent(p, "spin")
		}
		return nil
	})

	session.HandleEvent(&EventEmpty{})
	<-started

	// Parse errors report to the client while the run is still sending.
	for i := 0; i < 100; i++ {
		err := session.HandleRawEvent([]byte(`{"type":"unknown"}`))
		if err == nil {
			t.Fatal("expect a parse error")
		}
	}

	session.Close()

	if recorder.concurrent.Load() {
		t.Fatal("the transport was called concurrently")
	}
}

func TestSessionClosedIgnoresEvents(t *testing.T) {
	session, recorder := newTestSession(t, func(p *Params) error {
		addTestComponent(p, "comp")
		return nil
	})

	session.Close()

	session.HandleEvent(&EventEmpty{})
	if err := session.HandleRawEvent([]byte(`{"type":"unknown"}`)); err != nil {
		t.Fatalf("expect a closed session to drop the event, got %v", err)
	}

	if got := recorder.count(); got != 0 {
		t.Fatalf("expect no pack after Close, got %d", got)
	}
}
