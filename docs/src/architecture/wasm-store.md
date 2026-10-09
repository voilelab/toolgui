# Persistent Store (wasm)

> Design note (TG-102). It records the decisions that the implementation tickets
> build on. The store is implemented (TG-107, `toolgui/tgwasm/store.go`).
> Compaction (TG-108) and the JSON helpers and docs (TG-109) are not yet.

A wasm app keeps drafts and records that should survive a reload or a closed
tab. Today it has no API for that. offline-judge wraps IndexedDB by hand: about
370 lines of `syscall/js` with `await`, `jsTry` and cursor scans, plus a Go
global as an in-memory copy.

The hard part is not storage. A Go function on the JavaScript callback stack
cannot block, and every browser storage API except a sync access handle
answers with a promise. `tgframe/file_opfs.go` has already solved this for
uploads. The store reuses the same approach.

## Decision summary

| Question | Decision |
| --- | --- |
| Backing | OPFS. A store is one append-only log, read into memory when it opens. Each write is appended and flushed through a sync access handle before the call returns. |
| API | `tgwasm.OpenStore(name)` returns a `*Store` with `Get`, `Set`, `Delete` and `Keys(prefix)` on `[]byte`. `GetJSON` and `SetJSON` use `tgjson`. Namespaces are key prefixes. |
| Multiple tabs | The first tab owns the store. Later tabs open it read-only, from a snapshot taken when they opened it. A write there returns `ErrReadOnly`. There is no `BroadcastChannel` coordination. |
| Failures | No secure context, no worker, or a full quota is an error. The store never falls back to memory. |
| Other executors | Only `tgwasm`. Web does not get one on purpose. Wails apps can write files directly. There is no `tgframe` interface yet. |

## Backing: an OPFS log, not IndexedDB

IndexedDB has no synchronous API, so every call would be async. That is the
problem the app is trying to get away from.

OPFS with a sync access handle gives synchronous reads and writes once the
handle is open. Only opening the handle is async, and that happens once per
store.

* **In memory.** The whole store is a `map[string][]byte`. `Get` and `Keys`
  never touch the file. This is meant for KB to MB of data, not for a database.
* **Write-through.** `Set` and `Delete` append one record and call `flush()`
  before they return. When the call returns `nil`, the change is on disk. The
  map is updated only after the write succeeds, so a write that fails over
  quota leaves the old value in place.
* **A failed append is undone.** A short `write` or a failed `flush()` can leave
  part of the record on disk. A later record appended after it would be lost on
  the next open, even though its `Set` returned `nil`. So the store truncates
  the file back to its length before the append and flushes again. If that
  fails too, the store is broken: every later write returns that error, and
  nothing more is appended.
* **Append-only log.** Rewriting the whole file on every `Set` would cost the
  size of the store per write. Appending costs the size of one record. Each
  record has a length and a CRC. On open, a record cut short by a crash or a
  closed tab is dropped, together with anything after it.
* **Compaction on open.** When the log is much larger than its live data, the
  store writes the live data to the other of two files (`a` and `b`) under a
  higher generation number, ends it with a commit record, and flushes. The old
  file is left as it is, and the next compaction overwrites it. On open, the
  store reads the newest generation that has its commit record. A crash during
  compaction loses nothing, because the previous generation is still complete.

### Layout

```
toolgui-kv/            separate from toolgui-state, so the upload sweep never touches it
  <name>/
    .lock              its sync access handle is held for as long as the store is open
    a, b               generations of the log
```

Keeping the previous generation also protects a read-only tab that opens
during a compaction (see [Multiple tabs](#multiple-tabs)). Its two `getFile()`
calls can straddle the compaction, but the older file is never emptied, so at
least one of them is complete.

The name must match `[a-z0-9-]+`. One origin can hold several stores, for
example one per app on a shared host.

## Opening without blocking boot

`worker.ts` calls `go.run` and expects the bridge to exist when the Go program
first blocks. If `main` waited on a promise before `Run()`, the bridge would
not be there yet, and the boot would fail with *the wasm program installed no
bridge*.

So `OpenStore` returns at once and loads on a goroutine of its own, the same
way `opfsStateRoot` does. When a method is called depends on `Run`:

* **Before `Run`** installs the bridge, for example from `main` or from an app
  constructor, a method returns `ErrBeforeRun` at once. Waiting there would
  block `main` before the bridge exists, and boot would fail.
* **After `Run`**, a method waits for the load on its first call. That is safe
  because page functions run on the bridge's run goroutine (`b.runs.do`), not on
  the callback stack.

After the load, no method waits for anything. A value needed at startup is
read in the first page run, not in `main`.

```go
var store = tgwasm.OpenStore("judge")

func main() {
	app := tgframe.NewApp()
	app.AddPage("index", "Index", Index)
	tgwasm.NewExecutor(app).Run()
}

func Index(p *tgframe.Params) error {
	draft, err := tgwasm.GetJSON[Draft](store, "draft/0004/go")
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	...
	return tgwasm.SetJSON(store, "draft/0004/go", draft)
}
```

## API

```go
func OpenStore(name string) *Store

func (s *Store) Get(key string) ([]byte, error)   // fs.ErrNotExist when absent
func (s *Store) Set(key string, value []byte) error
func (s *Store) Delete(key string) error          // nil when absent
func (s *Store) Keys(prefix string) ([]string, error) // sorted
func (s *Store) ReadOnly() (bool, error)

func GetJSON[T any](s *Store, key string) (T, error)
func SetJSON(s *Store, key string, v any) error

var ErrReadOnly  // a write in a tab that does not own the store
var ErrBeforeRun // a call before Run installed the bridge
```

* **`[]byte` at the core.** Values are `[]byte`. Typed access goes through
  `tgjson`, so the store does not depend on an encoding.
* **`Get` returns an error, not `ok`.** A store that failed to load has to say
  so on every call. With `(value, ok)`, a load failure would look like an
  empty store, which is the silent fallback this design rejects. A missing
  key is `fs.ErrNotExist`, the same convention as `os.ReadFile`.
* **Prefixes, not namespaces.** `Keys("sub/0004/")` lists all submissions for
  one problem. Keys are sorted, so a timestamp in the key gives time order. A
  separate namespace type would add nothing that a prefix does not already do.
* `Get` returns a copy, and `Set` stores a copy, so the caller can reuse its
  slices.

## Multiple tabs

A sync access handle is exclusive, so a second tab of the same app cannot get
one.

| Option | Verdict |
| --- | --- |
| First tab owns, later tabs read-only | **Chosen.** It is simple and predictable, and the app can tell the user. |
| `BroadcastChannel`: one tab writes for the others | Rejected. It needs leader election, request and reply over messages, and handover when the leader closes. That is a distributed system inside one browser. |
| `createSyncAccessHandle({mode: "readwrite-unsafe"})` | Rejected. Chrome only, and each tab's in-memory copy would drift from the file. |
| IndexedDB, which allows many tabs | Rejected. Async, as explained above. |

The tab that gets the `.lock` handle owns the store. A tab that is refused the
lock (`NoModificationAllowedError`) reads the newest committed generation
with `getFile()`, which another tab's sync access handle does not block, and
opens read-only. If neither file has a committed generation, or a read fails
with `NotReadableError` because the owner changed the file mid-read, it reads
both files again, a few times, before it reports an error.

In a read-only tab:

* `Get` and `Keys` answer from that snapshot. Writes from the owning tab do not
  show up in it.
* `Set` and `Delete` return `ErrReadOnly`.
* `ReadOnly()` reports the mode, so a page can show *open in another tab*.

A read-only tab stays read-only after the owner closes. It upgrades on reload.
Upgrading in place with `navigator.locks` is possible later, but it is not part
of the first version.

## Failures

The rules match `file_opfs.go`. A storage problem is an error, never a silent
fallback to memory:

* Not in a dedicated worker, or not a secure context: the load fails, and every
  method returns that error, using the same wording as the upload store.
* Over quota: `Set` returns the `QuotaExceededError`, and the map keeps the old
  value.
* A corrupt record in the middle of the log, not just at the tail: the load
  fails. The store does not guess which records are still good.

Browsers may evict OPFS data under storage pressure. `OpenStore` does not call
`navigator.storage.persist()`, because some browsers show a prompt. An app that
needs the guarantee asks for it itself.

## Other executors

* **Web: deliberately none.** The server does not know which person is on the
  other end, so a server-side store would be shared by every user. An app that
  needs this stores data under its own login, in its own database.
* **Wails: none for now.** A desktop app can read and write files under
  `os.UserConfigDir()` directly, without async APIs or exclusive handles. That
  is the problem this store solves, and Wails does not have it.
* **No `tgframe` interface yet.** A shared interface is worth adding when a
  second implementation exists. The method set above already fits a
  file-backed version for Wails, so adding one later will not change the API.

## Go globals are still fine

In wasm, a Go global is the app state for one tab. It is not shared with other
users, as it would be on a server, so it is a valid place for data that only
needs to last until a reload. The store is for data that must survive the
reload. See [App Cache](app-cache.md#in-wasm).

## Out of scope

* **Upgrading in place** from read-only to owner when the owning tab closes.
* **Live updates across tabs.** A read-only tab does not see new writes.
* **Large data and blobs.** Files belong in the upload and download file store.
  This store keeps everything in memory.
* **Transactions** across keys. Store related data under one key, as one JSON
  value.
