package tgframe

import (
	"sync"
	"sync/atomic"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// ErrNoSession is returned when a transport acts before starting a session.
var ErrNoSession = tgutil.NewError("no session, call start first")

// ErrSessionReplaced is returned to sends of a session a later start replaced.
var ErrSessionReplaced = tgutil.NewError("session replaced by a later start")

// ErrNotOnPage is returned for an upload to a key [State.HasFileKey] rejects.
var ErrNotOnPage = tgutil.NewError("component is not on the page")

// SessionHost holds the one session of a single-user transport (a desktop
// window, a browser tab): its state, and uploads in flight keyed by an id the
// transport picks. U is the transport's upload type.
//
// It is safe for concurrent use.
type SessionHost[U any] struct {
	app  *App
	send SendPackFunc

	mu       sync.Mutex
	session  *Session
	state    *State
	uploads  map[string]U
	detached *atomic.Bool
}

// NewSessionHost returns a host running pages of app, sending packs with send.
func NewSessionHost[U any](app *App, send SendPackFunc) *SessionHost[U] {
	return &SessionHost[U]{app: app, send: send, uploads: map[string]U{}}
}

// Start opens a session on page with rawQuery (`group=a`) and a new state,
// replacing the current session even when it fails. It returns the session
// and closeOld, which closes the replaced one. closeOld waits for the old page
// func, so call it where blocking is fine.
//
// The caller runs the first draw, e.g. HandleEvent(&EventEmpty{}).
func (h *SessionHost[U]) Start(page, rawQuery string) (session *Session, closeOld func(), err error) {
	detached := new(atomic.Bool)
	send := func(pack any) error {
		if detached.Load() {
			return ErrSessionReplaced
		}
		return h.send(pack)
	}

	state := NewState()

	query, err := ParseQuery(rawQuery)
	if err == nil {
		session, err = NewSession(h.app, page, query, state, send)
	}

	h.mu.Lock()
	closeOld = h.detachLocked()
	if err == nil {
		h.session, h.state, h.detached = session, state, detached
	}
	h.mu.Unlock()

	if err != nil {
		state.Destroy()
		// Unwrapped: transports show it to the user as is.
		return nil, closeOld, err
	}

	return session, closeOld, nil
}

// Close detaches the current session and returns what closes it, as
// [SessionHost.Start] does.
func (h *SessionHost[U]) Close() func() {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.detachLocked()
}

// Session returns the current session, nil before a start.
func (h *SessionHost[U]) Session() *Session {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.session
}

// Do runs fn on the current state and uploads under the host's lock, so a
// start can't swap them meanwhile. It returns [ErrNoSession] before a start.
func (h *SessionHost[U]) Do(fn func(state *State, uploads map[string]U) error) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.state == nil {
		return ErrNoSession
	}

	return fn(h.state, h.uploads)
}

// detachLocked takes the session off the host and returns its closer. Its
// sends stop at once; the state is destroyed after the session closes, since
// its uploads may hold files GC can't reclaim. Pending uploads are dropped.
// It must be called with mu held.
func (h *SessionHost[U]) detachLocked() func() {
	session, state := h.session, h.state
	h.session, h.state = nil, nil

	if h.detached != nil {
		h.detached.Store(true)
		h.detached = nil
	}

	clear(h.uploads)

	return func() {
		if session != nil {
			session.Close()
		}
		if state != nil {
			state.Destroy()
		}
	}
}
