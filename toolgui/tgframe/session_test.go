package tgframe

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
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

		// Keep sending until an event interrupts us. An id of its own per
		// component: two sharing one fails the run on a duplicated id, and a
		// first run that got all the way through would report that failure
		// where the test below reads the second run's result.
		started <- struct{}{}
		for i := 0; i < spinCount; i++ {
			addTestComponent(p, fmt.Sprintf("spin-%d", i))

			// js/wasm has no async preemption: a run that never yields keeps
			// the one thread, so the event meant to cut this one could not
			// arrive and the run would finish instead of being cut.
			runtime.Gosched()
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

// TestSessionUnknownComponentIDRejected checks an event writing under an id
// the page never drew is turned away: nothing lands in the state, the client
// is told why, and the page is not rerun for it.
func TestSessionUnknownComponentIDRejected(t *testing.T) {
	var runs atomic.Int32

	session, recorder := newTestSession(t, func(p *Params) error {
		runs.Add(1)
		addTestComponent(p, "comp")
		return nil
	})
	defer session.Close()

	session.HandleEvent(&EventEmpty{})
	if result := <-recorder.results; !result.Success {
		t.Fatalf("expect the first run to succeed, got %q", result.Error)
	}

	for _, event := range []Event{
		&EventInput{ID: "made_up", Value: "typed"},
		&EventSelect{ID: "made_up", Value: 1},
		&EventSelect{ID: "made_up", Values: []int{1}},
		&EventCustom{ID: "made_up", Value: "sent"},
		&EventForm{Events: []Event{
			&EventInput{ID: "comp", Value: "fine"},
			&EventInput{ID: "made_up", Value: "not fine"},
		}},
	} {
		session.HandleEvent(event)

		result := <-recorder.results
		if result.Success {
			t.Fatalf("%T: expect the event to be rejected", event)
		}

		if !strings.Contains(result.Error, "made_up") {
			t.Errorf("%T: expect the id in the error, got %q", event, result.Error)
		}
	}

	if _, ok := session.state.Get[string]("made_up"); ok {
		t.Error("expect a rejected event to write nothing")
	}

	// The form was rejected whole, so the event beside the unknown one did
	// not land either.
	if _, ok := session.state.Get[string]("comp"); ok {
		t.Error("expect a rejected form to write none of its events")
	}

	if got := runs.Load(); got != 1 {
		t.Errorf("expect no rerun for a rejected event, ran %d times", got)
	}
}

// TestSessionKnownComponentIDApplied checks the whitelist only turns away what
// the page is not showing: the components it drew write their state as before.
func TestSessionKnownComponentIDApplied(t *testing.T) {
	session, recorder := newTestSession(t, func(p *Params) error {
		addTestComponent(p, "text")
		addTestComponent(p, "choice")
		addTestComponent(p, "guest")
		return nil
	})
	defer session.Close()

	session.HandleEvent(&EventEmpty{})
	if result := <-recorder.results; !result.Success {
		t.Fatalf("expect the first run to succeed, got %q", result.Error)
	}

	session.HandleEvent(&EventForm{Events: []Event{
		&EventInput{ID: "text", Value: "typed"},
		&EventSelect{ID: "choice", Value: 2},
		&EventCustom{ID: "guest", Value: "sent"},
	}})

	if result := <-recorder.results; !result.Success {
		t.Fatalf("expect the form to be applied, got %q", result.Error)
	}

	if got, ok := session.state.Get[string]("text"); !ok || got != "typed" {
		t.Errorf("text = %v, %v, want %q", got, ok, "typed")
	}

	if got, ok := session.state.Get[int]("choice"); !ok || got != 2 {
		t.Errorf("choice = %v, %v, want 2", got, ok)
	}

	if got, ok := session.state.Get[string]("guest"); !ok || got != "sent" {
		t.Errorf("guest = %v, %v, want %q", got, ok, "sent")
	}
}

// TestSessionEventBeforeFirstRunRejected pins what happens to an event that
// names an id before anything has been drawn: the page is showing nothing, so
// there is no id to accept. The rerun event a client opens with names none, so
// it goes through and the draw it asks for is what makes the ids known.
func TestSessionEventBeforeFirstRunRejected(t *testing.T) {
	session, recorder := newTestSession(t, func(p *Params) error {
		addTestComponent(p, "comp")
		return nil
	})
	defer session.Close()

	session.HandleEvent(&EventInput{ID: "comp", Value: "typed"})
	if result := <-recorder.results; result.Success {
		t.Fatal("expect an event before the first draw to be rejected")
	}

	session.HandleEvent(&EventEmpty{})
	if result := <-recorder.results; !result.Success {
		t.Fatalf("expect the rerun event to go through, got %q", result.Error)
	}

	session.HandleEvent(&EventInput{ID: "comp", Value: "typed"})
	if result := <-recorder.results; !result.Success {
		t.Fatalf("expect the event after the draw to be applied, got %q", result.Error)
	}

	if got, ok := session.state.Get[string]("comp"); !ok || got != "typed" {
		t.Errorf("comp = %v, %v, want %q", got, ok, "typed")
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

// TestSessionMasksPanicContent checks what a page panicked with never leaves
// the process. A panic value carries whatever the app was holding when it
// broke -- a file path, a query, a connection string -- and the browser, which
// nothing authenticates, is the one place it must not go.
func TestSessionMasksPanicContent(t *testing.T) {
	const secret = "user=admin password=hunter2"

	session, recorder := newTestSession(t, func(p *Params) error {
		panic(errors.New(secret))
	})
	defer session.Close()

	session.HandleEvent(&EventEmpty{})

	result := <-recorder.results
	if result.Success {
		t.Fatalf("expect a failure, got %#v", result)
	}

	if strings.Contains(result.Error, secret) {
		t.Errorf("Error = %q, carries the panic value", result.Error)
	}

	// The package path the framework's own errors are prefixed with is as
	// much of the server's inside as the panic value is.
	if strings.Contains(result.Error, "tgframe") {
		t.Errorf("Error = %q, carries a function path", result.Error)
	}

	if result.Error != InternalErrorMessage {
		t.Errorf("Error = %q, want %q", result.Error, InternalErrorMessage)
	}

	// Masked, not lost: the id is what ties the report to the log line.
	if result.ErrorID == "" {
		t.Error("expect an error id to look the report up by")
	}
}

// TestSessionMasksPanicValue checks a page that panicked with something other
// than an error is masked too. That one is formatted with %v, so whatever it
// holds ends up in the message.
func TestSessionMasksPanicValue(t *testing.T) {
	const secret = "/srv/toolgui/secrets.yaml"

	session, recorder := newTestSession(t, func(p *Params) error {
		panic(secret)
	})
	defer session.Close()

	session.HandleEvent(&EventEmpty{})

	result := <-recorder.results
	if strings.Contains(result.Error, secret) {
		t.Errorf("Error = %q, carries the panic value", result.Error)
	}
}

// TestSessionShowsPageError checks the other side of it: an error the page
// function returned is the page talking to its user, so it arrives whole.
func TestSessionShowsPageError(t *testing.T) {
	const message = "this file needs a header row"

	session, recorder := newTestSession(t, func(p *Params) error {
		return errors.New(message)
	})
	defer session.Close()

	session.HandleEvent(&EventEmpty{})

	result := <-recorder.results
	if result.Success {
		t.Fatalf("expect a failure, got %#v", result)
	}

	if result.Error != message {
		t.Errorf("Error = %q, want %q", result.Error, message)
	}

	// Nothing was hidden, so there is no log line to point at.
	if result.ErrorID != "" {
		t.Errorf("ErrorID = %q, want none", result.ErrorID)
	}
}

// TestSessionShowsFailedComponentError checks a failure the page reported with
// Container.Fail reaches the user the same way. It is the page's own report,
// not the framework's.
func TestSessionShowsFailedComponentError(t *testing.T) {
	const message = "row 3 does not match the head"

	session, recorder := newTestSession(t, func(p *Params) error {
		p.Main.Fail(errors.New(message))
		return nil
	})
	defer session.Close()

	session.HandleEvent(&EventEmpty{})

	result := <-recorder.results
	if result.Error != message {
		t.Errorf("Error = %q, want %q", result.Error, message)
	}
}

// TestSessionMasksParseError checks an event the parser refused is reported as
// the framework's own error, not with the parser's message.
func TestSessionMasksParseError(t *testing.T) {
	session, recorder := newTestSession(t, func(p *Params) error { return nil })
	defer session.Close()

	if err := session.HandleRawEvent([]byte(`{"type":"unknown"}`)); err == nil {
		t.Fatal("expect a parse error")
	}

	result := <-recorder.results
	if result.Error != InternalErrorMessage {
		t.Errorf("Error = %q, want %q", result.Error, InternalErrorMessage)
	}

	if result.ErrorID == "" {
		t.Error("expect an error id to look the report up by")
	}
}
