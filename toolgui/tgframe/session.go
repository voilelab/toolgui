package tgframe

import (
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
	Error   string `json:"error,omitempty"`
	Success bool   `json:"success"`
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

	// stopUpdating tells the running page func to interrupt itself.
	stopUpdating atomic.Bool

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

	return &Session{
		app:      app,
		pageName: pageName,
		state:    state,
		send:     send,
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

	sendNotifyPack := func(pack NotifyPack) {
		if s.stopUpdating.Load() {
			panic(ErrUpdateInterrupt)
		}

		err := s.sendPack(pack)
		if err != nil {
			panic(err)
		}
	}

	s.stopUpdating.Store(false)
	go func() {
		defer s.endRun()

		err := s.app.RunWithHandlingPanic(s.pageName, s.state, sendNotifyPack)
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

// beginRun send the stop signal and wait for the running page func.
func (s *Session) beginRun() {
	s.stopUpdating.Store(true)
	s.running.Lock()
}

func (s *Session) endRun() {
	s.running.Unlock()
}
