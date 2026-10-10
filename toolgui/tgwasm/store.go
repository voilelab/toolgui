//go:build js && wasm

package tgwasm

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"io/fs"
	"maps"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"syscall/js"
	"time"

	"github.com/voilelab/toolgui/toolgui/internal/opfs"
	"github.com/voilelab/toolgui/toolgui/tgjson"
	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// ErrReadOnly is returned by a write to a store another tab owns.
var ErrReadOnly = tgutil.NewError("store is read-only: another tab owns it")

// ErrBeforeRun is returned by a store call made before [Executor.Run]
// installed the bridge. Waiting there would block main, and boot would fail.
var ErrBeforeRun = tgutil.NewError("store used before Run, read it in a page run instead")

// bridgeRunning is set once the bridge is installed.
var bridgeRunning atomic.Bool

const (
	// storeRootName is the stores' directory. It is not under toolgui-state,
	// so the upload sweep never sees it.
	storeRootName = "toolgui-kv"
	storeLockName = ".lock"

	// storeReadTries is how many times a read-only tab reads the log before
	// it gives up: the owner may be mid-write.
	storeReadTries = 5
)

// storeReadGap is the pause between those reads. A variable for tests.
var storeReadGap = 100 * time.Millisecond

// storeCompactSlack is how far a log may outgrow twice its live data before
// an open compacts it. A variable for tests.
var storeCompactSlack int64 = 64 << 10

// storeFiles are the two generations of a log.
var storeFiles = [2]string{"a", "b"}

var storeNameRe = regexp.MustCompile(`^[a-z0-9-]+$`)

// Store is a persistent key-value store kept in the origin private file
// system. It is read into memory when it opens, and each write is appended
// and flushed before the call returns.
//
// The first tab to open a store owns it. Later tabs open it read-only, from
// a snapshot taken when they opened it.
//
// See docs/src/architecture/wasm-store.md.
type Store struct {
	name string

	// ready is closed once the load is done. err and readOnly are set before.
	ready    chan struct{}
	err      error
	readOnly bool

	mu   sync.Mutex
	data map[string][]byte

	// Owner only.
	lock *opfs.SyncFile
	file logFile
	end  int64

	// broken is set when a failed append could not be undone.
	broken error
}

// OpenStore return the store called name, which must match [a-z0-9-]+. It
// returns at once and loads on a goroutine of its own; its methods wait for
// the load on first call, and fail with [ErrBeforeRun] before [Executor.Run].
func OpenStore(name string) *Store {
	s := &Store{name: name, ready: make(chan struct{})}

	if !storeNameRe.MatchString(name) {
		s.err = tgutil.Errorf("invalid store name %q, want [a-z0-9-]+", name)
		close(s.ready)
		return s
	}

	go s.load()

	return s
}

// wait blocks until the store is loaded.
func (s *Store) wait() error {
	if !bridgeRunning.Load() {
		return ErrBeforeRun
	}

	<-s.ready
	return s.err
}

// Get return a copy of the value under key, or [fs.ErrNotExist].
func (s *Store) Get(key string) ([]byte, error) {
	if err := s.wait(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data[key]
	if !ok {
		return nil, fs.ErrNotExist
	}

	return bytes.Clone(v), nil
}

// Set store a copy of value under key. When it returns nil, the value is on
// disk.
func (s *Store) Set(key string, value []byte) error {
	if err := s.wait(); err != nil {
		return err
	}

	if s.readOnly {
		return ErrReadOnly
	}

	value = bytes.Clone(value)
	if value == nil {
		value = []byte{}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.append(encodeSet(key, value)); err != nil {
		return err
	}

	s.data[key] = value
	return nil
}

// Delete remove key. A missing key is not an error.
func (s *Store) Delete(key string) error {
	if err := s.wait(); err != nil {
		return err
	}

	if s.readOnly {
		return ErrReadOnly
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[key]; !ok {
		return nil
	}

	if err := s.append(encodeDelete(key)); err != nil {
		return err
	}

	delete(s.data, key)
	return nil
}

// Keys return the keys starting with prefix, sorted.
func (s *Store) Keys(prefix string) ([]string, error) {
	if err := s.wait(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	keys := []string{}
	for k := range s.data {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}

	slices.Sort(keys)
	return keys, nil
}

// ReadOnly report whether another tab owns the store.
func (s *Store) ReadOnly() (bool, error) {
	if err := s.wait(); err != nil {
		return false, err
	}

	return s.readOnly, nil
}

// GetJSON decode the value under key into a T, or return [fs.ErrNotExist].
func GetJSON[T any](s *Store, key string) (T, error) {
	var v T

	bs, err := s.Get(key)
	if err != nil {
		return v, err
	}

	if err := tgjson.Unmarshal(bs, &v); err != nil {
		return v, tgutil.Errorf("store %q, key %q: %w", s.name, key, err)
	}

	return v, nil
}

// SetJSON encode v as JSON and store it under key.
func SetJSON(s *Store, key string, v any) error {
	bs, err := tgjson.Marshal(v)
	if err != nil {
		return tgutil.Errorf("store %q, key %q: %w", s.name, key, err)
	}

	return s.Set(key, bs)
}

// append write one record at the end and flush it. A failed append is
// truncated away, so a later one is not lost behind it on the next open. It
// must be called with mu held.
func (s *Store) append(rec []byte) error {
	if s.broken != nil {
		return s.broken
	}

	at := s.end

	n, err := s.file.WriteAt(rec, at)
	if err == nil && n != len(rec) {
		err = tgutil.Errorf("wrote %d of %d bytes", n, len(rec))
	}

	if err == nil {
		err = s.file.Flush()
	}

	if err == nil {
		s.end += int64(len(rec))
		return nil
	}

	if undo := s.file.Truncate(at); undo != nil {
		s.broken = tgutil.Errorf("store %q is broken: %w", s.name, undo)
	} else if undo := s.file.Flush(); undo != nil {
		s.broken = tgutil.Errorf("store %q is broken: %w", s.name, undo)
	}

	return tgutil.Errorf("%w", err)
}

// close release the handles. Only tests call it: a store lives as long as
// the tab.
func (s *Store) close() {
	<-s.ready

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.file != nil {
		_ = s.file.Close()
		s.file = nil
	}

	if s.lock != nil {
		_ = s.lock.Close()
		s.lock = nil
	}
}

func (s *Store) load() {
	defer close(s.ready)

	s.err = s.open()
}

func (s *Store) open() error {
	origin, err := opfs.Origin()
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	root, err := opfs.AwaitCall(origin, "getDirectoryHandle", storeRootName, opfs.Create)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	dir, err := opfs.AwaitCall(root, "getDirectoryHandle", s.name, opfs.Create)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	lock, err := opfs.OpenSyncFile(dir, storeLockName)
	if err != nil {
		if strings.Contains(err.Error(), "NoModificationAllowedError") {
			s.readOnly = true
			return s.openReadOnly(dir)
		}

		return tgutil.Errorf("%w", err)
	}

	s.lock = lock

	if err := s.openOwner(dir); err != nil {
		_ = lock.Close()
		s.lock = nil
		return err
	}

	return nil
}

// openOwner read the newest committed generation and keep its file open for
// appends.
func (s *Store) openOwner(dir js.Value) error {
	var files [2]*opfs.SyncFile
	var logs [2]parsedLog

	defer func() {
		for _, f := range files {
			if f != nil && logFile(f) != s.file {
				_ = f.Close()
			}
		}
	}()

	for i, name := range storeFiles {
		f, err := opfs.OpenSyncFile(dir, name)
		if err != nil {
			return tgutil.Errorf("%w", err)
		}

		files[i] = f

		bs, err := f.ReadAll()
		if err != nil {
			return tgutil.Errorf("%w", err)
		}

		logs[i] = parseLog(bs)
	}

	i, err := pickLog(logs)
	if err != nil {
		return tgutil.Errorf("store %q: %w", s.name, err)
	}

	if i < 0 {
		// Nothing committed yet: start a new generation in a.
		i = 0
		gen := max(logs[0].gen, logs[1].gen) + 1
		rec := encodeSnapshot(gen, nil)

		if err := writeGen(files[0], rec); err != nil {
			return tgutil.Errorf("%w", err)
		}

		logs[0] = parsedLog{gen: gen, committed: true, data: map[string][]byte{}, good: int64(len(rec))}
	}

	f, l := files[i], logs[i]

	// Compact into the other file. The old one is kept as it is: a crash or
	// a read-only tab mid-compaction still finds it complete.
	if l.good > 2*snapshotSize(l.data)+storeCompactSlack {
		j := 1 - i
		snap := encodeSnapshot(l.gen+1, l.data)

		if err := writeGen(files[j], snap); err == nil {
			f = files[j]
			l.good, l.size = int64(len(snap)), int64(len(snap))
		} else {
			// Over quota, most likely. Keep the old log and free the space;
			// an uncommitted generation is ignored anyway.
			_ = files[j].Truncate(0)
			_ = files[j].Flush()
		}
	}

	// Drop a torn tail, so the next append is not lost behind it.
	if l.good < l.size {
		if err := f.Truncate(l.good); err != nil {
			return tgutil.Errorf("%w", err)
		}

		if err := f.Flush(); err != nil {
			return tgutil.Errorf("%w", err)
		}
	}

	s.data = l.data
	s.file = f
	s.end = l.good

	return nil
}

// openReadOnly read a snapshot through getFile, which the owner's sync access
// handle does not block. The owner may be mid-write, so it retries.
func (s *Store) openReadOnly(dir js.Value) error {
	var last error

	for try := range storeReadTries {
		if try > 0 {
			time.Sleep(storeReadGap)
		}

		var logs [2]parsedLog

		last = nil
		for i, name := range storeFiles {
			bs, err := readSnapshot(dir, name)
			if err != nil {
				last = err
				break
			}

			logs[i] = parseLog(bs)
		}

		if last != nil {
			continue
		}

		i, err := pickLog(logs)
		if err != nil {
			last = err
			continue
		}

		if i < 0 {
			last = tgutil.NewError("no committed generation")
			continue
		}

		s.data = logs[i].data
		return nil
	}

	return tgutil.Errorf("store %q: %w", s.name, last)
}

// readSnapshot read a whole file with getFile. A missing file reads empty.
func readSnapshot(dir js.Value, name string) ([]byte, error) {
	handle, err := opfs.AwaitCall(dir, "getFileHandle", name)
	if err != nil {
		if strings.Contains(err.Error(), "NotFoundError") {
			return nil, nil
		}

		return nil, tgutil.Errorf("%w", err)
	}

	file, err := opfs.AwaitCall(handle, "getFile")
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	buf, err := opfs.AwaitCall(file, "arrayBuffer")
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	arr := opfs.Uint8Array.New(buf)
	bs := make([]byte, arr.Length())
	js.CopyBytesToGo(bs, arr)

	return bs, nil
}

// pickLog return the index of the newest committed log, or -1 when there is
// none. A corrupt file is an error: which one held the data is a guess.
func pickLog(logs [2]parsedLog) (int, error) {
	for _, l := range logs {
		if l.err != nil {
			return -1, l.err
		}
	}

	order := []int{0, 1}
	if logs[1].gen > logs[0].gen {
		order = []int{1, 0}
	}

	for _, i := range order {
		if logs[i].committed {
			return i, nil
		}
	}

	return -1, nil
}

// writeGen replace f's contents with a whole generation and flush. It
// empties f first, so a crash mid-write leaves a torn tail, not new records
// in front of old bytes.
func writeGen(f logFile, rec []byte) error {
	if err := f.Truncate(0); err != nil {
		return err
	}

	if err := f.Flush(); err != nil {
		return err
	}

	n, err := f.WriteAt(rec, 0)
	if err != nil {
		return err
	}

	if n != len(rec) {
		return tgutil.Errorf("wrote %d of %d bytes", n, len(rec))
	}

	return f.Flush()
}

// logFile is what the owner appends through. An interface so tests can fail
// a write.
type logFile interface {
	WriteAt(bs []byte, at int64) (int, error)
	Flush() error
	Truncate(size int64) error
	Close() error
}

// A log is records of [length uint32][crc32 uint32][payload]. It starts with
// a header naming its generation, then the snapshot's sets, then a commit;
// appends follow the commit.
const (
	opHeader byte = iota + 1
	opCommit
	opSet
	opDelete

	recordHead   = 8
	storeVersion = 1
)

var storeMagic = []byte("tgkv")

func encodeRecord(payload []byte) []byte {
	rec := make([]byte, recordHead, recordHead+len(payload))
	binary.LittleEndian.PutUint32(rec[0:], uint32(len(payload)))
	binary.LittleEndian.PutUint32(rec[4:], crc32.ChecksumIEEE(payload))

	return append(rec, payload...)
}

func encodeHeader(gen uint64) []byte {
	p := append([]byte{opHeader}, storeMagic...)
	p = append(p, storeVersion)
	p = binary.LittleEndian.AppendUint64(p, gen)

	return encodeRecord(p)
}

func encodeCommit() []byte {
	return encodeRecord([]byte{opCommit})
}

// encodeSnapshot return a whole generation: header, data in key order,
// commit.
func encodeSnapshot(gen uint64, data map[string][]byte) []byte {
	rec := encodeHeader(gen)

	for _, k := range slices.Sorted(maps.Keys(data)) {
		rec = append(rec, encodeSet(k, data[k])...)
	}

	return append(rec, encodeCommit()...)
}

// snapshotSize return len(encodeSnapshot(gen, data)) without encoding it.
func snapshotSize(data map[string][]byte) int64 {
	n := int64(recordHead+1+len(storeMagic)+1+8) + recordHead + 1

	for k, v := range data {
		klen := len(binary.AppendUvarint(nil, uint64(len(k))))
		n += int64(recordHead + 1 + klen + len(k) + len(v))
	}

	return n
}

func encodeSet(key string, value []byte) []byte {
	p := []byte{opSet}
	p = binary.AppendUvarint(p, uint64(len(key)))
	p = append(p, key...)
	p = append(p, value...)

	return encodeRecord(p)
}

func encodeDelete(key string) []byte {
	return encodeRecord(append([]byte{opDelete}, key...))
}

// parsedLog is one file read back.
type parsedLog struct {
	gen       uint64 // 0 when there is no valid header
	committed bool
	data      map[string][]byte

	// good is where the last whole record ends, size the file's length.
	good int64
	size int64

	// err is corruption that is not a torn tail.
	err error
}

// parseLog read a log. A record cut short at the end is a torn tail and
// dropped. A bad record with more after it is corruption.
func parseLog(bs []byte) parsedLog {
	l := parsedLog{data: map[string][]byte{}, size: int64(len(bs))}

	off := 0
	for {
		payload, next, ok := readRecord(bs, off)
		if !ok {
			if next < len(bs) {
				// A bad checksum with more after it.
				l.err = tgutil.Errorf("corrupt record at offset %d", off)
			}

			break
		}

		if err := l.apply(payload); err != nil {
			l.err = tgutil.Errorf("corrupt record at offset %d: %w", off, err)
			break
		}

		off = next
		l.good = int64(off)
	}

	if l.err != nil || !l.committed {
		l.data = map[string][]byte{}
		l.good = 0
	}

	return l
}

// readRecord return the payload at off and where the next record starts. ok
// is false for a record cut short, where next is len(bs), or one whose
// checksum does not match.
func readRecord(bs []byte, off int) (payload []byte, next int, ok bool) {
	if len(bs)-off < recordHead {
		return nil, len(bs), false
	}

	n := int(binary.LittleEndian.Uint32(bs[off:]))
	sum := binary.LittleEndian.Uint32(bs[off+4:])

	start := off + recordHead
	if n > len(bs)-start {
		return nil, len(bs), false
	}

	next = start + n
	payload = bs[start:next]

	if crc32.ChecksumIEEE(payload) != sum {
		return nil, next, false
	}

	return payload, next, true
}

func (l *parsedLog) apply(p []byte) error {
	if len(p) == 0 {
		return tgutil.NewError("empty record")
	}

	op, rest := p[0], p[1:]

	if l.gen == 0 {
		if op != opHeader || len(rest) != len(storeMagic)+1+8 ||
			!bytes.Equal(rest[:len(storeMagic)], storeMagic) ||
			rest[len(storeMagic)] != storeVersion {
			return tgutil.NewError("no header")
		}

		l.gen = binary.LittleEndian.Uint64(rest[len(storeMagic)+1:])
		if l.gen == 0 {
			return tgutil.NewError("generation 0")
		}

		return nil
	}

	switch op {
	case opCommit:
		if l.committed {
			return tgutil.NewError("second commit")
		}

		l.committed = true
	case opSet:
		klen, n := binary.Uvarint(rest)
		if n <= 0 || klen > uint64(len(rest)-n) {
			return tgutil.NewError("bad key length")
		}

		key := string(rest[n : n+int(klen)])
		l.data[key] = bytes.Clone(rest[n+int(klen):])
	case opDelete:
		delete(l.data, string(rest))
	default:
		return tgutil.Errorf("unknown op %d", op)
	}

	return nil
}
