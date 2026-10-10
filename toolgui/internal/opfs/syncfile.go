//go:build js && wasm

package opfs

import (
	"io"
	"syscall/js"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// ChunkSize is the most bytes one read or write moves through the typed
// array.
const ChunkSize = 64 << 10

// SyncFile is a file behind a sync access handle. Its methods are synchronous
// and not safe for concurrent use.
type SyncFile struct {
	handle js.Value

	// buf is reused across calls, grown up to ChunkSize.
	buf js.Value
}

// NewSyncFile wraps an open sync access handle.
func NewSyncFile(handle js.Value) *SyncFile {
	return &SyncFile{handle: handle}
}

// OpenSyncFile opens name in dir, creating it if missing. It awaits, so it
// must not be called from a [js.Func] callback.
func OpenSyncFile(dir js.Value, name string) (*SyncFile, error) {
	fh, err := AwaitCall(dir, "getFileHandle", name, Create)
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	h, err := AwaitCall(fh, "createSyncAccessHandle")
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	return NewSyncFile(h), nil
}

// Size returns the file size.
func (f *SyncFile) Size() (int64, error) {
	size, err := Call(f.handle, "getSize")
	if err != nil {
		return 0, tgutil.Errorf("%w", err)
	}

	return int64(size.Float()), nil
}

// ReadAt reads len(p) bytes at off, a chunk at a time. A short read returns
// io.EOF.
func (f *SyncFile) ReadAt(p []byte, off int64) (int, error) {
	read := 0
	for read < len(p) {
		view := f.view(len(p) - read)

		got, err := Call(f.handle, "read", view, At(off+int64(read)))
		if err != nil {
			return read, tgutil.Errorf("%w", err)
		}

		n := got.Int()
		if n == 0 {
			return read, io.EOF
		}

		js.CopyBytesToGo(p[read:read+n], view)
		read += n
	}

	return read, nil
}

// ReadAll reads the whole file.
func (f *SyncFile) ReadAll() ([]byte, error) {
	size, err := f.Size()
	if err != nil {
		return nil, err
	}

	if size == 0 {
		return nil, nil
	}

	bs := make([]byte, size)

	n, err := f.ReadAt(bs, 0)
	if err == io.EOF {
		return nil, tgutil.Errorf("read %d of %d bytes: %w", n, size, io.ErrUnexpectedEOF)
	}

	if err != nil {
		return nil, err
	}

	return bs, nil
}

// WriteAt writes p at off, a chunk at a time. A short write is an error.
// A full origin fails with QuotaExceededError.
func (f *SyncFile) WriteAt(p []byte, off int64) (int, error) {
	written := 0
	for written < len(p) {
		chunk := p[written:min(len(p), written+ChunkSize)]

		view := f.view(len(chunk))
		js.CopyBytesToJS(view, chunk)

		got, err := Call(f.handle, "write", view, At(off+int64(written)))
		if err != nil {
			return written, tgutil.Errorf("%w", err)
		}

		n := int(got.Float())
		written += n

		if n != len(chunk) {
			return written, tgutil.Errorf("wrote %d of %d bytes", written, len(p))
		}
	}

	return written, nil
}

// Truncate resizes the file.
func (f *SyncFile) Truncate(size int64) error {
	_, err := Call(f.handle, "truncate", float64(size))
	return err
}

// Flush persists written bytes.
func (f *SyncFile) Flush() error {
	_, err := Call(f.handle, "flush")
	return err
}

// Close releases the handle. The SyncFile must not be used after.
func (f *SyncFile) Close() error {
	_, err := Call(f.handle, "close")
	f.handle, f.buf = js.Undefined(), js.Undefined()
	return err
}

// view returns an n-byte view of buf, n capped at ChunkSize.
func (f *SyncFile) view(n int) js.Value {
	n = min(n, ChunkSize)

	if !f.buf.Truthy() || f.buf.Length() < n {
		size := n
		if f.buf.Truthy() {
			size = min(max(n, 2*f.buf.Length()), ChunkSize)
		}

		f.buf = Uint8Array.New(size)
	}

	if f.buf.Length() == n {
		return f.buf
	}

	return f.buf.Call("subarray", 0, n)
}
