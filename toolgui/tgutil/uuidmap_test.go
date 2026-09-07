package tgutil

import (
	"testing"
	"time"
)

// TestUUIDMapCleanupDestroys checks an entry dropped for being stale is
// destroyed like a deleted one. Anything it holds outside memory — a state's
// uploaded files, say — is unreachable once the entry is gone.
func TestUUIDMapCleanupDestroys(t *testing.T) {
	destroyed := make(chan int, 4)

	ttl := 10 * time.Millisecond
	m := NewUUIDMap(func() *int { v := 0; return &v },
		func(v *int) { destroyed <- *v }, ttl)

	id := m.New()
	value, _ := m.Get(id)
	*value = 7

	// Only an entry no connection is using can go.
	m.SetAlive(id, false)

	time.Sleep(21 * ttl)

	// cleanup runs when the map is next asked for an id.
	m.New()

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

// TestUUIDMapCleanupKeepsAlive checks an entry a connection is still using
// stays, however old it is.
func TestUUIDMapCleanupKeepsAlive(t *testing.T) {
	destroyed := make(chan int, 4)

	ttl := 10 * time.Millisecond
	m := NewUUIDMap(func() *int { v := 0; return &v },
		func(v *int) { destroyed <- *v }, ttl)

	id := m.New()

	time.Sleep(21 * ttl)
	m.New()

	if len(destroyed) != 0 {
		t.Error("expect a live entry to survive cleanup")
	}

	if _, alive := m.Get(id); !alive {
		t.Error("expect the entry to still be there")
	}
}
