package tgframe

import (
	"bytes"
	"encoding/json"
	"io"
	"maps"
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

// Default sets the value of a key if the key is not set.
// If the key is set, it returns the value.
// If the key is not set, it sets the value and returns the value.
// The v should be a pointer.
// Example:
// ```go
//
//	type TODOList struct {
//		Items []string `json:"items"`
//	}
//
//	todoList := state.Default("todoList", &TODOList{}).(*TODOList)
//
// ```
func (s *State) Default(key string, v any) any {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	_, ok := s.values[key]
	if !ok {
		s.values[key] = v
	}

	return s.values[key]
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

// GetString gets the value of a key and returns it as a string.
func (s *State) GetString(key string) *string {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	val, ok := s.values[key]
	if !ok || val == nil {
		return nil
	}

	ss := val.(string)
	return &ss
}

// GetFloat gets the value of a key and returns it as a float64.
func (s *State) GetFloat(key string) *float64 {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	val, ok := s.values[key]
	if !ok || val == nil {
		return nil
	}

	f := val.(float64)
	return &f
}

// GetInt gets the value of a key and returns it as an int.
// If the key is not set, it returns nil.
func (s *State) GetInt(key string) *int {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	val, ok := s.values[key]
	if !ok || val == nil {
		return nil
	}

	i := val.(int)
	return &i
}

// GetBool gets the value of a key and returns it as a bool.
func (s *State) GetBool(key string) bool {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	val, ok := s.values[key]
	if !ok {
		return false
	}

	return val.(bool)
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
