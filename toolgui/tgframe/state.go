package tgframe

import (
	"bytes"
	"encoding/json"
	"io"
	"maps"
	"math"
	"runtime"
	"sync"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// State is the state of a user's session.
type State struct {
	values    map[string]any
	funcCache map[string]map[string]any

	// files is shared with the states cloned from this one, so no two of them
	// can hand out the same path. Its directory is made on the first upload,
	// so a state that never sees one leaves nothing behind.
	files *fileStore

	clickID string

	rwLock sync.RWMutex
}

// NewState creates a new state.
func NewState() *State {
	return &State{
		values:    make(map[string]any),
		files:     newFileStore(),
		funcCache: make(map[string]map[string]any),
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
		funcCache: maps.Clone(s.funcCache),
		clickID:   s.clickID,
	}
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

// Delete drops what key holds, value and uploaded file alike. It is how a
// widget's state is released when the widget leaves the page for good.
func (s *State) Delete(key string) {
	s.rwLock.Lock()
	delete(s.values, key)
	s.rwLock.Unlock()

	s.files.remove(key)
}

// GetObject gets the value of a key and unmarshals it to the out object.
func (s *State) GetObject(key string, out any) error {
	s.rwLock.RLock()
	val, ok := s.values[key]
	s.rwLock.RUnlock()

	if !ok {
		return nil
	}

	bs, err := json.Marshal(val)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	err = json.Unmarshal(bs, out)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	return nil
}

// toNumber reads any numeric type as a float64, so a default written from Go
// reads back like the float64 the frontend's JSON lands. A string is not one.
func toNumber(val any) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
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

// GetString reads a string, nil when the key holds none.
func (s *State) GetString(key string) *string {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	ss, ok := s.values[key].(string)
	if !ok {
		return nil
	}

	return &ss
}

// GetFloat reads any numeric type as a float64, nil when there is none.
func (s *State) GetFloat(key string) *float64 {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	f, ok := toNumber(s.values[key])
	if !ok {
		return nil
	}

	return &f
}

// GetInt is [State.GetFloat] truncated to an int, nil when an int cannot hold it.
func (s *State) GetInt(key string) *int {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	f, ok := toNumber(s.values[key])
	if !ok {
		return nil
	}

	// Bounds as float64: math.MaxInt has no exact one, so comparing against it
	// would let 2^63 through.
	if math.IsNaN(f) || f < float64(math.MinInt) || f >= -float64(math.MinInt) {
		return nil
	}

	i := int(f)
	return &i
}

// GetBool reads a bool, false when the key holds none.
func (s *State) GetBool(key string) bool {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	b, _ := s.values[key].(bool)
	return b
}

// WriteFile stores what r yields as the file under key, replacing whatever
// was there. The content is streamed to disk, so the upload never has to fit
// in memory.
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

// SetFuncCache sets the value of a key in the function cache.
func (s *State) SetFuncCache(key string, value any) {
	funcName := ""
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		funcName = runtime.FuncForPC(pc).Name()
	}
	s.SetFuncCacheWithFuncName(key, value, funcName)
}

// SetFuncCacheWithFuncName sets the value of a key in the function cache with a specific function name.
func (s *State) SetFuncCacheWithFuncName(key string, value any, funcName string) {
	if funcName == "" {
		pc, _, _, ok := runtime.Caller(1)
		if ok {
			funcName = runtime.FuncForPC(pc).Name()
		}
	}

	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	_, ok := s.funcCache[funcName]
	if !ok {
		s.funcCache[funcName] = make(map[string]any)
	}

	s.funcCache[funcName][key] = value
}

// GetFuncCache gets the value of a key in the function cache.
func (s *State) GetFuncCache(key string) any {
	funcName := ""
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		funcName = runtime.FuncForPC(pc).Name()
	}

	return s.GetFuncCacheWithFuncName(key, funcName)
}

// GetFuncCacheWithFuncName gets the value of a key in the function cache with a specific function name.
func (s *State) GetFuncCacheWithFuncName(key string, funcName string) any {
	if funcName == "" {
		pc, _, _, ok := runtime.Caller(1)
		if ok {
			funcName = runtime.FuncForPC(pc).Name()
		}
	}

	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	cache, ok := s.funcCache[funcName]
	if !ok {
		return nil
	}

	return cache[key]
}
