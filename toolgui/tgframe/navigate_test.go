package tgframe

import (
	"errors"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
)

// navigatePacks returns the NavigatePacks the recorder got, and the index of
// each in all packs.
func navigatePacks(r *packRecorder) ([]*NavigatePack, []int) {
	r.lock.Lock()
	defer r.lock.Unlock()

	var packs []*NavigatePack
	var idx []int
	for i, pack := range r.packs {
		if n, ok := pack.(*NavigatePack); ok {
			packs = append(packs, n)
			idx = append(idx, i)
		}
	}

	return packs, idx
}

// newNavigateSession is newTestSession with a "detail" page to go to.
func newNavigateSession(t *testing.T, runFunc RunFunc) (*Session, *packRecorder) {
	t.Helper()

	app := NewApp()
	app.AddPage(testPageName, "Test", runFunc)
	app.AddPage("detail", "Detail", func(*Params) error { return nil })

	recorder := newPackRecorder()
	session, err := NewSession(app, testPageName, nil, NewState(), recorder.send)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	return session, recorder
}

func TestNavigateSendsPackBeforeResult(t *testing.T) {
	session, recorder := newNavigateSession(t, func(p *Params) error {
		p.Navigate("detail", url.Values{"id": {"0004"}})
		return nil
	})
	defer session.Close()

	session.HandleEvent(&EventEmpty{})
	if r := <-recorder.results; !r.Success {
		t.Fatalf("result = %+v", r)
	}

	packs, idx := navigatePacks(recorder)
	if len(packs) != 1 || packs[0].Navigate.Page != "detail" ||
		packs[0].Navigate.Query.Get("id") != "0004" {
		t.Fatalf("navigate packs = %v", packs)
	}
	if idx[0] != recorder.count()-2 {
		t.Errorf("navigate pack at %d of %d, want just before the result",
			idx[0], recorder.count())
	}
}

func TestNavigateNilQueryNotNil(t *testing.T) {
	session, recorder := newNavigateSession(t, func(p *Params) error {
		p.Navigate("detail", nil)
		return nil
	})
	defer session.Close()

	session.HandleEvent(&EventEmpty{})
	<-recorder.results

	packs, _ := navigatePacks(recorder)
	if len(packs) != 1 || packs[0].Navigate.Query == nil {
		t.Fatalf("navigate packs = %v", packs)
	}
}

// Navigate wins over ReplaceQuery, whichever comes first.
func TestNavigateWinsOverReplaceQuery(t *testing.T) {
	session, recorder := newNavigateSession(t, func(p *Params) error {
		p.Navigate("detail", nil)
		p.ReplaceQuery(url.Values{"x": {"1"}})
		return nil
	})
	defer session.Close()

	session.HandleEvent(&EventEmpty{})
	<-recorder.results

	if packs, _ := navigatePacks(recorder); len(packs) != 1 {
		t.Errorf("navigate packs = %v", packs)
	}
	if packs, _ := queryPacks(recorder); len(packs) != 0 {
		t.Errorf("query packs = %v", packs)
	}
}

func TestNavigateUnknownPage(t *testing.T) {
	session, recorder := newNavigateSession(t, func(p *Params) error {
		p.Navigate("detail", nil)
		p.Navigate("https://evil.example", nil)
		return nil
	})
	defer session.Close()

	session.HandleEvent(&EventEmpty{})
	r := <-recorder.results
	if r.Success || !strings.Contains(r.Error, ErrPageNotFound.Error()) {
		t.Errorf("result = %+v", r)
	}

	// The failed call is the last one: the earlier page is dropped too.
	if packs, _ := navigatePacks(recorder); len(packs) != 0 {
		t.Errorf("navigate packs = %v", packs)
	}
}

func TestNavigateTooLarge(t *testing.T) {
	session, recorder := newNavigateSession(t, func(p *Params) error {
		p.Navigate("detail", url.Values{"x": {strings.Repeat("a", MaxQuerySize)}})
		return nil
	})
	defer session.Close()

	session.HandleEvent(&EventEmpty{})
	if r := <-recorder.results; r.Success {
		t.Error("an oversized query passed")
	}
	if packs, _ := navigatePacks(recorder); len(packs) != 0 {
		t.Errorf("navigate packs = %v", packs)
	}
}

// A cut run's Navigate is dropped.
func TestNavigateCutRunDropped(t *testing.T) {
	var runs atomic.Int32
	started := make(chan struct{}, 1)

	session, recorder := newNavigateSession(t, func(p *Params) error {
		if runs.Add(1) == 1 {
			p.Navigate("detail", nil)
			started <- struct{}{}
			<-p.Context.Done()
			return p.Context.Err()
		}
		return nil
	})
	defer session.Close()

	session.HandleEvent(&EventEmpty{})
	<-started
	session.HandleEvent(&EventEmpty{})
	<-recorder.results

	if packs, _ := navigatePacks(recorder); len(packs) != 0 {
		t.Errorf("navigate packs = %v", packs)
	}
}

// A run outside a session checks the page, and has nowhere to send it.
func TestNavigateWithoutSession(t *testing.T) {
	app := NewApp()
	app.AddPage(testPageName, "Test", func(p *Params) error {
		p.Navigate("nope", nil)
		return nil
	})

	err := app.Run(testPageName, NewState(), func(NotifyPack) {})
	if !errors.Is(err, ErrPageNotFound) {
		t.Errorf("err = %v, want ErrPageNotFound", err)
	}

	(&Params{}).Navigate("x", nil)
}

// A failed Navigate leaves the address bar alone too.
func TestNavigateFailedDropsReplaceQuery(t *testing.T) {
	session, recorder := newNavigateSession(t, func(p *Params) error {
		p.ReplaceQuery(url.Values{"x": {"1"}})
		p.Navigate("nope", nil)
		return nil
	})
	defer session.Close()

	session.HandleEvent(&EventEmpty{})
	if r := <-recorder.results; r.Success {
		t.Error("an unknown page passed")
	}
	if packs, _ := queryPacks(recorder); len(packs) != 0 {
		t.Errorf("query packs = %v", packs)
	}
}
