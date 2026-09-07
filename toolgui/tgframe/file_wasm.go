//go:build js && wasm

package tgframe

import (
	"bytes"
	"io"
	"sync"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// memBodies keeps a state's uploads in memory. In the browser there is no
// filesystem to put them on: wasm_exec.js answers open and mkdir with ENOSYS.
// A file therefore has to fit in the tab, the same as it did to get here.
type memBodies struct{}

func newFileBodies() fileBodies {
	return memBodies{}
}

func (memBodies) newBody() (fileBody, error) {
	return &memBody{}, nil
}

// destroy has nothing to do: the store drops the files, and with them the
// last reference to what they held.
func (memBodies) destroy() {}

type memBody struct {
	lock sync.Mutex
	bs   []byte
}

func (b *memBody) open() (FileReader, error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	// The reader is capped at what is there now. An append can only add past
	// that, so what it reads doesn't change under it.
	return nopCloser{bytes.NewReader(b.bs)}, nil
}

func (b *memBody) write(r io.Reader, atEnd bool) (int64, error) {
	bs, err := io.ReadAll(r)
	if err != nil {
		return 0, tgutil.Errorf("%w", err)
	}

	b.lock.Lock()
	defer b.lock.Unlock()

	if atEnd {
		b.bs = append(b.bs, bs...)
	} else {
		b.bs = bs
	}

	return int64(len(bs)), nil
}

func (b *memBody) remove() {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.bs = nil
}

// nopCloser gives a bytes.Reader the Close a [FileReader] needs.
type nopCloser struct {
	*bytes.Reader
}

func (nopCloser) Close() error { return nil }
