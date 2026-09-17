//go:build js && wasm

package tgframe

import (
	"syscall/js"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// BrowserUpload is a file the page writes into the state's storage itself,
// rather than handing its bytes to Go.
//
// The split is forced by the browser, not chosen. A writable stream is the
// only way to copy a picked file without the whole of it passing through the
// tab's heap, and a sync access handle is the only way to read one back
// without awaiting a promise -- which the store cannot do, because an upload
// arrives on the JavaScript callback stack. Neither side can do both: the two
// are exclusive holds on the same file. So the page writes and Go reads, in
// that order, and nobody should move either half back across.
//
// The order matters. [BrowserUpload.Take] must not be called until the page
// has closed its writable stream and opened a sync access handle of its own;
// asking for one while the write is still open gets NoModificationAllowedError.
type BrowserUpload struct {
	body *opfsBody
}

// NewBrowserUpload reserves an empty file for the page to stream an upload
// into. Nothing is written and no handle is open: the file is the page's until
// [BrowserUpload.Take] or [BrowserUpload.Discard] says what became of it.
//
// It is safe to call from a JavaScript callback, which is where an upload
// begins. Like [State.NewFile] it fails rather than waits.
func (s *State) NewBrowserUpload() (*BrowserUpload, error) {
	bodies, ok := s.files.bodies.(*opfsBodies)
	if !ok {
		return nil, tgutil.NewError("this state does not keep its files in the browser")
	}

	body, err := bodies.reserve()
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	return &BrowserUpload{body: body}, nil
}

// Dir is the reserved file's directory, named from the origin private file
// system's root down, so the page can walk to it with getDirectoryHandle.
func (u *BrowserUpload) Dir() []string {
	return u.body.bodies.dirPath()
}

// Name is what the reserved file is called inside [BrowserUpload.Dir]. It is
// the state's own name for the file and has nothing to do with the name the
// user's file had: that one is the browser's to choose, and this one names an
// entry the store removes by itself.
func (u *BrowserUpload) Name() string {
	return u.body.name
}

// Take turns a written reservation into a file under the given upload name.
// handle is the sync access handle the page opened once its writable stream
// was closed; the file is read and written through it from here on.
//
// A failure leaves the reservation to [BrowserUpload.Discard]: nothing partly
// written should reach a page as a file it can read.
func (u *BrowserUpload) Take(name string, handle js.Value) (*File, error) {
	size, err := u.body.adopt(handle)
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	file := &File{name: name, body: u.body}
	file.setSize(size)

	return file, nil
}

// Discard drops the reservation and whatever the page managed to write. An
// upload that failed halfway -- out of quota, cancelled, or a worker that went
// away -- leaves nothing behind for the origin's quota to carry.
func (u *BrowserUpload) Discard() {
	u.body.remove()
}

// setSize records how much the file holds. A browser-written file is never
// written through [File.write], which is what keeps the count everywhere else,
// so the size is read off the handle when the file is taken over.
func (f *File) setSize(n int64) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.size = n
}
