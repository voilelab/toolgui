# Persistent Store (wasm)

How `tgwasm.Store` (`toolgui/tgwasm/store.go`) works. For usage, see
[Keeping data across reloads](../hello-world/wasm.md#keeping-data-across-reloads).

A Go function on the JavaScript callback stack cannot block, and every browser
storage API except a sync access handle answers with a promise. So the store
uses OPFS with a sync access handle, the same approach as
`tgframe/file_opfs.go`. Only opening the handle is async, once per store.
IndexedDB is ruled out: it has no synchronous API.

## The log

* **In memory.** The whole store is a `map[string][]byte`. `Get` and `Keys`
  never touch the file. It is meant for KB to MB of data, not for a database.
* **Write-through.** `Set` and `Delete` append one record and `flush()` before
  they return. The map is updated only after the write succeeds, so a write
  over quota keeps the old value.
* **A failed append is undone.** The file is truncated back to its length
  before the append, so a later record is not lost behind a partial one. If
  that fails too, the store is broken and every later write returns the error.
* **Records** have a length and a CRC. On open, a record cut short at the tail
  is dropped. A corrupt record in the middle fails the load.
* **Compaction on open.** When the log is more than twice the compacted size
  plus 64 KiB, the live data is written to the other of two files under a
  higher generation, ended with a commit record. The old file is left until the
  next compaction, so a crash loses nothing. On open, the newest committed
  generation wins. A failed compaction empties the other file and keeps the old
  one. Read-only tabs never compact.

```
toolgui-kv/            separate from toolgui-state, so the upload sweep skips it
  <name>/              [a-z0-9-]+
    .lock              its sync access handle is held while the store is open
    a, b               generations of the log
```

## Opening without blocking boot

`worker.ts` expects the bridge to exist when the Go program first blocks, so
`OpenStore` returns at once and loads on its own goroutine, like
`opfsStateRoot`.

* Before `Run` installs the bridge, methods return `ErrBeforeRun`.
* After `Run`, the first call waits for the load. Page functions run on the
  bridge's run goroutine, not the callback stack, so that is safe.

## Multiple tabs

A sync access handle is exclusive. The tab that gets the `.lock` handle owns
the store. A tab refused it (`NoModificationAllowedError`) reads the newest
committed generation with `getFile()` and opens read-only, retrying a few times
if the owner changes the file mid-read. Keeping the previous generation means
at least one file is complete even during a compaction.

A read-only tab answers from its snapshot, returns `ErrReadOnly` on writes, and
stays read-only until reload. `BroadcastChannel` coordination and
`readwrite-unsafe` handles were rejected as too complex and Chrome-only.

## Failures

Never a silent fallback to memory. No dedicated worker or no secure context
fails the load, and every method returns that error. `persist()` is not called,
since some browsers prompt for it.

## Other executors

* **Web: none.** A server-side store would be shared by every user.
* **Wails: none yet.** It can write under `os.UserConfigDir()` directly. The
  method set fits a file-backed version, so a `tgframe` interface can come
  with a second implementation.

## Out of scope

* Upgrading a read-only tab in place, or live updates across tabs.
* Large data and blobs: use the file store.
* Transactions across keys: keep related data under one key.
