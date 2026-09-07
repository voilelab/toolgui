//go:build !js || !wasm

package tgframe

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// diskBodies keeps a state's uploads in a directory of its own. It's made on
// the first upload, so a state that never sees one leaves nothing behind.
type diskBodies struct {
	dir string
	seq int
}

func newFileBodies() fileBodies {
	return &diskBodies{}
}

func (d *diskBodies) newBody() (fileBody, error) {
	if d.dir == "" {
		dir, err := os.MkdirTemp("", "toolgui-state-")
		if err != nil {
			return nil, tgutil.Errorf("%w", err)
		}

		d.dir = dir
	}

	d.seq++

	// The file is named after the counter: naming it after the upload would
	// mean trusting a name the browser chose.
	body := &diskBody{path: filepath.Join(d.dir, fmt.Sprint(d.seq))}

	fp, err := os.OpenFile(body.path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	if err := fp.Close(); err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	return body, nil
}

func (d *diskBodies) destroy() {
	dir := d.dir
	d.dir = ""

	if dir == "" {
		return
	}

	if err := os.RemoveAll(dir); err != nil {
		slog.Error("remove state files", "dir", dir, "error", err)
	}
}

type diskBody struct {
	path string
}

func (b *diskBody) open() (FileReader, error) {
	fp, err := os.Open(b.path)
	if err != nil {
		return nil, tgutil.Errorf("%w", err)
	}

	return fp, nil
}

func (b *diskBody) write(r io.Reader, atEnd bool) (int64, error) {
	flag := os.O_CREATE | os.O_WRONLY
	if atEnd {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}

	fp, err := os.OpenFile(b.path, flag, 0o600)
	if err != nil {
		return 0, tgutil.Errorf("%w", err)
	}

	// io.Copy works through a fixed buffer, so what's held is the buffer and
	// not the upload.
	n, err := io.Copy(fp, r)

	closeErr := fp.Close()
	if err == nil {
		err = closeErr
	}

	if err != nil {
		return n, tgutil.Errorf("%w", err)
	}

	return n, nil
}

func (b *diskBody) remove() {
	if err := os.Remove(b.path); err != nil && !os.IsNotExist(err) {
		slog.Error("remove file", "path", b.path, "error", err)
	}
}
