package tgutil

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// mustNew returns the id of a new entry, failing the test if the map refuses
// to make one.
func mustNew[T any](t *testing.T, m UUIDMap[T]) string {
	t.Helper()

	id, err := m.New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return id
}

// TestUUIDMapCleanupDestroys checks an entry dropped for being stale is
// destroyed like a deleted one. Anything it holds outside memory — a state's
// uploaded files, say — is unreachable once the entry is gone.
func TestUUIDMapCleanupDestroys(t *testing.T) {
	destroyed := make(chan int, 4)

	ttl := 10 * time.Millisecond
	m := NewUUIDMap(func() *int { v := 0; return &v },
		func(v *int) { destroyed <- *v }, ttl)

	id := mustNew(t, m)
	value, _ := m.Get(id)
	*value = 7

	// Only an entry no connection is using can go.
	m.SetAlive(id, false)

	time.Sleep(3 * ttl)

	// cleanup runs when the map is next asked for an id.
	mustNew(t, m)

	select {
	case got := <-destroyed:
		if got != 7 {
			t.Errorf("destroyed %d, want 7", got)
		}
	default:
		t.Fatal("the stale entry was dropped without being destroyed")
	}

	if _, ok := m.Get(id); ok {
		t.Error("expect the stale entry to be gone")
	}
}

// TestUUIDMapCleanupDropsExpired checks the map really stops holding an entry
// that outlived the ttl with nothing using it. Get reports whether an entry is
// alive, not whether it is there, so the count is what says it is gone.
func TestUUIDMapCleanupDropsExpired(t *testing.T) {
	ttl := 10 * time.Millisecond
	m := NewUUIDMap(func() *int { v := 0; return &v }, func(v *int) {}, ttl)

	id := mustNew(t, m)
	m.SetAlive(id, false)

	time.Sleep(3 * ttl)

	// A sweep runs at most once per ttl, and one is due by now.
	mustNew(t, m)

	if got := m.Size(); got != 1 {
		t.Errorf("Size = %d, want 1: the expired entry is still held", got)
	}
}

// TestUUIDMapCleanupKeepsAlive checks an entry a connection is still using
// stays, however old it is.
func TestUUIDMapCleanupKeepsAlive(t *testing.T) {
	destroyed := make(chan int, 4)

	ttl := 10 * time.Millisecond
	m := NewUUIDMap(func() *int { v := 0; return &v },
		func(v *int) { destroyed <- *v }, ttl)

	id := mustNew(t, m)

	time.Sleep(3 * ttl)
	mustNew(t, m)

	if len(destroyed) != 0 {
		t.Error("expect a live entry to survive cleanup")
	}

	if _, alive := m.Get(id); !alive {
		t.Error("expect the entry to still be there")
	}
}

// TestUUIDMapMaxSize checks a map at its limit turns a caller away instead of
// carrying on handing out entries.
func TestUUIDMapMaxSize(t *testing.T) {
	m := NewUUIDMap(func() *int { v := 0; return &v },
		func(v *int) {}, time.Minute)
	m.SetMaxSize(2)

	mustNew(t, m)
	mustNew(t, m)

	if _, err := m.New(); !errors.Is(err, ErrUUIDMapFull) {
		t.Errorf("New err = %v, want ErrUUIDMapFull", err)
	}

	if got := m.Size(); got != 2 {
		t.Errorf("Size = %d, want 2", got)
	}
}

// TestUUIDMapMaxSizeReclaims checks a full map sweeps before refusing, so a
// service whose entries have all gone stale comes back on its own rather than
// staying full.
func TestUUIDMapMaxSizeReclaims(t *testing.T) {
	ttl := 10 * time.Millisecond
	m := NewUUIDMap(func() *int { v := 0; return &v }, func(v *int) {}, ttl)
	m.SetMaxSize(1)

	id := mustNew(t, m)
	m.SetAlive(id, false)

	time.Sleep(2 * ttl)

	// The one entry is stale, so the map has room for this one.
	newID := mustNew(t, m)

	if newID == id {
		t.Fatal("expect a new id")
	}

	if got := m.Size(); got != 1 {
		t.Errorf("Size = %d, want 1", got)
	}
}

// TestUUIDMapAcquire checks Acquire takes an idle entry and says why it could
// not when it could not.
func TestUUIDMapAcquire(t *testing.T) {
	m := NewUUIDMap(func() *int { v := 0; return &v },
		func(v *int) {}, time.Minute)

	if _, err := m.Acquire("nothing"); !errors.Is(err, ErrUUIDNotFound) {
		t.Errorf("Acquire of an unknown id: %v, want ErrUUIDNotFound", err)
	}

	// New hands the entry out alive, so the connection that made it holds it.
	id := mustNew(t, m)
	if _, err := m.Acquire(id); !errors.Is(err, ErrUUIDAlive) {
		t.Errorf("Acquire of a held id: %v, want ErrUUIDAlive", err)
	}

	m.SetAlive(id, false)

	value, err := m.Acquire(id)
	if err != nil {
		t.Fatalf("Acquire of an idle id: %v", err)
	}

	if value == nil {
		t.Fatal("expect the entry, got nil")
	}

	// It is this caller's now, so the next one is turned away.
	if _, err := m.Acquire(id); !errors.Is(err, ErrUUIDAlive) {
		t.Errorf("Acquire after a take-over: %v, want ErrUUIDAlive", err)
	}
}

// TestUUIDMapAcquireRace checks two connections racing for the same idle entry
// do not both come away with it. Sharing one state is how two pages end up
// writing over each other's clicks and uploads.
func TestUUIDMapAcquireRace(t *testing.T) {
	const racers = 8

	m := NewUUIDMap(func() *int { v := 0; return &v },
		func(v *int) {}, time.Minute)

	for round := 0; round < 100; round++ {
		id := mustNew(t, m)
		m.SetAlive(id, false)

		start := make(chan struct{})
		var won atomic.Int32
		var wg sync.WaitGroup

		for i := 0; i < racers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				<-start

				value, err := m.Acquire(id)
				if err == nil {
					won.Add(1)
					// Touch it: -race has something to report if a second
					// winner is writing here too.
					*value++
					return
				}

				if !errors.Is(err, ErrUUIDAlive) {
					t.Errorf("the losing Acquire: %v, want ErrUUIDAlive", err)
				}
			}()
		}

		close(start)
		wg.Wait()

		if got := won.Load(); got != 1 {
			t.Fatalf("round %d: %d connections took the state, want 1",
				round, got)
		}

		m.Del(id)
	}
}
