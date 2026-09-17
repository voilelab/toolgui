# FileUpload

FileUpload create a fileupload and return its selected file.

## API

```go
type FileObject struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size int    `json:"size"`
}

func (f *FileObject) Open() (tgframe.FileReader, error)
func (f *FileObject) Bytes() ([]byte, error)

func FileUpload(c *tgframe.Container, label, accept string, conf ...*FileUploadConf) *FileObject
```

* `c` is Parent container.
* `label` is the label for options group.
* `accept` is the file type to accept.
* `conf` is an optional configuration, at most one.
* Return the selected file object. nil if no file is selected.

```go
// FileUploadConf is the configuration for the FileUpload component.
type FileUploadConf struct {
	tgframe.Base // ID

	// Disabled is true if the fileupload is disabled.
	Disabled bool
}
```

There is no `Default` here, unlike the other inputs. A file input is the one
control a page cannot fill in on the app user's behalf: the browser refuses to
have its value set from script, so a default would read back in Go while the
box on screen stayed empty. A page that wants to start from a file it already
has should read that file itself rather than ask for one.

On a server the upload is streamed to disk rather than kept in memory, so a
file only has to fit on disk. In the browser it is kept in the origin private
file system, which is what a tab has instead of one, so a stored file is not a
second copy sitting in the tab's memory for as long as the session lasts. It
does still cross in one piece on the way in, so an upload has to fit in the tab
to arrive. `Size` is the size of what was stored.

* `Open` returns a reader over the content, which the caller closes. It reads
  at an offset too, so `archive/zip` and the image decoders can work straight
  off it.
* `Bytes` reads the whole file into memory. Prefer `Open` for anything that
  can work on a stream.

## Example

```go
{{#include ../../../demos/fileupload.go:demo}}
```

Reading the content through a stream:

```go
fp, err := fileObj.Open()
if err != nil {
    return err
}
defer fp.Close()

img, err := jpeg.Decode(fp)
```

![fileupload component](fileupload.png)
