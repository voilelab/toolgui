package tgutil

import (
	"errors"
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
