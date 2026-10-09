package tgframe

import (
	"sync/atomic"
	"testing"
	"time"
)

// rerunPage is a page that counts its runs and, while block is set, waits on
// release before returning, ignoring its context.
type rerunPage struct {
	runs    atomic.Int32
	block   atomic.Bool
	started chan struct{}
	release chan struct{}
}

func newRerunPage() *rerunPage {
	return &rerunPage{
		started: make(chan struct{}, 16),
		release: make(chan struct{}),
	}
}

func (r *rerunPage) run(p *Params) error {
	r.runs.Add(1)
	r.started <- struct{}{}
	if r.block.Load() {
		<-r.release
	}
	return nil
}

func openSession(t *testing.T, app *App, page string) (*Session, *packRecorder) {
	t.Helper()

	recorder := newPackRecorder()
	session, err := NewSession(app, page, nil, NewState(), recorder.send)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	t.Cleanup(session.Close)

	return session, recorder
}

func waitResult(t *testing.T, recorder *packRecorder) {
	t.Helper()

	select {
	case <-recorder.results:
	case <-time.After(interruptWait):
		t.Fatal("run did not finish")
	}
}

func expectNoResult(t *testing.T, recorder *packRecorder) {
	t.Helper()

	select {
	case <-recorder.results:
		t.Fatal("unexpected run")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestRerunAllRerunsOpenSessions(t *testing.T) {
	a, b := newRerunPage(), newRerunPage()
	app := NewApp()
	app.AddPage("a", "A", a.run)
	app.AddPage("b", "B", b.run)

	sa, ra := openSession(t, app, "a")
	sb, rb := openSession(t, app, "b")
	sa.HandleEvent(&EventEmpty{})
	sb.HandleEvent(&EventEmpty{})
	waitResult(t, ra)
	waitResult(t, rb)

	app.RerunAll()
	waitResult(t, ra)
	waitResult(t, rb)

	if a.runs.Load() != 2 || b.runs.Load() != 2 {
		t.Fatalf("runs = %d, %d, want 2, 2", a.runs.Load(), b.runs.Load())
	}
}

func TestRerunPageOnlyNamed(t *testing.T) {
	a, b := newRerunPage(), newRerunPage()
	app := NewApp()
	app.AddPage("a", "A", a.run)
	app.AddPage("b", "B", b.run)

	sa, ra := openSession(t, app, "a")
	sb, rb := openSession(t, app, "b")
	sa.HandleEvent(&EventEmpty{})
	sb.HandleEvent(&EventEmpty{})
	waitResult(t, ra)
	waitResult(t, rb)

	app.RerunPage("a")
	waitResult(t, ra)
	expectNoResult(t, rb)

	if a.runs.Load() != 2 || b.runs.Load() != 1 {
		t.Fatalf("runs = %d, %d, want 2, 1", a.runs.Load(), b.runs.Load())
	}
}

// TestRerunAllSkipsUnstartedSession checks a session with nothing drawn yet
// is left for its own first run.
func TestRerunAllSkipsUnstartedSession(t *testing.T) {
	page := newRerunPage()
	app := NewApp()
	app.AddPage("a", "A", page.run)

	_, recorder := openSession(t, app, "a")

	app.RerunAll()
	expectNoResult(t, recorder)
}

// TestRerunAllWaitsForRunningPage checks a run in flight is not cut, and is
// followed by one rerun however many calls came in.
func TestRerunAllWaitsForRunningPage(t *testing.T) {
	page := newRerunPage()
	app := NewApp()

	var cut atomic.Bool
	app.AddPage("a", "A", func(p *Params) error {
		err := page.run(p)
		if p.Context.Err() != nil {
			cut.Store(true)
		}
		return err
	})

	session, recorder := openSession(t, app, "a")

	session.HandleEvent(&EventEmpty{})
	waitResult(t, recorder)

	page.block.Store(true)
	session.HandleEvent(&EventEmpty{})
	<-page.started
	<-page.started

	for range 5 {
		app.RerunAll()
	}

	page.block.Store(false)
	close(page.release)

	waitResult(t, recorder)
	waitResult(t, recorder)
	expectNoResult(t, recorder)

	if cut.Load() {
		t.Error("RerunAll cut the run in flight")
	}
	if got := page.runs.Load(); got != 3 {
		t.Fatalf("runs = %d, want 3", got)
	}
}

// TestRerunAllDroppedByNewRun checks a user event starting a run clears the
// pending rerun: that run reads the new data already.
func TestRerunAllDroppedByNewRun(t *testing.T) {
	page := newRerunPage()
	app := NewApp()
	app.AddPage("a", "A", func(p *Params) error {
		page.runs.Add(1)
		page.started <- struct{}{}
		if page.block.Load() {
			select {
			case <-page.release:
			case <-p.Context.Done():
			}
		}
		return nil
	})

	session, recorder := openSession(t, app, "a")
	session.HandleEvent(&EventEmpty{})
	waitResult(t, recorder)

	page.block.Store(true)
	session.HandleEvent(&EventEmpty{})
	<-page.started
	<-page.started

	app.RerunAll()

	// Cuts the blocked run; this one is the rerun.
	page.block.Store(false)
	session.HandleEvent(&EventEmpty{})
	waitResult(t, recorder)
	expectNoResult(t, recorder)

	if got := page.runs.Load(); got != 3 {
		t.Fatalf("runs = %d, want 3", got)
	}
}

func TestClosedSessionLeavesRegistry(t *testing.T) {
	page := newRerunPage()
	app := NewApp()
	app.AddPage("a", "A", page.run)

	session, recorder := openSession(t, app, "a")
	session.HandleEvent(&EventEmpty{})
	waitResult(t, recorder)

	session.Close()

	app.sessionsLock.Lock()
	n := len(app.sessions)
	app.sessionsLock.Unlock()
	if n != 0 {
		t.Fatalf("registry holds %d sessions, want 0", n)
	}

	app.RerunAll()
	expectNoResult(t, recorder)
}

// TestRerunAllDoesNotBlock checks a session stuck waiting on a page func
// doesn't hold RerunAll up.
func TestRerunAllDoesNotBlock(t *testing.T) {
	page := newRerunPage()
	app := NewApp()
	app.AddPage("a", "A", page.run)

	session, recorder := openSession(t, app, "a")
	session.HandleEvent(&EventEmpty{})
	waitResult(t, recorder)

	page.block.Store(true)
	session.HandleEvent(&EventEmpty{})
	<-page.started
	<-page.started

	// Holds handling while it waits for the page func that ignores its
	// context.
	go session.HandleEvent(&EventEmpty{})
	time.Sleep(20 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		app.RerunAll()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(interruptWait):
		t.Fatal("RerunAll blocked on a busy session")
	}

	page.block.Store(false)
	close(page.release)
}
