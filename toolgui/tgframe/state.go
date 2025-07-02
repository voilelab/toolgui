package tgframe

import (
	"encoding/json"
	"maps"
	"runtime"
	"sync"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// State is the state of a user's session.
type State struct {
	values    map[string]any
	files     map[string][]byte
	funcCache map[string]map[string]any

	clickID string

	rwLock sync.RWMutex
}

// NewState creates a new state.
func NewState() *State {
	return &State{
		values:    make(map[string]any),
		files:     make(map[string][]byte),
		funcCache: make(map[string]map[string]any),
	}
}

// Destroy release the resource.
func (s *State) Destroy() {
}

// Clone do a swallow copy on [State].
func (s *State) Clone() *State {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()
	return &State{
		values:    maps.Clone(s.values),
		files:     maps.Clone(s.files),
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

// SetFile sets the value of a key to a file.
func (s *State) SetFile(key string, bs []byte) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	s.files[key] = bs
}

func (s *State) GetFile(key string) []byte {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	return s.files[key]
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
