package tgframe

import (
	"bytes"
	"io"
	"maps"
	"math"
	"reflect"
	"sync"

	"github.com/voilelab/toolgui/toolgui/tgjson"
	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// State is the state of a user's session.
type State struct {
	values    map[string]any
	funcCache map[string]any

	// files is shared with the states cloned from this one, so no two of them
	// can hand out the same path. On a server its directory waits for the
	// first upload, so a state that never sees one leaves nothing behind; in
	// the browser it is opened with the state, because what an upload needs
	// there has to be ready before one can arrive.
	files *fileStore

	// downloads is shared with the states cloned from this one, like files:
	// the token a run hands the client is fetched back through the state the
	// transport holds, which is not the clone the run drew on.
	downloads *downloadStore

	clickID string

	// runIDs is the set of component ids the last run of the page drew. An
	// upload names the component it belongs to, and this is what says the
	// name is one of the page's own rather than one the caller made up.
	runIDs map[string]bool

	// menuIDs is the set of click ids the app's menu declares. It is the
	// app's, not a run's: a menu item belongs to no run, so [State.runIDs]
	// never holds one, and the two sets stay apart.
	menuIDs map[string]bool

	rwLock sync.RWMutex
}

// NewState creates a new state.
func NewState() *State {
	return &State{
		values:    make(map[string]any),
		files:     newFileStore(),
		downloads: newDownloadStore(),
		funcCache: make(map[string]any),
	}
}

// Destroy release the resource.
func (s *State) Destroy() {
	s.files.destroy()
}

// Clone do a swallow copy on [State]. The copy shares the uploaded files, so
// only one of the two may be destroyed.
func (s *State) Clone() *State {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()
	return &State{
		values:    maps.Clone(s.values),
		files:     s.files,
		downloads: s.downloads,
		funcCache: maps.Clone(s.funcCache),
		runIDs:    maps.Clone(s.runIDs),
		menuIDs:   s.menuIDs,
		clickID:   s.clickID,
	}
}

// setRunIDs records the component ids a run drew.
func (s *State) setRunIDs(ids map[string]bool) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	s.runIDs = maps.Clone(ids)
}

// setMenuIDs records the click ids the app's menu declares. The map is the
// app's and is never written after startup, so it is shared rather than
// cloned.
func (s *State) setMenuIDs(ids map[string]bool) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	s.menuIDs = ids
}

// HasMenuID reports whether the app's menu declares the click id. It is what
// [MenuClicked] checks a click against, the way [State.HasComponentID] guards
// a component's.
func (s *State) HasMenuID(id string) bool {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	return s.menuIDs[id]
}

// HasComponentID reports whether the last run of the page drew a component
// under id. A transport that stores something the client names -- an upload
// under its component id -- checks the name here first, or a caller could
// write to any key it likes.
func (s *State) HasComponentID(id string) bool {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	return s.runIDs[id]
}

// SetClickID set the id of clicked button.
func (s *State) SetClickID(id string) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()
	s.clickID = id
}

// GetClickID get the id of clicked button.
func (s *State) GetClickID() string {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()
	return s.clickID
}

// Set sets the value of a key.
func (s *State) Set(key string, v any) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	s.values[key] = v
}

// Delete drops what key holds -- value, uploaded file and offered download
// alike. It is how a widget's state is released when the widget leaves the
// page for good.
func (s *State) Delete(key string) {
	s.rwLock.Lock()
	delete(s.values, key)
	s.rwLock.Unlock()

	s.files.remove(key)
	s.downloads.remove(key)
}

// GetObject reads what key holds through a JSON round trip, into out.
//
// It is kept alongside [State.Get] because the two answer different
// questions. Get is a type assertion: it reads a value back as the type it
// was stored as, and nothing else. GetObject re-decodes it, so a value that
// arrived from the frontend as a map or a []float64 reads back into the Go
// struct or []int it stands for. Reach for Get for a value the page itself
// wrote, and for GetObject for one the client sent.
//
// A missing key is not an error: out is left as it was.
func (s *State) GetObject(key string, out any) error {
	s.rwLock.RLock()
	val, ok := s.values[key]
	s.rwLock.RUnlock()

	if !ok {
		return nil
	}

	bs, err := tgjson.Marshal(val)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	err = tgjson.Unmarshal(bs, out)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	return nil
}

// numberOf reads what a key holds as a number, whatever numeric type it was
// stored as. It goes by kind rather than by concrete type, so a page's own
// domain type -- a `type Count int` -- is a number here, the way [Numeric]
// says one is. A string is not one, and neither is a uintptr.
func numberOf(val any) (reflect.Value, bool) {
	if val == nil {
		return reflect.Value{}, false
	}

	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Float32, reflect.Float64:
		return v, true
	}

	return reflect.Value{}, false
}

// narrowInt64 converts i to an integral T, false when T cannot hold it. T may
// be narrower than an int64 -- int on a 32-bit platform -- and a conversion
// between integer types is a defined truncation, so the round trip settles it.
func narrowInt64[T Numeric](i int64) (T, bool) {
	t := T(i)
	if int64(t) != i {
		return 0, false
	}

	return t, true
}

// Get returns the value under key as a T, false when it is missing or another
// type. Reading a key the user filled in by hand never panics.
func (s *State) Get[T any](key string) (T, bool) {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	v, ok := s.values[key].(T)
	return v, ok
}

// Default returns a pointer to the T under key, storing v there first when the
// key holds nothing of that type. The state keeps the pointer, so writes
// through it survive the rerun; a key holding another type is overwritten.
func (s *State) Default[T any](key string, v T) *T {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	if p, ok := s.values[key].(*T); ok {
		return p
	}

	s.values[key] = &v
	return &v
}

// Numeric is the value type [State.GetNumber] reads a number back as. The
// tildes let a user's own named type be one, so a page can keep its domain
// type all the way in.
type Numeric interface {
	~int | ~int64 | ~float64
}

// GetNumber returns the number under key as a T, false when the key holds
// nothing numeric, or a number T cannot hold.
//
// Numbers are the one place [State.Get] is too literal to be useful. The
// frontend sends every number as JSON, so an event lands a float64 whatever
// the component's own type is, while a default written from Go carries
// whichever integer type was at hand; this reads either, so Set(key, 30),
// Set(key, int64(30)) and Set(key, 30.0) are the same value. A string is
// still not a number.
//
// An integer is read exactly: a stored id past 2^53 comes back as it went
// in, rather than rounded through a float64 on the way out. A float read as
// an integral T truncates, as the number components do, and a number T cannot
// hold is absent rather than whatever the conversion happened to produce.
func (s *State) GetNumber[T Numeric](key string) (T, bool) {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	v, ok := numberOf(s.values[key])
	if !ok {
		return 0, false
	}

	// Written as arithmetic rather than a type switch because a named type's
	// dynamic type is itself, not the type it is defined from.
	integral := T(1)/T(2) == T(0)

	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		f := v.Float()
		if !integral {
			// A floating point T holds every number a float64 can, NaN and
			// the infinities included.
			return T(f), true
		}

		// Go leaves a float-to-integer conversion unspecified outside the
		// target's range, and the platforms disagree on what they do there:
		// amd64 wraps to MinInt64, wasm saturates at MaxInt64. So the range
		// is checked in float64 first, against bounds that are exact -- 2^63
		// has a float64, math.MaxInt64 does not.
		if math.IsNaN(f) || f < float64(math.MinInt64) || f >= -float64(math.MinInt64) {
			return 0, false
		}

		return narrowInt64[T](int64(f))

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64:
		u := v.Uint()
		if !integral {
			return T(u), true
		}

		if u > math.MaxInt64 {
			return 0, false
		}

		return narrowInt64[T](int64(u))

	default:
		i := v.Int()
		if !integral {
			return T(i), true
		}

		return narrowInt64[T](i)
	}
}

// WriteFile stores what r yields as the file under key, replacing whatever
// was there. It is streamed to wherever the build keeps files, so what r
// yields is never held in memory all at once.
func (s *State) WriteFile(key, name string, r io.Reader) (*File, error) {
	file, err := s.NewFile(name)
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	if err := file.write(r, false); err != nil {
		file.body.remove()
		return nil, tgutil.Errorf("%w", err)
	}

	s.PutFile(key, file)
	return file, nil
}

// NewFile makes an empty file of the state's own, under no key. A transport
// that receives an upload in pieces fills one of these with [File.Append] and
// hands it to [State.PutFile] when the last piece lands, so a page never
// reads a file that is still arriving, and two uploads racing for the same
// key can't be spliced together.
func (s *State) NewFile(name string) (*File, error) {
	file, err := s.files.newFile(name)
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	return file, nil
}

// PutFile stores file under key, dropping whatever the key held.
func (s *State) PutFile(key string, file *File) {
	s.files.put(key, file)
}

// SetFile stores bs as the file under key.
func (s *State) SetFile(key, name string, bs []byte) (*File, error) {
	return s.WriteFile(key, name, bytes.NewReader(bs))
}

// GetFile returns the file stored under key, nil when there is none.
func (s *State) GetFile(key string) *File {
	return s.files.get(key)
}

// SetDownload offers bs to the app user as a file to fetch. owner is the id of
// the component offering it, name the filename to offer it under and mime what
// to serve it as. The bytes go where the build keeps files, and what the
// component puts in its pack is [Download.Token].
//
// A rerun that offers the same file again gets the same download back, so the
// token the client holds keeps working and nothing is written twice. Different
// bytes replace it, and the token before them stops being fetchable.
func (s *State) SetDownload(owner, name, mime string, bs []byte) (*Download, error) {
	download, err := s.downloads.set(s.files, owner, name, mime, bs)
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	return download, nil
}

// GetDownload returns the download token names, nil when this state offers
// none under it. Only this state's own: a token made for another state is not
// found here, which is what keeps one page's output out of another's reach.
func (s *State) GetDownload(token string) *Download {
	return s.downloads.get(token)
}

// SetFuncCache stores value in the function cache under key, a place for what
// a run computed and the next run would rather not compute again.
//
// The key is the whole of the namespace: two calls naming the same key read
// and write the same entry, wherever in the page they are written. So a key
// has to say what the value was computed from -- the inputs, or a hash of
// them -- or a later run reads back a result for inputs it no longer has.
func (s *State) SetFuncCache[T any](key string, value T) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	s.funcCache[key] = value
}

// GetFuncCache returns the value under key in the function cache as a T,
// false when the key holds nothing or holds another type.
func (s *State) GetFuncCache[T any](key string) (T, bool) {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	v, ok := s.funcCache[key].(T)
	return v, ok
}
