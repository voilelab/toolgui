package tgframe

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// ErrUpdateInterrupt is raised at panic when current state is going to interrupt.
var ErrUpdateInterrupt = tgutil.NewError("update interrupt")

// ReadyPack tells the client the previous run is cut and a new one is starting.
type ReadyPack struct {
	Ready bool `json:"ready"`
}

// ResultPack reports the result of a run to the client.
type ResultPack struct {
	Error   string `json:"error,omitzero"`
	Success bool   `json:"success"`

	// Fatal marks an error the same request would run into again, such as a
	// page name the app doesn't have. A client is meant to give up on it
	// rather than reconnect.
	Fatal bool `json:"fatal,omitzero"`
}

// SendPackFunc sends a pack ([NotifyPack], [ReadyPack] or [ResultPack]) to the
// GUI client. It's the only thing a [Session] needs from a transport.
// A Session serializes its calls, so it doesn't have to be safe for
// concurrent use.
type SendPackFunc func(pack any) error

// Session binds a state to a page and turns user events into page runs.
// It's transport-agnostic: web, desktop or any other executor only has to
// feed it events and provide a [SendPackFunc].
//
// A Session is safe for concurrent use, so a transport may hand it events
// from more than one goroutine.
type Session struct {
	app      *App
	pageName string
	state    *State
	send     SendPackFunc

	// handling serializes the stop-apply-run sequence, so an event can't
	// have its stop signal cleared by the event before it.
	handling sync.Mutex

	// sendLock serializes the calls to send.
	sendLock sync.Mutex

	// ctx is cancelled by Close. Every run derives its context from it, so
	// closing stops whatever is running.
	ctx    context.Context
	cancel context.CancelFunc

	// cancelRun cuts the run in flight. It's replaced under handling by each
	// new run.
	cancelRun context.CancelFunc

	// running is held while a page func is running. It's locked before a run
	// is launched and unlocked by the runner goroutine.
	running sync.Mutex

	closed atomic.Bool
}

// NewSession return a Session running page `pageName` of app with state.
// Return an error if the page does not exist.
func NewSession(app *App, pageName string, state *State, send SendPackFunc) (*Session, error) {
	if !app.HasPage(pageName) {
		return nil, tgutil.Errorf("%w: `%s`", ErrPageNotFound, pageName)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Session{
		app:      app,
		pageName: pageName,
		state:    state,
		send:     send,

		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// HandleRawEvent parse a raw event and rerun the page with it.
// On a parse error it reports the error to the client and returns it.
// It does nothing on a closed session.
func (s *Session) HandleRawEvent(bs []byte) error {
	if s.closed.Load() {
		return nil
	}

	event, err := ParseEvent(bs)
	if err != nil {
		s.sendResult(&ResultPack{Error: err.Error()})
		return tgutil.Errorf("%w", err)
	}

	s.HandleEvent(event)
	return nil
}

// HandleEvent apply the event to the state and rerun the page.
// A running page func is interrupted first, so the caller doesn't have to
// wait for it. It does nothing on a closed session.
func (s *Session) HandleEvent(event Event) {
	s.handling.Lock()
	defer s.handling.Unlock()

	if s.closed.Load() {
		return
	}

	s.beginRun()

	// tell client we cut the previous runner
	err := s.sendPack(&ReadyPack{Ready: true})
	if err != nil {
		slog.Error("send ready pack", "error", err)
	}

	// Clear temp state
	s.state.SetClickID("")
	event.ApplyState(s.state)

	runCtx, cancelRun := context.WithCancel(s.ctx)
	s.cancelRun = cancelRun

	sendNotifyPack := func(pack NotifyPack) {
		// A page func that never looks at its context is still cut here, at
		// the next thing it draws.
		if runCtx.Err() != nil {
			panic(ErrUpdateInterrupt)
		}

		err := s.sendPack(pack)
		if err != nil {
			panic(err)
		}
	}

	go func() {
		defer cancelRun()
		defer s.endRun()

		err := s.app.RunContextWithHandlingPanic(
			runCtx, s.pageName, s.state, sendNotifyPack)

		// Cancelled means the run was cut, so what it came back with is how
		// it unwound, not a failure, and the run replacing it is about to
		// paint the screen anyway. Asking the context and not the error is
		// what leaves an app's own context.Canceled a reportable error.
		if runCtx.Err() != nil {
			return
		}

		if err != nil {
			s.sendResult(&ResultPack{Error: err.Error()})
			slog.Error("run err", "error", err)
			return
		}

		s.sendResult(&ResultPack{Success: true})
	}()
}

// Close interrupt the running page func and wait for it.
// Events received after Close are ignored.
func (s *Session) Close() {
	s.handling.Lock()
	defer s.handling.Unlock()

	s.closed.Store(true)
	s.cancel()
	s.beginRun()
	s.endRun()
}

// sendPack send a pack to the client. Sends are serialized, so a transport
// never sees two of them at once.
func (s *Session) sendPack(pack any) error {
	s.sendLock.Lock()
	defer s.sendLock.Unlock()

	return s.send(pack)
}

// sendResult send a result pack. A failed send is only logged: it means the
// client is gone and there is nowhere left to report it to.
func (s *Session) sendResult(pack *ResultPack) {
	err := s.sendPack(pack)
	if err != nil {
		slog.Error("send result pack", "error", err)
	}
}

// beginRun cut the running page func and wait for it.
func (s *Session) beginRun() {
	if s.cancelRun != nil {
		s.cancelRun()
	}

	s.running.Lock()
}

func (s *Session) endRun() {
	s.running.Unlock()
}
