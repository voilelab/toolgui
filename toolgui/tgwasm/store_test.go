//go:build js && wasm

package tgwasm

import (
	"crypto/rand"
	"errors"
	"io/fs"
	"slices"
	"strings"
	"syscall/js"
	"testing"
	"time"

	"github.com/voilelab/toolgui/toolgui/internal/opfs"
)

// These need a browser: `task test_wasm` runs them in a dedicated worker.

// testStore open a fresh store as if Run had installed the bridge, and
// removes it at the end of the test.
func testStore(t *testing.T) (*Store, string) {
	t.Helper()

	bridgeRunning.Store(true)

	name := "test-" + strings.ToLower(rand.Text())
	s := OpenStore(name)

	t.Cleanup(func() {
		s.close()
		removeStoreDir(t, name)
	})

	return s, name
}

// reopen close s and open its name again, as a reload would.
func reopen(t *testing.T, s *Store, name string) *Store {
	t.Helper()

	s.close()

	s2 := OpenStore(name)
	t.Cleanup(s2.close)

	return s2
}

func storeDir(t *testing.T, name string) js.Value {
	t.Helper()

	origin, err := opfs.Origin()
	if err != nil {
		t.Fatal(err)
	}

	root, err := opfs.AwaitCall(origin, "getDirectoryHandle", storeRootName)
	if err != nil {
		t.Fatal(err)
	}

	dir, err := opfs.AwaitCall(root, "getDirectoryHandle", name)
	if err != nil {
		t.Fatal(err)
	}

	return dir
}

func removeStoreDir(t *testing.T, name string) {
	origin, err := opfs.Origin()
	if err != nil {
		return
	}

	root, err := opfs.AwaitCall(origin, "getDirectoryHandle", storeRootName)
	if err != nil {
		return
	}

	_, _ = opfs.AwaitCall(root, "removeEntry", name, opfs.Recursive)
}

// editFile run fn on the bytes of one of a closed store's files and writes
// the result back.
func editFile(t *testing.T, name, file string, fn func([]byte) []byte) {
	t.Helper()

	f, err := openSyncFile(storeDir(t, name), file)
	if err != nil {
		t.Fatal(err)
	}
	defer f.close()

	bs, err := f.readAll()
	if err != nil {
		t.Fatal(err)
	}

	bs = fn(bs)

	if err := f.truncate(0); err != nil {
		t.Fatal(err)
	}

	if _, err := f.write(bs, 0); err != nil {
		t.Fatal(err)
	}

	if err := f.flush(); err != nil {
		t.Fatal(err)
	}
}

func mustGet(t *testing.T, s *Store, key, want string) {
	t.Helper()

	got, err := s.Get(key)
	if err != nil {
		t.Fatalf("Get(%q): %v", key, err)
	}

	if string(got) != want {
		t.Fatalf("Get(%q) = %q, want %q", key, got, want)
	}
}

func mustSet(t *testing.T, s *Store, key, value string) {
	t.Helper()

	if err := s.Set(key, []byte(value)); err != nil {
		t.Fatalf("Set(%q): %v", key, err)
	}
}

func TestStoreSurvivesReopen(t *testing.T) {
	s, name := testStore(t)

	mustSet(t, s, "draft/1", "one")
	mustSet(t, s, "draft/2", "two")
	mustSet(t, s, "draft/1", "uno")
	mustSet(t, s, "sub/1", "x")

	if err := s.Delete("draft/2"); err != nil {
		t.Fatal(err)
	}

	if err := s.Delete("missing"); err != nil {
		t.Fatalf("Delete of a missing key: %v", err)
	}

	s = reopen(t, s, name)

	mustGet(t, s, "draft/1", "uno")
	mustGet(t, s, "sub/1", "x")

	if _, err := s.Get("draft/2"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Get of a deleted key: %v, want fs.ErrNotExist", err)
	}

	keys, err := s.Keys("draft/")
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(keys, []string{"draft/1"}) {
		t.Fatalf("Keys = %q", keys)
	}

	ro, err := s.ReadOnly()
	if err != nil || ro {
		t.Fatalf("ReadOnly = %v, %v; want false", ro, err)
	}
}

func TestStoreJSON(t *testing.T) {
	type draft struct {
		Lang string
		Code string
		Tags []string
	}

	s, name := testStore(t)

	want := draft{Lang: "go", Code: "package main", Tags: []string{"a", "b"}}
	if err := SetJSON(s, "draft/1", want); err != nil {
		t.Fatal(err)
	}

	s = reopen(t, s, name)

	got, err := GetJSON[draft](s, "draft/1")
	if err != nil {
		t.Fatal(err)
	}

	if got.Lang != want.Lang || got.Code != want.Code ||
		!slices.Equal(got.Tags, want.Tags) {
		t.Fatalf("GetJSON = %+v, want %+v", got, want)
	}

	if _, err := GetJSON[draft](s, "missing"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("GetJSON of a missing key: %v, want fs.ErrNotExist", err)
	}

	mustSet(t, s, "bad", "{")
	if _, err := GetJSON[draft](s, "bad"); err == nil {
		t.Fatal("GetJSON of invalid JSON: nil error")
	}
}

func TestStoreCopiesValues(t *testing.T) {
	s, _ := testStore(t)

	v := []byte("abc")
	if err := s.Set("k", v); err != nil {
		t.Fatal(err)
	}

	v[0] = 'x'
	mustGet(t, s, "k", "abc")

	got, _ := s.Get("k")
	got[0] = 'y'
	mustGet(t, s, "k", "abc")
}

func TestStoreKeysSorted(t *testing.T) {
	s, _ := testStore(t)

	for _, k := range []string{"b", "a/2", "c", "a/1"} {
		mustSet(t, s, k, k)
	}

	keys, err := s.Keys("")
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(keys, []string{"a/1", "a/2", "b", "c"}) {
		t.Fatalf("Keys = %q", keys)
	}
}

func TestStoreDropsTornTail(t *testing.T) {
	s, name := testStore(t)

	mustSet(t, s, "a", "1")
	mustSet(t, s, "b", "2")
	s.close()

	// Half of a record, as a closed tab would leave.
	editFile(t, name, "a", func(bs []byte) []byte {
		rec := encodeSet("c", []byte("3"))
		return append(bs, rec[:len(rec)/2]...)
	})

	s = reopen(t, s, name)
	mustGet(t, s, "a", "1")
	mustGet(t, s, "b", "2")

	if _, err := s.Get("c"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Get of the torn record: %v", err)
	}

	// The tail is gone from the file, so this append is not lost behind it.
	mustSet(t, s, "d", "4")

	s = reopen(t, s, name)
	mustGet(t, s, "d", "4")
}

func TestStoreFailsOnCorruption(t *testing.T) {
	s, name := testStore(t)

	mustSet(t, s, "a", "1")
	mustSet(t, s, "b", "2")
	mustSet(t, s, "c", "3")
	s.close()

	// Flip a byte inside the record for b.
	editFile(t, name, "a", func(bs []byte) []byte {
		i := len(bs) - len(encodeSet("c", []byte("3"))) - 1
		bs[i] ^= 0xff
		return bs
	})

	s = reopen(t, s, name)
	if _, err := s.Get("a"); err == nil || !strings.Contains(err.Error(), "corrupt") {
		t.Fatalf("Get after mid-log corruption: %v, want corrupt", err)
	}
}

func TestStoreSecondOpenIsReadOnly(t *testing.T) {
	s, name := testStore(t)

	mustSet(t, s, "k", "v")

	s2 := OpenStore(name)
	t.Cleanup(s2.close)

	ro, err := s2.ReadOnly()
	if err != nil || !ro {
		t.Fatalf("ReadOnly = %v, %v; want true", ro, err)
	}

	mustGet(t, s2, "k", "v")

	if err := s2.Set("k", []byte("w")); !errors.Is(err, ErrReadOnly) {
		t.Fatalf("Set: %v, want ErrReadOnly", err)
	}

	if err := s2.Delete("k"); !errors.Is(err, ErrReadOnly) {
		t.Fatalf("Delete: %v, want ErrReadOnly", err)
	}

	// A snapshot: the owner's later writes do not show.
	mustSet(t, s, "k", "w")
	mustGet(t, s2, "k", "v")
}

// failingLog fail writes while fail is set; short writes half the record.
type failingLog struct {
	logFile
	fail  bool
	short bool
}

func (f *failingLog) write(bs []byte, at int64) (int, error) {
	if !f.fail {
		return f.logFile.write(bs, at)
	}

	if f.short {
		return f.logFile.write(bs[:len(bs)/2], at)
	}

	return 0, errors.New("QuotaExceededError: over quota")
}

func TestStoreQuotaKeepsOldValue(t *testing.T) {
	s, _ := testStore(t)

	mustSet(t, s, "k", "old")

	if _, err := s.Get("k"); err != nil {
		t.Fatal(err)
	}

	f := &failingLog{logFile: s.file, fail: true}
	s.file = f

	if err := s.Set("k", []byte("new")); err == nil {
		t.Fatal("Set over quota succeeded")
	}

	mustGet(t, s, "k", "old")
}

func TestStoreUndoesFailedAppend(t *testing.T) {
	s, name := testStore(t)

	mustSet(t, s, "a", "1")

	f := &failingLog{logFile: s.file, fail: true, short: true}
	s.file = f

	if err := s.Set("b", []byte("2")); err == nil {
		t.Fatal("short write succeeded")
	}

	f.fail = false
	mustSet(t, s, "c", "3")

	s = reopen(t, s, name)
	mustGet(t, s, "a", "1")
	mustGet(t, s, "c", "3")

	if _, err := s.Get("b"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Get of the failed write: %v", err)
	}
}

func TestStoreBeforeRun(t *testing.T) {
	_, name := testStore(t)

	bridgeRunning.Store(false)
	t.Cleanup(func() { bridgeRunning.Store(true) })

	s := OpenStore(name + "-early")
	t.Cleanup(func() {
		s.close()
		removeStoreDir(t, name+"-early")
	})

	start := time.Now()
	if _, err := s.Get("k"); !errors.Is(err, ErrBeforeRun) {
		t.Fatalf("Get before Run: %v, want ErrBeforeRun", err)
	}

	if err := s.Set("k", nil); !errors.Is(err, ErrBeforeRun) {
		t.Fatalf("Set before Run: %v, want ErrBeforeRun", err)
	}

	if d := time.Since(start); d > 100*time.Millisecond {
		t.Fatalf("calls before Run took %v", d)
	}

	bridgeRunning.Store(true)
	mustSet(t, s, "k", "v")
}

func TestStoreInvalidName(t *testing.T) {
	bridgeRunning.Store(true)

	for _, name := range []string{"", "A", "a/b", "a_b", ".."} {
		if _, err := OpenStore(name).Get("k"); err == nil {
			t.Errorf("OpenStore(%q) loaded", name)
		}
	}
}

// withGlobal shadow a property of the global object for the test.
func withGlobal(t *testing.T, prop string, value any) {
	t.Helper()

	object := js.Global().Get("Object")
	old := object.Call("getOwnPropertyDescriptor", js.Global(), prop)

	object.Call("defineProperty", js.Global(), prop, map[string]any{
		"value":        value,
		"configurable": true,
		"writable":     true,
	})

	t.Cleanup(func() {
		if old.IsUndefined() {
			js.Global().Delete(prop)
			return
		}

		object.Call("defineProperty", js.Global(), prop, old)
	})
}

func TestStoreNoFallbackToMemory(t *testing.T) {
	bridgeRunning.Store(true)

	cases := []struct {
		prop  string
		value any
		want  string
	}{
		{"FileSystemFileHandle", js.Undefined(), "dedicated Web Worker"},
		{"isSecureContext", false, "secure context"},
	}

	for _, c := range cases {
		t.Run(c.prop, func(t *testing.T) {
			withGlobal(t, c.prop, c.value)

			s := OpenStore("test-no-opfs")

			check := func(what string, err error) {
				if err == nil || !strings.Contains(err.Error(), c.want) {
					t.Errorf("%s: %v, want %q", what, err, c.want)
				}
			}

			_, err := s.Get("k")
			check("Get", err)
			check("Set", s.Set("k", nil))
			check("Delete", s.Delete("k"))
			_, err = s.Keys("")
			check("Keys", err)
			_, err = s.ReadOnly()
			check("ReadOnly", err)
		})
	}
}

func TestParseLog(t *testing.T) {
	log := append(encodeHeader(3), encodeSet("a", []byte("1"))...)
	log = append(log, encodeCommit()...)
	log = append(log, encodeSet("b", []byte("2"))...)
	log = append(log, encodeDelete("a")...)

	l := parseLog(log)
	if l.err != nil || !l.committed || l.gen != 3 || l.good != int64(len(log)) {
		t.Fatalf("parseLog = %+v", l)
	}

	if len(l.data) != 1 || string(l.data["b"]) != "2" {
		t.Fatalf("data = %q", l.data)
	}

	// Cut before the commit: not committed.
	newer := append(encodeHeader(4), encodeSet("a", []byte("1"))...)
	l = parseLog(newer[:len(newer)-1])
	if l.err != nil || l.committed || l.gen != 4 {
		t.Fatalf("uncommitted parseLog = %+v", l)
	}

	// A broken header with more after it is corruption, not an empty file.
	bad := slices.Clone(log)
	bad[recordHead] ^= 0xff
	if l := parseLog(bad); l.err == nil {
		t.Fatalf("parseLog of a broken header = %+v", l)
	}

	// Newer and uncommitted loses to older and committed.
	i, err := pickLog([2]parsedLog{parseLog(log), l})
	if err != nil || i != 0 {
		t.Fatalf("pickLog = %d, %v", i, err)
	}
}

// setCompactSlack set storeCompactSlack for the test.
func setCompactSlack(t *testing.T, v int64) {
	old := storeCompactSlack
	storeCompactSlack = v
	t.Cleanup(func() { storeCompactSlack = old })
}

// readLog parse one of a store's files. It works while the store is open.
func readLog(t *testing.T, name, file string) ([]byte, parsedLog) {
	t.Helper()

	bs, err := readSnapshot(storeDir(t, name), file)
	if err != nil {
		t.Fatal(err)
	}

	return bs, parseLog(bs)
}

func TestStoreCompactsOnOpen(t *testing.T) {
	s, name := testStore(t)

	mustSet(t, s, "keep", "k")

	value := strings.Repeat("x", 1024)
	for range 200 {
		mustSet(t, s, "draft", value)
	}

	before, old := readLog(t, name, "a")

	s = reopen(t, s, name)
	mustGet(t, s, "keep", "k")
	mustGet(t, s, "draft", value)

	// b holds one copy of the live data, under a newer generation.
	want := encodeSnapshot(old.gen+1, map[string][]byte{"keep": []byte("k"), "draft": []byte(value)})

	bs, l := readLog(t, name, "b")
	if !slices.Equal(bs, want) || !l.committed || s.end != int64(len(want)) {
		t.Fatalf("b is %d bytes, gen %d; want %d bytes, gen %d", len(bs), l.gen, len(want), old.gen+1)
	}

	// a is left as it was.
	if after, _ := readLog(t, name, "a"); !slices.Equal(after, before) {
		t.Fatalf("a changed: %d bytes, was %d", len(after), len(before))
	}

	// Appends go to b, and the next compaction overwrites a.
	mustSet(t, s, "draft", "new")

	s = reopen(t, s, name)
	mustGet(t, s, "draft", "new")
	mustGet(t, s, "keep", "k")
}

func TestStoreSmallLogNotCompacted(t *testing.T) {
	s, name := testStore(t)

	mustSet(t, s, "k", "1")
	mustSet(t, s, "k", "2")

	s = reopen(t, s, name)
	mustGet(t, s, "k", "2")

	if bs, _ := readLog(t, name, "b"); len(bs) != 0 {
		t.Fatalf("b is %d bytes, want empty", len(bs))
	}
}

// halfCompacted set up a store whose a is committed and whose b is a newer
// generation cut short by cut, as a crash mid-compaction leaves it.
func halfCompacted(t *testing.T, cut func(snap []byte) []byte) (*Store, string) {
	t.Helper()

	s, name := testStore(t)

	mustSet(t, s, "a", "1")
	mustSet(t, s, "b", "2")
	mustSet(t, s, "a", "3")
	s.close()

	_, l := readLog(t, name, "a")
	snap := encodeSnapshot(l.gen+1, l.data)
	editFile(t, name, "b", func([]byte) []byte { return cut(snap) })

	return s, name
}

func TestStoreCompactInterrupted(t *testing.T) {
	cuts := map[string]func([]byte) []byte{
		"half":      func(snap []byte) []byte { return snap[:len(snap)/2] },
		"no commit": func(snap []byte) []byte { return snap[:len(snap)-len(encodeCommit())] },
		"header":    func(snap []byte) []byte { return snap[:len(encodeHeader(1))] },
	}

	for what, cut := range cuts {
		t.Run(what, func(t *testing.T) {
			s, name := halfCompacted(t, cut)

			// The previous generation is read as it is.
			setCompactSlack(t, 1<<30)
			s = reopen(t, s, name)
			mustGet(t, s, "a", "3")
			mustGet(t, s, "b", "2")

			// A new compaction overwrites the torn one. Negative slack
			// forces it.
			setCompactSlack(t, -1<<30)
			s = reopen(t, s, name)
			mustGet(t, s, "a", "3")
			mustGet(t, s, "b", "2")

			if _, l := readLog(t, name, "b"); !l.committed {
				t.Fatalf("b not committed after compaction: %+v", l)
			}

			mustSet(t, s, "c", "4")
			s = reopen(t, s, name)
			mustGet(t, s, "c", "4")
		})
	}
}

func TestStoreReadOnlyDuringCompaction(t *testing.T) {
	s, name := halfCompacted(t, func(snap []byte) []byte { return snap[:len(snap)/2] })

	// Another tab owns the store and is writing b.
	lock, err := openSyncFile(storeDir(t, name), storeLockName)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.close()

	s = reopen(t, s, name)

	if ro, err := s.ReadOnly(); err != nil || !ro {
		t.Fatalf("ReadOnly = %v, %v; want true", ro, err)
	}

	mustGet(t, s, "a", "3")
	mustGet(t, s, "b", "2")
}

func TestEncodeSnapshot(t *testing.T) {
	snap := encodeSnapshot(7, map[string][]byte{"b": []byte("2"), "a": []byte("1")})

	l := parseLog(snap)
	if l.err != nil || !l.committed || l.gen != 7 || l.good != int64(len(snap)) {
		t.Fatalf("parseLog = %+v", l)
	}

	if len(l.data) != 2 || string(l.data["a"]) != "1" || string(l.data["b"]) != "2" {
		t.Fatalf("data = %q", l.data)
	}

	big := map[string][]byte{"": nil, strings.Repeat("k", 300): []byte(strings.Repeat("v", 70000))}
	for _, data := range []map[string][]byte{nil, l.data, big} {
		if got, want := snapshotSize(data), len(encodeSnapshot(7, data)); got != int64(want) {
			t.Fatalf("snapshotSize = %d, want %d", got, want)
		}
	}

	if !slices.Equal(snap, encodeSnapshot(7, map[string][]byte{"a": []byte("1"), "b": []byte("2")})) {
		t.Fatal("encodeSnapshot is not deterministic")
	}
}
