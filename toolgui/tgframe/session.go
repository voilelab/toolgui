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

	// ErrorID names the log line with the full error, when the client only
	// gets its kind.
	ErrorID string `json:"error_id,omitzero"`

	// Fatal marks an error a retry would hit again (e.g. unknown page); the
	// client should give up rather than reconnect.
	Fatal bool `json:"fatal,omitzero"`
}

// SendPackFunc sends a pack ([NotifyPack], [ReadyPack], [ResultPack],
// [QueryPack] or [NavigatePack]) to the client. It's all a [Session] needs
// from a transport. Calls are serialized.
type SendPackFunc func(pack any) error

// Session binds a state to a page and turns user events into page runs.
// Transports only feed it events and provide a [SendPackFunc].
//
// A Session is safe for concurrent use.
type Session struct {
	app      *App
	pageName string

	// query is the page query, replaced by [Params.ReplaceQuery]. Guarded by
	// running.
	query url.Values
	state *State
	send  SendPackFunc

	// handling serializes the stop-apply-run sequence.
	handling sync.Mutex

	// sendLock serializes the calls to send, and guards sent.
	sendLock sync.Mutex

	// sent is what the client holds, so unchanged components go out as keep
	// packs. Starts empty per session.
	sent *sentCache

	// ctx is cancelled by Close; run contexts derive from it.
	ctx    context.Context
	cancel context.CancelFunc

	// cancelRun cuts the run in flight. Replaced under handling.
	cancelRun context.CancelFunc

	// running is held while a page func runs; unlocked by the runner
	// goroutine.
	running sync.Mutex

	closed atomic.Bool

	// rerunLock guards the fields below. Held briefly, so [App.RerunAll]
	// never waits on a run.
	rerunLock sync.Mutex

	// started is set by the first run; before it there is nothing to rerun.
	started bool

	// runActive is set while a page func runs.
	runActive bool

	// pending is a requested server rerun; any run starting clears it.
	pending bool

	// kicking is set while a goroutine is about to start a server rerun, so
	// a burst spawns one.
	kicking bool
}

// NewSession return a Session running page `pageName` of app with state.
// query is read by every run as [Params.Query]; nil is empty. It fails if the
// page does not exist, or with [ErrQueryTooLarge] if query exceeds
// [MaxQuerySize]. Both are fatal.
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

// HandleEvent apply the event to the state and rerun the page, interrupting
// a running page func first. It does nothing on a closed session.
//
// An event naming an id the page isn't showing is reported and dropped.
func (s *Session) HandleEvent(event Event) {
	s.handling.Lock()
	defer s.handling.Unlock()

	if s.closed.Load() {
		return
	}

	// Check client ids before applying, so unknown keys can't grow the state
	// or cut a healthy run.
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
		// Cut a page func that ignores its context at its next draw.
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

		// A cut run's error is just how it unwound. Check the context, not
		// the error, so an app's own context.Canceled is still reported.
		if runCtx.Err() != nil {
			return
		}

		// Before the result, so the client has the query when the run ends.
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

// requestRerun reruns the page without cutting a run in flight (it reruns
// after). It doesn't block.
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

// replaceQuery stores q for later runs and tells the client, if it changed.
// Runs with s.running held.
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

// sendNavigate tells the client to open another page in a new session.
func (s *Session) sendNavigate(nav *Navigation) {
	err := s.sendPack(&NavigatePack{Navigate: nav})
	if err != nil {
		slog.Error("send navigate pack", "error", err)
	}
}

// sendPack send a pack to the client, serialized.
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

	err = s.send(out)
	if c, ok := pack.(*notifyPackCreate); ok && err != nil {
		// The client may never have got it: don't keep it next run.
		s.sent.remove(c.Key)
	}
	return err
}

// sendSuccess ends a run on the client and in the cache. Only success drops
// what the run didn't send.
func (s *Session) sendSuccess() {
	s.sendLock.Lock()
	defer s.sendLock.Unlock()

	s.sent.endRun()
	err := s.send(&ResultPack{Success: true})
	if err != nil {
		slog.Error("send result pack", "error", err)
	}
}

// sendResult send a result pack. A failed send (client gone) is only logged.
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
