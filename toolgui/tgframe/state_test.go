package tgframe

import (
	"errors"
	"math"
	"testing"
)

const testStateKey = "number_component_Age"

// TestGetNumberNumericSources covers the two routes a number reaches the state
// by: the frontend, which sends JSON and so lands a float64, and user code
// writing a default with Set, which writes whatever Go type it had at hand.
func TestGetNumberNumericSources(t *testing.T) {
	cases := []struct {
		name string
		val  any
	}{
		{"int", 30},
		{"int8", int8(30)},
		{"int16", int16(30)},
		{"int32", int32(30)},
		{"int64", int64(30)},
		{"uint", uint(30)},
		{"uint8", uint8(30)},
		{"uint16", uint16(30)},
		{"uint32", uint32(30)},
		{"uint64", uint64(30)},
		{"float32", float32(30)},
		{"float64", 30.0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			state := NewState()
			state.Set(testStateKey, c.val)

			got, ok := state.GetNumber[float64](testStateKey)
			if !ok {
				t.Fatalf("GetNumber[float64] = _, false, want 30")
			}
			if got != 30 {
				t.Errorf("GetNumber[float64] = %v, want 30", got)
			}

			gotInt, ok := state.GetNumber[int](testStateKey)
			if !ok {
				t.Fatalf("GetNumber[int] = _, false, want 30")
			}
			if gotInt != 30 {
				t.Errorf("GetNumber[int] = %v, want 30", gotInt)
			}
		})
	}
}

// TestGettersOnWrongType pins the point of the whole exercise: a value of the
// wrong type reads back as absent, and never panics the page it was read from.
func TestGettersOnWrongType(t *testing.T) {
	cases := []struct {
		name string
		val  any
	}{
		{"string", "abc"},
		{"numeric string", "30"},
		{"bool", true},
		{"nil", nil},
		{"struct", struct{}{}},
		{"slice", []int{1}},
		{"uintptr", uintptr(30)},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			state := NewState()
			state.Set(testStateKey, c.val)

			if got, ok := state.GetNumber[float64](testStateKey); ok {
				t.Errorf("GetNumber[float64] = %v, true, want false", got)
			}
			if got, ok := state.GetNumber[int](testStateKey); ok {
				t.Errorf("GetNumber[int] = %v, true, want false", got)
			}
			// Each getter is skipped for the one case that is its own type.
			if _, isBool := c.val.(bool); !isBool {
				if got, ok := state.Get[bool](testStateKey); ok {
					t.Errorf("Get[bool] = %v, true, want false", got)
				}
			}
			if _, isString := c.val.(string); !isString {
				if got, ok := state.Get[string](testStateKey); ok {
					t.Errorf("Get[string] = %v, true, want false", got)
				}
			}
		})
	}
}

func TestGettersOnMissingKey(t *testing.T) {
	state := NewState()

	if got, ok := state.Get[string]("missing"); ok {
		t.Errorf("Get[string] = %v, true, want false", got)
	}
	if got, ok := state.Get[bool]("missing"); ok {
		t.Errorf("Get[bool] = %v, true, want false", got)
	}
	if got, ok := state.GetNumber[float64]("missing"); ok {
		t.Errorf("GetNumber[float64] = %v, true, want false", got)
	}
	if got, ok := state.GetNumber[int]("missing"); ok {
		t.Errorf("GetNumber[int] = %v, true, want false", got)
	}
}

func TestGettersOnRightType(t *testing.T) {
	state := NewState()
	state.Set("s", "abc")
	state.Set("b", true)
	state.Set("f", 1.5)

	if got, ok := state.Get[string]("s"); !ok || got != "abc" {
		t.Errorf("Get[string] = %v, %v, want abc, true", got, ok)
	}
	if got, ok := state.Get[bool]("b"); !ok || !got {
		t.Errorf("Get[bool] = %v, %v, want true, true", got, ok)
	}
	if got, ok := state.GetNumber[float64]("f"); !ok || got != 1.5 {
		t.Errorf("GetNumber[float64] = %v, %v, want 1.5, true", got, ok)
	}
	// An int reads a fractional value truncated, the way the number
	// component does when its T is integral.
	if got, ok := state.GetNumber[int]("f"); !ok || got != 1 {
		t.Errorf("GetNumber[int] = %v, %v, want 1, true", got, ok)
	}
}

// myCount and myRatio are named types, to pin that GetNumber handles one on
// both sides. A named type's dynamic type is itself, not the type it is
// defined from, so a type switch on the stored value would miss it.
type myCount int

type myRatio float64

func TestGetNumberNamedType(t *testing.T) {
	state := NewState()
	state.Set(testStateKey, 30.0)

	if got, ok := state.GetNumber[myCount](testStateKey); !ok || got != 30 {
		t.Errorf("GetNumber[myCount] = %v, %v, want 30, true", got, ok)
	}

	// Still integral, so a number an int cannot hold is still absent.
	state.Set(testStateKey, math.NaN())
	if got, ok := state.GetNumber[myCount](testStateKey); ok {
		t.Errorf("GetNumber[myCount] = %v, true, want false", got)
	}

	// And a domain type the page stored is a number on the way out, or a
	// page could not keep one in the state at all.
	state.Set(testStateKey, myCount(30))
	if got, ok := state.GetNumber[myCount](testStateKey); !ok || got != 30 {
		t.Errorf("GetNumber[myCount] = %v, %v, want 30, true", got, ok)
	}
	if got, ok := state.GetNumber[float64](testStateKey); !ok || got != 30 {
		t.Errorf("GetNumber[float64] = %v, %v, want 30, true", got, ok)
	}

	state.Set(testStateKey, myRatio(1.5))
	if got, ok := state.GetNumber[float64](testStateKey); !ok || got != 1.5 {
		t.Errorf("GetNumber[float64] = %v, %v, want 1.5, true", got, ok)
	}
}

// TestGetNumberKeepsLargeIntegers pins that an integer is not routed through
// a float64 on the way out. Past 2^53 that rounds, which would quietly hand
// back an id or a counter that is not the one the page stored.
func TestGetNumberKeepsLargeIntegers(t *testing.T) {
	cases := []struct {
		name string
		val  any
		want int64
	}{
		{"past 2^53", int64(1<<53 + 1), 1<<53 + 1},
		{"MaxInt64", int64(math.MaxInt64), math.MaxInt64},
		{"MinInt64", int64(math.MinInt64), math.MinInt64},
		{"uint64 MaxInt64", uint64(math.MaxInt64), math.MaxInt64},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			state := NewState()
			state.Set(testStateKey, c.val)

			if got, ok := state.GetNumber[int64](testStateKey); !ok || got != c.want {
				t.Errorf("GetNumber[int64] = %v, %v, want %v, true", got, ok, c.want)
			}
		})
	}
}

// TestGetNumberUnrepresentable pins the numbers an int cannot hold. Go leaves
// the conversion unspecified for these, and the platforms disagree on what
// they do: amd64 wraps 2^63 to MinInt64, wasm saturates it at MaxInt64. So
// the getter reports them as absent rather than handing either back.
func TestGetNumberUnrepresentable(t *testing.T) {
	cases := []struct {
		name string
		val  any
	}{
		{"NaN", math.NaN()},
		{"+Inf", math.Inf(1)},
		{"-Inf", math.Inf(-1)},
		{"past MaxInt", -float64(math.MinInt)},
		{"past MinInt", float64(math.MinInt) * 2},
		{"uint64 past MaxInt", uint64(math.MaxInt64) + 1},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			state := NewState()
			state.Set(testStateKey, c.val)

			if got, ok := state.GetNumber[int](testStateKey); ok {
				t.Errorf("GetNumber[int] = %v, true, want false", got)
			}
		})
	}

	// int is what it is on the platform, so the 64-bit bound is pinned
	// through int64 as well: 2^63 is exactly a float64, and no int64.
	state := NewState()
	state.Set(testStateKey, -float64(math.MinInt64))
	if got, ok := state.GetNumber[int64](testStateKey); ok {
		t.Errorf("GetNumber[int64] = %v, true, want false", got)
	}

	// The edges themselves still read: only past them is out.
	state.Set(testStateKey, float64(math.MinInt))
	if got, ok := state.GetNumber[int](testStateKey); !ok || got != math.MinInt {
		t.Errorf("GetNumber[int] = %v, %v, want MinInt, true", got, ok)
	}

	state.Set(testStateKey, float64(math.MinInt64))
	if got, ok := state.GetNumber[int64](testStateKey); !ok || got != math.MinInt64 {
		t.Errorf("GetNumber[int64] = %v, %v, want MinInt64, true", got, ok)
	}

	// A float64 T holds every number the state can land, infinities included.
	state.Set(testStateKey, math.Inf(1))
	if got, ok := state.GetNumber[float64](testStateKey); !ok || !math.IsInf(got, 1) {
		t.Errorf("GetNumber[float64] = %v, %v, want +Inf, true", got, ok)
	}
}

func TestGet(t *testing.T) {
	state := NewState()
	state.Set("s", "abc")

	if got, ok := state.Get[string]("s"); !ok || got != "abc" {
		t.Errorf("Get[string] = %v, %v, want abc, true", got, ok)
	}
	// Get is a plain type assertion: unlike GetNumber it does not read one
	// numeric type as another.
	if got, ok := state.Get[float64]("s"); ok || got != 0 {
		t.Errorf("Get[float64] = %v, %v, want 0, false", got, ok)
	}
	if got, ok := state.Get[string]("missing"); ok || got != "" {
		t.Errorf("Get[string] = %q, %v, want \"\", false", got, ok)
	}
}

type testTODOList struct {
	Items []string
}

func TestDefault(t *testing.T) {
	state := NewState()

	list := state.Default("todoList", testTODOList{})
	if len(list.Items) != 0 {
		t.Fatalf("Items = %v, want empty", list.Items)
	}

	list.Items = append(list.Items, "buy milk")

	// The state keeps the pointer, so the next run reads the write back.
	again := state.Default("todoList", testTODOList{})
	if len(again.Items) != 1 || again.Items[0] != "buy milk" {
		t.Errorf("Items = %v, want [buy milk]", again.Items)
	}
}

func TestDefaultOnWrongType(t *testing.T) {
	state := NewState()
	state.Set("todoList", "not a list")

	list := state.Default("todoList", testTODOList{Items: []string{"a"}})
	if len(list.Items) != 1 || list.Items[0] != "a" {
		t.Errorf("Items = %v, want [a]", list.Items)
	}
}

// TestGetIsNotGetNumber pins the one difference between the two generic
// readers: Get asks for the type the value was stored as, GetNumber asks for
// a number whatever type it was stored as.
func TestGetIsNotGetNumber(t *testing.T) {
	state := NewState()
	state.Set(testStateKey, 30)

	if got, ok := state.Get[float64](testStateKey); ok {
		t.Errorf("Get[float64] = %v, true, want false", got)
	}
	if got, ok := state.GetNumber[float64](testStateKey); !ok || got != 30 {
		t.Errorf("GetNumber[float64] = %v, %v, want 30, true", got, ok)
	}
}

func TestFuncCache(t *testing.T) {
	state := NewState()

	if got, ok := state.GetFuncCache[[]string]("missing"); ok {
		t.Errorf("GetFuncCache = %v, true, want false", got)
	}

	state.SetFuncCache("files", []string{"a", "b"})

	got, ok := state.GetFuncCache[[]string]("files")
	if !ok || len(got) != 2 || got[0] != "a" {
		t.Errorf("GetFuncCache = %v, %v, want [a b], true", got, ok)
	}

	// The cache holds any, so a key read back as the wrong type reads as a
	// miss rather than panicking the page.
	if got, ok := state.GetFuncCache[string]("files"); ok {
		t.Errorf("GetFuncCache[string] = %v, true, want false", got)
	}
}

// setThroughHelper and readThroughHelper are two different functions on
// purpose: see TestFuncCacheSurvivesRefactor.
func setThroughHelper(s *State, key string, v []string) {
	s.SetFuncCache(key, v)
}

func readThroughHelper(s *State, key string) ([]string, bool) {
	return s.GetFuncCache[[]string](key)
}

// TestFuncCacheSurvivesRefactor is the regression this API exists for. The
// cache used to take its namespace from runtime.Caller, so moving a pair of
// calls into a helper silently stopped them hitting the same entry -- no
// error, just a cache that never warmed. The key is the whole of the
// namespace now, so where the calls are written does not matter.
func TestFuncCacheSurvivesRefactor(t *testing.T) {
	state := NewState()

	// Written here, read from another function.
	state.SetFuncCache("files", []string{"a"})
	if got, ok := readThroughHelper(state, "files"); !ok || len(got) != 1 {
		t.Errorf("read through helper = %v, %v, want [a], true", got, ok)
	}

	// Written from one function, read from a different one.
	setThroughHelper(state, "more", []string{"b", "c"})
	if got, ok := readThroughHelper(state, "more"); !ok || len(got) != 2 {
		t.Errorf("read through helper = %v, %v, want [b c], true", got, ok)
	}

	// And read back here, where neither call was written.
	if got, ok := state.GetFuncCache[[]string]("more"); !ok || len(got) != 2 {
		t.Errorf("GetFuncCache = %v, %v, want [b c], true", got, ok)
	}
}

// TestFuncCacheCloneIsIndependent pins that a clone gets its own cache: the
// map used to be nested, and maps.Clone only copies the outer one.
func TestFuncCacheCloneIsIndependent(t *testing.T) {
	state := NewState()
	state.SetFuncCache("files", []string{"a"})

	clone := state.Clone()
	if got, ok := clone.GetFuncCache[[]string]("files"); !ok || len(got) != 1 {
		t.Fatalf("clone GetFuncCache = %v, %v, want [a], true", got, ok)
	}

	clone.SetFuncCache("files", []string{"b", "c"})

	if got, ok := state.GetFuncCache[[]string]("files"); !ok || len(got) != 1 {
		t.Errorf("GetFuncCache = %v, %v, want [a], true", got, ok)
	}
}

func TestDeleteFuncCache(t *testing.T) {
	state := NewState()
	state.SetFuncCache("files", []string{"a"})
	state.DeleteFuncCache("files")

	if got, ok := state.GetFuncCache[[]string]("files"); ok {
		t.Errorf("GetFuncCache = %v, true, want false", got)
	}
}

func TestMemo(t *testing.T) {
	state := NewState()

	calls := 0
	memo := func(key string) string {
		t.Helper()
		v, err := state.Memo("slot", key, func() (string, error) {
			calls++
			return "v_" + key, nil
		})
		if err != nil {
			t.Fatalf("Memo(%q) error = %v", key, err)
		}
		return v
	}

	if got := memo("a"); got != "v_a" || calls != 1 {
		t.Fatalf("Memo(a) = %q, calls %d, want v_a, 1", got, calls)
	}
	if got := memo("a"); got != "v_a" || calls != 1 {
		t.Errorf("Memo(a) again = %q, calls %d, want v_a, 1", got, calls)
	}

	// A new key replaces the slot's entry rather than adding one.
	if got := memo("b"); got != "v_b" || calls != 2 {
		t.Errorf("Memo(b) = %q, calls %d, want v_b, 2", got, calls)
	}
	if len(state.memos) != 1 {
		t.Errorf("len(memos) = %d, want 1", len(state.memos))
	}
	if got := memo("a"); got != "v_a" || calls != 3 {
		t.Errorf("Memo(a) after b = %q, calls %d, want v_a, 3", got, calls)
	}
}

func TestMemoError(t *testing.T) {
	state := NewState()
	wantErr := errors.New("boom")

	if _, err := state.Memo("slot", "k", func() (int, error) {
		return 0, wantErr
	}); !errors.Is(err, wantErr) {
		t.Fatalf("Memo error = %v, want %v", err, wantErr)
	}

	// The failure was not kept, so the next call computes again.
	got, err := state.Memo("slot", "k", func() (int, error) { return 7, nil })
	if err != nil || got != 7 {
		t.Errorf("Memo = %v, %v, want 7, nil", got, err)
	}
}

// TestMemoTypeMismatch pins that a slot read as another type is a miss, not
// a panic.
func TestMemoTypeMismatch(t *testing.T) {
	state := NewState()
	if _, err := state.Memo("slot", "k", func() (int, error) { return 1, nil }); err != nil {
		t.Fatalf("Memo error = %v", err)
	}

	got, err := state.Memo("slot", "k", func() (string, error) { return "s", nil })
	if err != nil || got != "s" {
		t.Errorf("Memo = %q, %v, want s, nil", got, err)
	}
}

func TestHasFileKey(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	s.setRunIDs(
		map[string]bool{"up": true, "single": true},
		map[string]bool{"up": true})

	for key, want := range map[string]bool{
		"single":             true,
		FileKey("single", 0): false,
		"up":                 true,
		FileKey("up", 0):     true,
		FileKey("up", 999):   true,
		FileKey("up", 1000):  false,
		"up/-1":              false,
		"up/01":              false,
		"up/x":               false,
		FileKey("other", 0):  false,
		"other":              false,
	} {
		if got := s.HasFileKey(key); got != want {
			t.Errorf("HasFileKey(%q) = %v, want %v", key, got, want)
		}
	}
}

// TestIndexedFilesCleanup checks index 0 starts a new pick, and Delete drops
// a component's indexed files.
func TestIndexedFilesCleanup(t *testing.T) {
	s := NewState()
	defer s.Destroy()

	for i := range 3 {
		if _, err := s.SetFile(FileKey("up", i), "a", []byte("x")); err != nil {
			t.Fatalf("SetFile: %v", err)
		}
	}

	if _, err := s.SetFile(FileKey("other", 1), "a", []byte("x")); err != nil {
		t.Fatalf("SetFile: %v", err)
	}

	if _, err := s.SetFile(FileKey("up", 0), "b", []byte("y")); err != nil {
		t.Fatalf("SetFile: %v", err)
	}

	if s.GetFile(FileKey("up", 0)) == nil {
		t.Error("expect the new index 0")
	}

	for _, key := range []string{FileKey("up", 1), FileKey("up", 2)} {
		if s.GetFile(key) != nil {
			t.Errorf("expect %s of the previous pick dropped", key)
		}
	}

	s.Delete("up")
	if s.GetFile(FileKey("up", 0)) != nil {
		t.Error("expect Delete to drop indexed files")
	}

	if s.GetFile(FileKey("other", 1)) == nil {
		t.Error("expect another component's file kept")
	}
}
