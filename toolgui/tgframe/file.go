package tgframe

import (
	"io"
	"os"
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
