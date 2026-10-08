package tgframe

import (
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
)

// queryPacks returns the QueryPacks the recorder got, and the index of each in
// all packs.
func queryPacks(r *packRecorder) ([]*QueryPack, []int) {
	r.lock.Lock()
	defer r.lock.Unlock()

	var packs []*QueryPack
	var idx []int
	for i, pack := range r.packs {
		if q, ok := pack.(*QueryPack); ok {
			packs = append(packs, q)
			idx = append(idx, i)
		}
	}

	return packs, idx
}

func TestReplaceQuerySendsPackBeforeResult(t *testing.T) {
	var seen atomic.Value
	session, recorder := newTestSession(t, func(p *Params) error {
		seen.Store(p.Query.Encode())
		p.ReplaceQuery(url.Values{"group": {"a"}})
		return nil
	})
	defer session.Close()

	session.HandleEvent(&EventEmpty{})
	<-recorder.results

	packs, idx := queryPacks(recorder)
	if len(packs) != 1 || packs[0].ReplaceQuery.Get("group") != "a" {
		t.Fatalf("query packs = %v", packs)
	}
	if idx[0] != recorder.count()-2 {
		t.Errorf("query pack at %d of %d, want just before the result",
			idx[0], recorder.count())
	}

	// The next run reads it, and the same query is not sent again.
	session.HandleEvent(&EventEmpty{})
	<-recorder.results

	if got := seen.Load(); got != "group=a" {
		t.Errorf("next run Query = %q", got)
	}
	if packs, _ := queryPacks(recorder); len(packs) != 1 {
		t.Errorf("unchanged query sent again: %d packs", len(packs))
	}
}

func TestReplaceQueryEmptyClears(t *testing.T) {
	app := NewApp()
	app.AddPage(testPageName, "Test", func(p *Params) error {
		p.ReplaceQuery(nil)
		return nil
	})

	recorder := newPackRecorder()
	session, err := NewSession(app, testPageName, url.Values{"x": {"1"}},
		NewState(), recorder.send)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	session.HandleEvent(&EventEmpty{})
	<-recorder.results

	packs, _ := queryPacks(recorder)
	if len(packs) != 1 || packs[0].ReplaceQuery == nil ||
		len(packs[0].ReplaceQuery) != 0 {
		t.Fatalf("query packs = %v", packs)
	}
}

func TestReplaceQueryTooLarge(t *testing.T) {
	session, recorder := newTestSession(t, func(p *Params) error {
		p.ReplaceQuery(url.Values{"x": {strings.Repeat("a", MaxQuerySize)}})
		return nil
	})
	defer session.Close()

	session.HandleEvent(&EventEmpty{})
	result := <-recorder.results

	if result.Success || !strings.Contains(result.Error, ErrQueryTooLarge.Error()) {
		t.Errorf("result = %+v, want a query too large error", result)
	}
	if packs, _ := queryPacks(recorder); len(packs) != 0 {
		t.Errorf("oversized query sent: %v", packs)
	}
}

// A cut run's ReplaceQuery is dropped; the run that finishes decides.
func TestReplaceQueryCutRunDropped(t *testing.T) {
	var runs atomic.Int32
	var seen atomic.Value
	started := make(chan struct{}, 1)

	session, recorder := newTestSession(t, func(p *Params) error {
		seen.Store(p.Query.Encode())

		switch runs.Add(1) {
		case 1:
			p.ReplaceQuery(url.Values{"x": {"cut"}})
			started <- struct{}{}
			<-p.Context.Done()
			return p.Context.Err()
		case 2:
			p.ReplaceQuery(url.Values{"x": {"done"}})
		}
		return nil
	})
	defer session.Close()

	session.HandleEvent(&EventEmpty{})
	<-started
	session.HandleEvent(&EventEmpty{})
	<-recorder.results

	packs, _ := queryPacks(recorder)
	if len(packs) != 1 || packs[0].ReplaceQuery.Get("x") != "done" {
		t.Fatalf("query packs = %v", packs)
	}

	session.HandleEvent(&EventEmpty{})
	<-recorder.results
	if got := seen.Load(); got != "x=done" {
		t.Errorf("next run Query = %q", got)
	}
}

// A run outside a session has nowhere to send it, and doesn't fail.
func TestReplaceQueryWithoutSession(t *testing.T) {
	app := NewApp()
	app.AddPage(testPageName, "Test", func(p *Params) error {
		p.ReplaceQuery(url.Values{"x": {"1"}})
		return nil
	})

	if err := app.Run(testPageName, NewState(), func(NotifyPack) {}); err != nil {
		t.Fatal(err)
	}

	(&Params{}).ReplaceQuery(url.Values{"x": {"1"}})
}
