package tgframe

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgjson"
)

// interruptWait bounds how long a cut run may take to unwind.
const interruptWait = time.Second

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

	// The cut run reports nothing: the new run's ready pack already told the
	// client to drop what was on the screen.
	second := <-recorder.results
	if !second.Success {
		t.Fatalf("expect the second run to succeed, got %q", second.Error)
	}

	if got := recorder.count(); got >= spinCount {
		t.Fatalf("expect the first run to be cut short, got %d packs", got)
	}
}

// TestSessionInterruptsPageBlockedOnContext checks a page func that draws
// nothing is still cut, through its context, instead of holding the next
// event until it finishes on its own.
func TestSessionInterruptsPageBlockedOnContext(t *testing.T) {
	var runs atomic.Int32
	started := make(chan struct{}, 1)
	unwound := make(chan struct{})
	reran := make(chan struct{})

	session, _ := newTestSession(t, func(p *Params) error {
		if runs.Add(1) > 1 {
			close(reran)
			return nil
		}

		started <- struct{}{}
		<-p.Context.Done()
		close(unwound)
		return p.Context.Err()
	})
	defer session.Close()

	session.HandleEvent(&EventEmpty{})
	<-started

	handled := make(chan struct{})
	go func() {
		session.HandleEvent(&EventEmpty{})
		close(handled)
	}()

	select {
	case <-handled:
	case <-time.After(interruptWait):
		t.Fatal("the blocked run was not cut")
	}

	select {
	case <-unwound:
	case <-time.After(interruptWait):
		t.Fatal("the blocked run did not return")
	}

	select {
	case <-reran:
	case <-time.After(interruptWait):
		t.Fatal("the new run did not start")
	}
}

// TestSessionInterruptedRunSendsNoResult checks a cut run stays quiet: the
// context.Canceled it comes back with is how it unwound, not a failure the
// client should hear about.
func TestSessionInterruptedRunSendsNoResult(t *testing.T) {
	var runs atomic.Int32
	started := make(chan struct{}, 1)

	session, recorder := newTestSession(t, func(p *Params) error {
		if runs.Add(1) > 1 {
			addTestComponent(p, "comp")
			return nil
		}

		started <- struct{}{}
		<-p.Context.Done()
		return p.Context.Err()
	})
	defer session.Close()

	session.HandleEvent(&EventEmpty{})
	<-started

	// The cut run sends its result, if any, before this returns.
	session.HandleEvent(&EventEmpty{})

	result := <-recorder.results
	if !result.Success {
		t.Fatalf("expect the second run's result first, got %q", result.Error)
	}
}

// TestSessionInterruptedRunKeepsComponentIDs checks a cut run does not take
// the screen's component ids with it. The run before it is still what the
// client is looking at, and a click or an upload naming one of its components
// has to keep passing State.HasComponentID.
func TestSessionInterruptedRunKeepsComponentIDs(t *testing.T) {
	var runs atomic.Int32
	started := make(chan struct{}, 1)

	session, recorder := newTestSession(t, func(p *Params) error {
		if runs.Add(1) == 1 {
			addTestComponent(p, "comp")
			return nil
		}

		// Cut before drawing anything, and return rather than panic.
		started <- struct{}{}
		<-p.Context.Done()
		return p.Context.Err()
	})
	defer session.Close()

	session.HandleEvent(&EventEmpty{})
	if result := <-recorder.results; !result.Success {
		t.Fatalf("expect the first run to succeed, got %q", result.Error)
	}

	session.HandleEvent(&EventEmpty{})
	<-started

	// Returns once the cut run has unwound.
	session.HandleEvent(&EventEmpty{})

	if !session.state.HasComponentID("comp") {
		t.Fatal("expect the ids of the last finished run to stay on the screen")
	}
}

// TestSessionCloseCancelsRunContext checks Close cuts the run in flight
// through its context, not only at the next thing it draws.
func TestSessionCloseCancelsRunContext(t *testing.T) {
	started := make(chan struct{}, 1)
	unwound := make(chan error, 1)

	session, _ := newTestSession(t, func(p *Params) error {
		started <- struct{}{}
		<-p.Context.Done()
		unwound <- p.Context.Err()
		return nil
	})

	session.HandleEvent(&EventEmpty{})
	<-started

	closed := make(chan struct{})
	go func() {
		session.Close()
		close(closed)
	}()

	select {
	case <-closed:
	case <-time.After(interruptWait):
		t.Fatal("Close did not cut the running page")
	}

	select {
	case err := <-unwound:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expect a cancelled context, got %v", err)
		}
	default:
		t.Fatal("expect the page func to see its context done")
	}
}

// TestRunWithHandlingPanicKeepsErrorChain checks a panicked error keeps its
// chain, so an interrupt is still recognisable under ErrPanic.
func TestRunWithHandlingPanicKeepsErrorChain(t *testing.T) {
	app := NewApp()
	app.AddPage(testPageName, "Test", func(p *Params) error {
		addTestComponent(p, "comp")
		return nil
	})

	err := app.RunWithHandlingPanic(testPageName, NewState(),
		func(pack NotifyPack) { panic(ErrUpdateInterrupt) })

	if !errors.Is(err, ErrPanic) {
		t.Fatalf("expect ErrPanic, got %v", err)
	}

	if !errors.Is(err, ErrUpdateInterrupt) {
		t.Fatalf("expect ErrUpdateInterrupt in the chain, got %v", err)
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

// TestResultPackOmitsZeroFields pins what `omitzero` buys a client: a
// successful run reports `success` and nothing else. Under `omitempty` a
// false bool is written, so every result would carry a `"fatal":false` the
// frontend never asked for.
func TestResultPackOmitsZeroFields(t *testing.T) {
	for _, tc := range []struct {
		name string
		pack *ResultPack
		want string
	}{
		{"success", &ResultPack{Success: true}, `{"success":true}`},
		{"error", &ResultPack{Error: "boom"}, `{"error":"boom","success":false}`},
		{"fatal", &ResultPack{Error: "no page", Fatal: true}, `{"error":"no page","success":false,"fatal":true}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bs, err := tgjson.Marshal(tc.pack)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			if string(bs) != tc.want {
				t.Errorf("marshal = %s, want %s", bs, tc.want)
			}
		})
	}
}
