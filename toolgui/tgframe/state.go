package tgframe

import (
	"bytes"
	"io"
	"maps"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/voilelab/toolgui/toolgui/tgjson"
	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// State is the state of a user's session.
type State struct {
	values    map[string]any
	funcCache map[string]any

	// memos is what [State.Memo] keeps, one entry per slot.
	memos map[string]memoEntry

	// memoSeq counts computations per slot so an older one can't overwrite a
	// newer result.
	memoSeq map[string]uint64

	// resetKeys is the last reset key per component id, kept apart from
	// values to avoid key collisions.
	resetKeys map[string]string

	// files is shared with clones so no two hand out the same path. On a
	// server its directory is made on first upload; in the browser it is
	// opened up front.
	files *fileStore

	// downloads is shared with clones: tokens are fetched through the
	// transport's state, not the run's clone.
	downloads *downloadStore

	clickID string

	// runIDs are the component ids the last run drew; uploads are checked
	// against them.
	runIDs map[string]bool

	// indexedFileIDs is the subset of runIDs that accept [FileKey] uploads.
	indexedFileIDs map[string]bool

	// menuIDs are the click ids the app's menu declares, kept apart from
	// runIDs since menu items belong to no run.
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
		memos:     make(map[string]memoEntry),
		memoSeq:   make(map[string]uint64),
		resetKeys: make(map[string]string),
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
		values:         maps.Clone(s.values),
		files:          s.files,
		downloads:      s.downloads,
		funcCache:      maps.Clone(s.funcCache),
		memos:          maps.Clone(s.memos),
		memoSeq:        maps.Clone(s.memoSeq),
		resetKeys:      maps.Clone(s.resetKeys),
		runIDs:         maps.Clone(s.runIDs),
		indexedFileIDs: maps.Clone(s.indexedFileIDs),
		menuIDs:        s.menuIDs,
		clickID:        s.clickID,
	}
}

// setRunIDs records the component ids a run drew, and which of them accept
// [FileKey] uploads.
func (s *State) setRunIDs(ids, indexedFileIDs map[string]bool) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	s.runIDs = maps.Clone(ids)
	s.indexedFileIDs = maps.Clone(indexedFileIDs)
}

// setMenuIDs records the menu's click ids. The map is read-only after
// startup, so it is shared, not cloned.
func (s *State) setMenuIDs(ids map[string]bool) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	s.menuIDs = ids
}

// HasMenuID reports whether the app's menu declares the click id. Used by
// [MenuClicked].
func (s *State) HasMenuID(id string) bool {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	return s.menuIDs[id]
}

// HasComponentID reports whether the last run drew a component under id.
// Transports check client-named keys (e.g. upload targets) here first.
func (s *State) HasComponentID(id string) bool {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	return s.runIDs[id]
}

// MaxFileKeyIndex caps the index in a [FileKey], so a caller can't fill the
// disk with new indexes.
const MaxFileKeyIndex = 1000

// FileKey is the key of the i-th file of a multi-file upload under id.
func FileKey(id string, i int) string {
	return id + "/" + strconv.Itoa(i)
}

// splitFileKey is the inverse of [FileKey].
func splitFileKey(key string) (string, int, bool) {
	i := strings.LastIndexByte(key, '/')
	if i < 0 {
		return "", 0, false
	}

	n, err := strconv.Atoi(key[i+1:])
	if err != nil || n < 0 || n >= MaxFileKeyIndex || FileKey(key[:i], n) != key {
		return "", 0, false
	}

	return key[:i], n, true
}

// HasFileKey reports whether an upload may be stored under key: a drawn
// component id, or a [FileKey] of a drawn [IndexedFileComponent].
func (s *State) HasFileKey(key string) bool {
	if s.HasComponentID(key) {
		return true
	}

	id, _, ok := splitFileKey(key)
	if !ok {
		return false
	}

	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	return s.indexedFileIDs[id]
}

// removeIndexedFiles drops every [FileKey] file under id.
func (s *State) removeIndexedFiles(id string) {
	s.files.removeWhere(func(key string) bool {
		owner, _, ok := splitFileKey(key)
		return ok && owner == id
	})
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

// SwapResetKey records resetKey for id and returns the one recorded before,
// false when there was none.
func (s *State) SwapResetKey(id, resetKey string) (string, bool) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	last, ok := s.resetKeys[id]
	s.resetKeys[id] = resetKey
	return last, ok
}

// Delete drops everything under key: value, reset key, file and download.
// Used when a widget leaves the page for good.
func (s *State) Delete(key string) {
	s.rwLock.Lock()
	delete(s.values, key)
	delete(s.resetKeys, key)
	s.rwLock.Unlock()

	s.files.remove(key)
	s.removeIndexedFiles(key)
	s.downloads.remove(key)
}

// GetObject decodes what key holds into out via a JSON round trip.
//
// Unlike [State.Get] (a type assertion), it converts values the frontend sent
// (maps, []float64) into the Go type they stand for. Use Get for values the
// page wrote, GetObject for values the client sent.
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

// numberOf returns val as a number if its kind is numeric, so named types
// like `type Count int` count. Strings and uintptr don't.
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

// narrowInt64 converts i to an integral T, false when T can't hold it (e.g.
// int on 32-bit).
func narrowInt64[T Numeric](i int64) (T, bool) {
	t := T(i)
	if int64(t) != i {
		return 0, false
	}

	return t, true
}

// Get returns the value under key as a T, false when missing or of another
// type. It never panics.
func (s *State) Get[T any](key string) (T, bool) {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	v, ok := s.values[key].(T)
	return v, ok
}

// Default returns a pointer to the T under key, storing v first if the key
// holds no *T (another type is overwritten). Writes through it survive reruns.
func (s *State) Default[T any](key string, v T) *T {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	if p, ok := s.values[key].(*T); ok {
		return p
	}

	s.values[key] = &v
	return &v
}

// Numeric is the type [State.GetNumber] reads into; named types are allowed.
type Numeric interface {
	~int | ~int64 | ~float64
}

// GetNumber returns the number under key as a T, false when the key holds
// nothing numeric or a number T can't hold.
//
// Unlike [State.Get], it accepts any numeric type: the frontend sends JSON
// float64s while Go code may store ints, so Set(key, 30), Set(key, int64(30))
// and Set(key, 30.0) read the same. Strings are not numbers.
//
// Integers are read exactly (no float64 rounding past 2^53). A float read as
// an integral T truncates; out-of-range values return false.
func (s *State) GetNumber[T Numeric](key string) (T, bool) {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	v, ok := numberOf(s.values[key])
	if !ok {
		return 0, false
	}

	// Arithmetic, not a type switch, so named types work.
	integral := T(1)/T(2) == T(0)

	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		f := v.Float()
		if !integral {
			// A float T holds any float64, NaN and Inf included.
			return T(f), true
		}

		// Out-of-range float-to-int is platform dependent, so check the range
		// first, against exact float64 bounds (2^63, not MaxInt64).
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

// WriteFile streams r into the file under key, replacing what was there.
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

// NewFile makes an empty file under no key. Chunked transports fill it with
// [File.Append] and hand it to [State.PutFile] when done, so pages never see
// a partial file.
func (s *State) NewFile(name string) (*File, error) {
	file, err := s.files.newFile(name)
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	return file, nil
}

// PutFile stores file under key, dropping whatever the key held.
//
// Index 0 of a [FileKey] starts a new multi-file pick and drops the previous
// one.
func (s *State) PutFile(key string, file *File) {
	if id, i, ok := splitFileKey(key); ok && i == 0 {
		s.removeIndexedFiles(id)
	}

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

// SetDownload offers bs as a file to fetch. owner is the offering component's
// id, name the filename and mime the content type. The component sends
// [Download.Token] in its pack.
//
// Offering the same file again returns the same download and token.
// Different bytes replace it and invalidate the old token.
func (s *State) SetDownload(owner, name, mime string, bs []byte) (*Download, error) {
	download, err := s.downloads.set(s.files, owner, name, mime, bs)
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	return download, nil
}

// GetDownload returns the download for token, nil when this state offers
// none. Tokens from other states are never found.
func (s *State) GetDownload(token string) *Download {
	return s.downloads.get(token)
}

// SetFuncCache stores value in the function cache under key, to reuse
// results across runs.
//
// Keys are global to the state, so a key must encode the inputs (or a hash
// of them), or a later run reads a stale result.
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

// DeleteFuncCache removes the entry under key from the function cache.
func (s *State) DeleteFuncCache(key string) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	delete(s.funcCache, key)
}

type memoEntry struct {
	key   string
	value any
}

// Memo returns fn's result for key, calling fn only when slot holds no
// result for key. A slot keeps only its latest key.
//
// key should encode fn's inputs, as with [State.SetFuncCache]. Errors are
// returned and not cached.
func (s *State) Memo[T any](slot, key string, fn func() (T, error)) (T, error) {
	s.rwLock.RLock()
	e, ok := s.memos[slot]
	s.rwLock.RUnlock()

	if ok && e.key == key {
		if v, ok := e.value.(T); ok {
			return v, nil
		}
	}

	s.rwLock.Lock()
	s.memoSeq[slot]++
	seq := s.memoSeq[slot]
	s.rwLock.Unlock()

	// fn runs unlocked: it may be slow, and may use the state itself.
	v, err := fn()
	if err != nil {
		return v, err
	}

	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	// Only the latest computation for the slot is kept.
	if s.memoSeq[slot] == seq {
		s.memos[slot] = memoEntry{key: key, value: v}
	}
	return v, nil
}
