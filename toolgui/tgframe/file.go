package tgframe

import (
	"io"
	"sync"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// FileReader is how an uploaded file is read. It's a stream, and it also
// reads at an offset, which is what archive/zip and the image decoders want.
type FileReader interface {
	io.ReadSeekCloser
	io.ReaderAt
}

// fileBody is where one file's bytes are kept.
type fileBody interface {
	open() (FileReader, error)

	// write copies r in, onto the end when atEnd is set and over what is
	// there otherwise, and returns how much it wrote.
	write(r io.Reader, atEnd bool) (int64, error)

	remove()
}

// fileBodies makes the bodies of one state's files and clears up after them.
// Which one a build gets is what decides where an upload is kept: on disk on
// a server, in memory in the browser, which has no filesystem to put it on.
// Its methods are called under the store's lock, so they need none of their
// own.
type fileBodies interface {
	newBody() (fileBody, error)
	destroy()
}

// File is a file the user uploaded. Where its content lives is the build's
// business: on a server it's on disk, so a page can take a file larger than
// the memory the process has to spare.
type File struct {
	name string
	body fileBody

	lock sync.RWMutex
	size int64
}

// Name return the name the file was uploaded under.
func (f *File) Name() string {
	return f.name
}

// Size return the number of bytes stored so far.
func (f *File) Size() int64 {
	f.lock.RLock()
	defer f.lock.RUnlock()

	return f.size
}

// Open return a reader over the content. The caller closes it.
func (f *File) Open() (FileReader, error) {
	fp, err := f.body.open()
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	return fp, nil
}

// Bytes read the whole content into memory. Prefer [File.Open] for anything
// that can work on a stream.
func (f *File) Bytes() ([]byte, error) {
	fp, err := f.Open()
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}
	defer fp.Close()

	bs, err := io.ReadAll(fp)
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	return bs, nil
}

// Append copies what r yields onto the end of the file. It's what lets a
// transport that can only carry a chunk at a time build a file up.
func (f *File) Append(r io.Reader) error {
	return f.write(r, true)
}

func (f *File) write(r io.Reader, atEnd bool) error {
	n, err := f.body.write(r, atEnd)

	f.lock.Lock()
	if atEnd {
		f.size += n
	} else {
		f.size = n
	}
	f.lock.Unlock()

	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	return nil
}

// fileStore is a state's uploads: the file each key currently holds, and
// wherever their bytes are kept. A cloned state shares one of these, so the
// two can't hand out the same place to write.
type fileStore struct {
	lock   sync.Mutex
	bodies fileBodies
	files  map[string]*File
}

func newFileStore() *fileStore {
	return &fileStore{bodies: newFileBodies(), files: make(map[string]*File)}
}

// newFile makes an empty file. It belongs to no key until [fileStore.put]
// takes it, so an upload in progress can't be read as the page's current file.
func (s *fileStore) newFile(name string) (*File, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	body, err := s.bodies.newBody()
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	return &File{name: name, body: body}, nil
}

// put stores file under key and drops what the key held. The old file is
// unreachable once replaced, and a session that uploads all day shouldn't
// fill the disk with them.
func (s *fileStore) put(key string, file *File) {
	s.lock.Lock()
	old := s.files[key]
	s.files[key] = file
	s.lock.Unlock()

	if old != nil && old != file {
		old.body.remove()
	}
}

// remove drops the file under key, and its bytes with it.
func (s *fileStore) remove(key string) {
	s.lock.Lock()
	old := s.files[key]
	delete(s.files, key)
	s.lock.Unlock()

	if old != nil {
		old.body.remove()
	}
}

func (s *fileStore) get(key string) *File {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.files[key]
}

// destroy drops every file the store holds.
func (s *fileStore) destroy() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.files = make(map[string]*File)
	s.bodies.destroy()
}
