# Download Button

DownloadButton create a download button component.

## API

### Interface

```go
func DownloadButton(c *tgframe.Container, text string, body []byte, conf ...*DownloadButtonConf) bool
```

### Parameters

* `c` is Parent container.
* `text` is the text on button.
* `body` is the bytes of file.
* `conf` is an optional configuration, at most one.
  The file name is set by `conf.Filename`.

```go
type DownloadButtonConf struct {
	tgframe.Base // ID

	// MIME specifies the Multipurpose Internet Mail Extension (MIME) type of the downloaded content.
	// Defaults to "application/octet-stream" if not provided.
	MIME string

	// Color defines the color of the download button.
	Color tcutil.Color

	// Disabled indicates whether the download button should be initially disabled.
	Disabled bool

	// Filename sets the suggested filename for the downloaded content when clicked.
	Filename string
}
```

## Size

The content travels in the component, as a `data:` URI: base64, so a third
larger again than the bytes, and it is sent on every run that draws the button
and held in the tab as a string for as long as it is on screen. That is the
right trade below a few hundred kilobytes, where the round trip it saves is
worth more than the transport it costs.

Above that, use [`DownloadFile`](download_file.md), which keeps the bytes in the
state's file storage and puts a token in the component instead. Its page has the
comparison in full.

## Example

```go
tgcomp.DownloadButton(p.Main, "Download", []byte("123"),
    &tgcomp.DownloadButtonConf{Filename: "123.txt"})
```

![download button component](download_button.png)

## Asking before the button is drawn

```go
func DownloadButtonClicked(c *tgframe.Container, text string,
    conf ...*DownloadButtonConf) bool
```

[`ButtonClicked`](button.md#asking-before-the-button-is-drawn) for a download
button: it reports the same click, read off the run's state rather than the
component, so a page can ask before the button is written. Give it the same
`text` and `conf` the `DownloadButton` call gets.

It takes no `body`: the button's id comes from `text` (or the conf id), not
from what it hands over, so a page can ask about the click before it has the
bytes to offer.
