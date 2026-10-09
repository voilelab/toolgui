package tgframe

import (
	"context"
	"log/slog"
	"net/url"
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

	// ErrorID names the log line carrying what really went wrong, for an
	// error the client is only told the kind of. It is empty for an error
	// whose message is already the whole of it.
	ErrorID string `json:"error_id,omitzero"`

	// Fatal marks an error the same request would run into again, such as a
	// page name the app doesn't have. A client is meant to give up on it
	// rather than reconnect.
	Fatal bool `json:"fatal,omitzero"`
}

// SendPackFunc sends a pack ([NotifyPack], [ReadyPack], [ResultPack],
// [QueryPack] or [NavigatePack]) to the GUI client. It's the only thing a
// [Session] needs from a transport.
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

	// query is the page query, replaced by [Params.ReplaceQuery]. It's only
	// touched with running held.
	query url.Values
	state *State
	send  SendPackFunc

	// handling serializes the stop-apply-run sequence, so an event can't
	// have its stop signal cleared by the event before it.
	handling sync.Mutex

	// sendLock serializes the calls to send, and guards sent.
	sendLock sync.Mutex

	// sent is what the client holds, so an unchanged component goes out as a
	// keep pack. Per session: a new page or a reconnect starts empty.
	sent *sentCache

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

	// rerunLock guards the fields below. It's only held for a few lines, so
	// [App.RerunAll] never waits on a run.
	rerunLock sync.Mutex

	// started is set by the first run: a session that never ran has nothing
	// on screen to refresh.
	started bool

	// runActive is set while a page func runs.
	runActive bool

	// pending is a server rerun not yet done. A run starting clears it, since
	// it reads the new data anyway.
	pending bool

	// kicking is set while a goroutine is on its way to start a server rerun,
	// so a burst of calls spawns one.
	kicking bool
}

// NewSession return a Session running page `pageName` of app with state.
// query is the page query every run reads as [Params.Query]; nil is an empty
// one. Return an error if the page does not exist, or wrapping
// [ErrQueryTooLarge] if query is over [MaxQuerySize] encoded. Either is fatal:
// the same request would fail the same way.
func NewSession(app *App, pageName string, query url.Values,
	state *State, send SendPackFunc) (*Session, error) {

	if !app.HasPage(pageName) {
		return nil, tgutil.Errorf("%w: `%s`", ErrPageNotFound, pageName)
	}

	if err := checkQuery(query); err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	s := &Session{
		app:      app,
		pageName: pageName,
		query:    cloneQuery(query),
		state:    state,
		send:     send,
		sent:     newSentCache(),

		ctx:    ctx,
		cancel: cancel,
	}
	app.addSession(s)

	return s, nil
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
		s.sendResult(ReportError("parse event", err))
		return tgutil.Errorf("%w", err)
	}

	s.HandleEvent(event)
	return nil
}

// HandleEvent apply the event to the state and rerun the page.
// A running page func is interrupted first, so the caller doesn't have to
// wait for it. It does nothing on a closed session.
//
// An event writing under an id the page is not showing is reported to the
// client and dropped: neither the state nor the run in flight is touched by
// it.
func (s *Session) HandleEvent(event Event) {
	s.handling.Lock()
	defer s.handling.Unlock()

	if s.closed.Load() {
		return
	}

	// The client fills in the ids an event writes, so they are checked against
	// what the page drew before anything is applied. A key belonging to no
	// component is one no run ever reads and no released slot ever deletes, so
	// taking it would let a client grow the state without bound. Checking
	// before beginRun also keeps a made-up id from cutting a healthy run.
	if err := validateEvent(event, s.state); err != nil {
		s.sendResult(&ResultPack{Error: err.Error()})
		slog.Warn("reject event", "error", err)
		return
	}

	s.startRun(event)
}

// startRun cuts the run in flight and runs the page with event. It must be
// called with handling held.
func (s *Session) startRun(event Event) {
	s.beginRun()

	s.rerunLock.Lock()
	s.started, s.runActive, s.pending = true, true, false
	s.rerunLock.Unlock()

	// tell client we cut the previous runner
	err := s.sendReady()
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

		err := s.sendNotify(pack)
		if err != nil {
			panic(err)
		}
	}

	go func() {
		defer cancelRun()
		defer s.endRun()
		defer s.finishRun()

		var replaced url.Values
		var navigate *Navigation
		err := s.app.runContextWithHandlingPanic(
			runCtx, s.pageName, s.query, s.state, sendNotifyPack,
			func(q url.Values, nav *Navigation) { replaced, navigate = q, nav })

		// Cancelled means the run was cut, so what it came back with is how
		// it unwound, not a failure, and the run replacing it is about to
		// paint the screen anyway. Asking the context and not the error is
		// what leaves an app's own context.Canceled a reportable error.
		if runCtx.Err() != nil {
			return
		}

		// Before the result, so the client has the new query by the time it
		// sees the run end.
		switch {
		case navigate != nil:
			s.sendNavigate(navigate)
		case replaced != nil:
			s.replaceQuery(replaced)
		}

		if err != nil {
			s.sendResult(ReportError("run err", err))
			return
		}

		s.sendSuccess()
	}()
}

// Close interrupt the running page func and wait for it.
// Events received after Close are ignored.
func (s *Session) Close() {
	s.handling.Lock()
	defer s.handling.Unlock()

	s.closed.Store(true)
	s.app.removeSession(s)
	s.cancel()
	s.beginRun()
	s.endRun()
}

// requestRerun reruns the page as the rerun button would, but never cuts a
// run in flight: that one is followed by a rerun once it ends. It doesn't
// block.
func (s *Session) requestRerun() {
	s.rerunLock.Lock()
	defer s.rerunLock.Unlock()

	if !s.started {
		return
	}

	s.pending = true
	s.kickLocked()
}

// kickLocked spawns a goroutine starting the pending rerun, unless a run is
// active or one is already on its way. It must be called with rerunLock held.
func (s *Session) kickLocked() {
	if !s.pending || s.runActive || s.kicking || s.closed.Load() {
		return
	}

	s.kicking = true
	go s.kick()
}

func (s *Session) kick() {
	s.handling.Lock()
	defer s.handling.Unlock()

	s.rerunLock.Lock()
	s.kicking = false
	// A user event may have started a run while this waited for handling.
	ready := s.pending && !s.runActive && !s.closed.Load()
	s.rerunLock.Unlock()

	if ready {
		s.startRun(&EventEmpty{})
	}
}

// finishRun marks the run ended and starts the rerun it held back, if any.
// It runs before endRun, so the next run can't start before it.
func (s *Session) finishRun() {
	s.rerunLock.Lock()
	defer s.rerunLock.Unlock()

	s.runActive = false
	s.kickLocked()
}

// replaceQuery stores q as the query of later runs and tells the client, if it
// differs from the current one. It runs with s.running held, which is what
// orders it before the next run reads s.query.
func (s *Session) replaceQuery(q url.Values) {
	if q.Encode() == s.query.Encode() {
		return
	}

	s.query = q

	err := s.sendPack(&QueryPack{ReplaceQuery: q})
	if err != nil {
		slog.Error("send query pack", "error", err)
	}
}

// sendNavigate tells the client to open another page. The client leaves for
// a new session, so this one keeps its page and query.
func (s *Session) sendNavigate(nav *Navigation) {
	err := s.sendPack(&NavigatePack{Navigate: nav})
	if err != nil {
		slog.Error("send navigate pack", "error", err)
	}
}

// sendPack send a pack to the client. Sends are serialized, so a transport
// never sees two of them at once.
func (s *Session) sendPack(pack any) error {
	s.sendLock.Lock()
	defer s.sendLock.Unlock()

	return s.send(pack)
}

// sendReady starts a run on the client and in the cache alike.
func (s *Session) sendReady() error {
	s.sendLock.Lock()
	defer s.sendLock.Unlock()

	s.sent.beginRun()
	return s.send(&ReadyPack{Ready: true})
}

// sendNotify sends pack, or the keep pack standing for it.
func (s *Session) sendNotify(pack NotifyPack) error {
	s.sendLock.Lock()
	defer s.sendLock.Unlock()

	out, err := s.sent.filter(pack)
	if err != nil {
		return err
	}

	return s.send(out)
}

// sendSuccess ends a run on the client and in the cache alike. Only a success
// drops what the run did not send; a failed or cut run leaves the tree alone.
func (s *Session) sendSuccess() {
	s.sendLock.Lock()
	defer s.sendLock.Unlock()

	s.sent.endRun()
	err := s.send(&ResultPack{Success: true})
	if err != nil {
		slog.Error("send result pack", "error", err)
	}
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
