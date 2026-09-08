package tgutil

import (
	"errors"
	"sync"
	"time"
	"uuid"
)

// ErrUUIDMapFull is the error [UUIDMap.New] returns when the map already
// holds as many items as it may.
var ErrUUIDMapFull = errors.New("uuidmap is full")

// UUIDMap provide a goroutine-safe mapping from UUID to T.
type UUIDMap[T any] interface {
	// New create a (id -> T) mapping and return id. It returns
	// [ErrUUIDMapFull] when the map is at its size limit and nothing in it
	// can be reclaimed.
	New() (string, error)

	// Get return (T, alive) by id, return nil if id does not exist.
	Get(id string) (*T, bool)

	// Del delete uuid.
	Del(id string)

	// SetAlive flag of id
	SetAlive(id string, alive bool)

	// SetMaxSize limit the number of T the map holds. A size of 0 or less is
	// no limit, which is the default.
	SetMaxSize(size int)

	// Size return the number of T.
	Size() int

	// Destroy the resource hold by the State.
	Destroy()
}

type dataPair[T any] struct {
	value     *T
	timestamp time.Time
	alive     bool
}

type uuidmap[T any] struct {
	lock        sync.RWMutex
	data        map[string]*dataPair[T]
	constructor func() *T
	destructor  func(*T)

	ttl           time.Duration
	latestCleanup time.Time

	// maxSize is the number of items the map takes, 0 for no limit.
	maxSize int
}

// NewUUIDMap create T by providing the constructor and destructor the of T.
// An item nothing has kept alive for ttl is reclaimed. The map takes any
// number of items until [UUIDMap.SetMaxSize] gives it a limit.
func NewUUIDMap[T any](
	constructor func() *T, destructor func(*T), ttl time.Duration) UUIDMap[T] {

	return &uuidmap[T]{
		data:          make(map[string]*dataPair[T]),
		constructor:   constructor,
		destructor:    destructor,
		ttl:           ttl,
		latestCleanup: time.Now(),
	}
}

// Destroy release the resources hold by data.
func (ss *uuidmap[T]) Destroy() {
	for _, d := range ss.data {
		ss.destructor(d.value)
	}
}

// SetMaxSize limit the number of item.
func (ss *uuidmap[T]) SetMaxSize(size int) {
	ss.lock.Lock()
	defer ss.lock.Unlock()

	ss.maxSize = size
}

// Size return the number of item.
func (ss *uuidmap[T]) Size() int {
	return len(ss.data)
}

// cleanup drop the items that outlived the ttl and that nothing is using. It
// runs at most once per ttl, unless force asks for a sweep now: a map at its
// limit has to know whether it is really full before turning a caller away.
//
// Called with the lock held.
func (ss *uuidmap[T]) cleanup(force bool) {
	if !force && time.Since(ss.latestCleanup) < ss.ttl {
		return
	}

	ss.latestCleanup = time.Now()

	ids := []string{}
	for id, d := range ss.data {
		if time.Since(d.timestamp) > ss.ttl && !d.alive {
			ids = append(ids, id)
		}
	}
	for _, id := range ids {
		// An expired entry is as gone as a deleted one, so it gets the same
		// destructor: whatever it holds outside memory is nobody's otherwise.
		ss.destructor(ss.data[id].value)
		delete(ss.data, id)
	}
}

// New create a (id -> T) mapping and return id.
func (ss *uuidmap[T]) New() (string, error) {
	ss.lock.Lock()
	defer ss.lock.Unlock()

	ss.cleanup(false)

	if ss.maxSize > 0 && len(ss.data) >= ss.maxSize {
		// A full map may only be full of items nobody came back for, so it
		// sweeps off-schedule rather than staying full until the next one.
		ss.cleanup(true)

		if len(ss.data) >= ss.maxSize {
			return "", Errorf("%w", ErrUUIDMapFull)
		}
	}

	id := uuid.New().String()
	ss.data[id] = &dataPair[T]{
		value:     ss.constructor(),
		timestamp: time.Now(),
		alive:     true,
	}
	return id, nil
}

// SetAlive flag of id
func (ss *uuidmap[T]) SetAlive(id string, alive bool) {
	ss.lock.RLock()
	defer ss.lock.RUnlock()

	d, ok := ss.data[id]
	if !ok {
		return
	}

	d.timestamp = time.Now()
	d.alive = alive
}

// Get return T by id, return nil if id does not exist.
func (ss *uuidmap[T]) Get(id string) (*T, bool) {
	ss.lock.RLock()
	defer ss.lock.RUnlock()

	d, ok := ss.data[id]
	if !ok {
		return nil, false
	}

	d.timestamp = time.Now()
	return d.value, d.alive
}

// Del delete uuid.
func (ss *uuidmap[T]) Del(id string) {
	ss.lock.Lock()
	defer ss.lock.Unlock()

	d, ok := ss.data[id]
	if !ok {
		return
	}

	ss.destructor(d.value)
	delete(ss.data, id)
}
