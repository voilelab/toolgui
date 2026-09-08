# Fileupload

Fileupload create a fileupload and return its selected file.

## API

```go
type FileObject struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size int    `json:"size"`
}

func (f *FileObject) Open() (tgframe.FileReader, error)
func (f *FileObject) Bytes() ([]byte, error)

func Fileupload(c *tgframe.Container, label, accept string, conf ...*FileuploadConf) *FileObject
```

* `c` is Parent container.
* `label` is the label for options group.
* `accept` is the file type to accept.
* `conf` is an optional configuration, at most one.
* Return the selected file object. nil if no file is selected.

```go
// FileuploadConf is the configuration for the Fileupload component.
type FileuploadConf struct {
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
file only has to fit on disk. In the browser, where a WebAssembly app has no
filesystem to use, it stays in the tab. `Size` is the size of what was stored.

* `Open` returns a reader over the content, which the caller closes. It reads
  at an offset too, so `archive/zip` and the image decoders can work straight
  off it.
* `Bytes` reads the whole file into memory. Prefer `Open` for anything that
  can work on a stream.

## Example

```go
fileObj := tgcomp.Fileupload(p.Main, "Fileupload", ".jpg,.png")
if fileObj != nil {
    tgcomp.Text(p.Main, "Fileupload filename: "+fileObj.Name)
    tgcomp.Text(p.Main, fmt.Sprintf("Fileupload bytes length: %d", fileObj.Size))
}
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
