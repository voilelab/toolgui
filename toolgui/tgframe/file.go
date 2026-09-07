package tgframe

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// FileReader is how an uploaded file is read. It's a stream, and it also
// reads at an offset, which is what archive/zip and the image decoders want.
type FileReader interface {
	io.ReadSeekCloser
	io.ReaderAt
}

// File is a file the user uploaded. Its content lives on disk, so a page can
// take a file larger than the memory the process has to spare.
type File struct {
	name string
	path string

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
	fp, err := os.Open(f.path)
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	return fp, nil
}

// Bytes read the whole content into memory. Prefer [File.Open] for anything
// that can work on a stream.
func (f *File) Bytes() ([]byte, error) {
	bs, err := os.ReadFile(f.path)
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

// write copies r into the file, truncating it first unless atEnd is set.
func (f *File) write(r io.Reader, atEnd bool) error {
	flag := os.O_CREATE | os.O_WRONLY
	if atEnd {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}

	fp, err := os.OpenFile(f.path, flag, 0o600)
	if err != nil {
		return tgutil.Errorf("%w", err)
	}

	// io.Copy works through a fixed buffer, so what's held is the buffer and
	// not the upload.
	n, err := io.Copy(fp, r)

	closeErr := fp.Close()
	if err == nil {
		err = closeErr
	}

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

// remove drops the content from disk. The File is unusable afterwards.
func (f *File) remove() {
	if err := os.Remove(f.path); err != nil && !os.IsNotExist(err) {
		slog.Error("remove file", "path", f.path, "error", err)
	}
}

// fileStore is where a state's uploads live: a directory of its own, and the
// file each key currently holds. A cloned state shares one of these, so the
// two can't hand out the same path.
type fileStore struct {
	lock  sync.Mutex
	dir   string
	seq   int
	files map[string]*File
}

func newFileStore() *fileStore {
	return &fileStore{files: make(map[string]*File)}
}

// newFile makes an empty file in the store's directory. It belongs to no key
// until [fileStore.put] takes it, so an upload in progress can't be read as
// the page's current file.
func (s *fileStore) newFile(name string) (*File, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.dir == "" {
		dir, err := os.MkdirTemp("", "toolgui-state-")
		if err != nil {
			return nil, tgutil.Errorf("%w", err)
		}

		s.dir = dir
	}

	s.seq++

	// The file is named after the counter: naming it after the upload would
	// mean trusting a name the browser chose.
	file := &File{name: name, path: filepath.Join(s.dir, fmt.Sprint(s.seq))}

	fp, err := os.OpenFile(file.path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	if err := fp.Close(); err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	return file, nil
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
		old.remove()
	}
}

func (s *fileStore) get(key string) *File {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.files[key]
}

// destroy removes every file the store holds, the directory included.
func (s *fileStore) destroy() {
	s.lock.Lock()
	dir := s.dir
	s.dir = ""
	s.files = make(map[string]*File)
	s.lock.Unlock()

	if dir == "" {
		return
	}

	if err := os.RemoveAll(dir); err != nil {
		slog.Error("remove state files", "dir", dir, "error", err)
	}
}
