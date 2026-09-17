# Download File

DownloadFile create a button that hands the app user a file to download,
fetched by token rather than carried in the page.

## API

### Interface

```go
func DownloadFile(c *tgframe.Container, text string, body []byte, conf ...*DownloadFileConf) bool
```

### Parameters

* `c` is Parent container.
* `text` is the text on button.
* `body` is the bytes of file.
* `conf` is an optional configuration, at most one.
  The file name is set by `conf.Filename`.

```go
type DownloadFileConf struct {
	tgframe.Base // ID

	// MIME specifies the Multipurpose Internet Mail Extension (MIME) type of the downloaded content.
	// Defaults to "application/octet-stream" if not provided.
	MIME string

	// Color defines the color of the download button.
	Color tcutil.Color

	// Disabled indicates whether the download button should be initially disabled.
	Disabled bool

	// Filename sets the suggested filename for the downloaded content when clicked.
	// Defaults to the content's MD5 in hex.
	Filename string
}
```

It returns whether this run is handling a click on it, the same as
[`DownloadButton`](download_button.md).

## Example

```go
{{#include ../../../demos/download_file.go:demo}}
```

<div data-toolgui-demo="download_file" data-toolgui-demo-height="640"></div>

## `DownloadFile` or `DownloadButton`

Both draw the same button. What differs is where the bytes travel.

`DownloadButton` puts the content in the component itself, as a
`data:` URI. The content is base64, which is a third larger again than the
bytes, and it travels in the render tree: over the update socket on a server,
through the worker's message in the browser, and it sits in the tab as a string
for as long as the button is on screen. A rerun that draws the same button
sends it all again.

`DownloadFile` stores the bytes where the build keeps files — on disk on a
server, in the origin private file system in a browser build — and puts only an
unguessable token in the component. A click fetches the file through the state's
own transport, straight from that file into a blob the browser keeps wherever it
keeps blobs, and saves it. Nothing about the content is in the render tree and
nothing of it is held in the tab's heap, so the file costs what it weighs
whether it is a megabyte or a gigabyte.

Roughly:

| | `DownloadButton` | `DownloadFile` |
| --- | --- | --- |
| Where the content is | in the component, as base64 | in the state's file storage |
| What the pack carries | the whole file, ×1.33 | a token |
| Fetch on click | none, the URI is already there | one, through the state's transport |
| Sensible up to | a few hundred kilobytes | whatever the disk or the origin's quota holds |

So: below a few hundred kilobytes, `DownloadButton` — the round trip saved is
worth more than the transport it costs. Above that, `DownloadFile`. A megabyte
is where it starts to matter; anything the page builds from a query, an export
or an archive belongs here from the start.

## Which builds it works in

| Executor | `DownloadButton` | `DownloadFile` |
| --- | --- | --- |
| Web (`tgexec.WebExecutor`) | yes | yes |
| Browser (`tgwasm`) | yes | yes |
| Desktop (`tgwails`) | yes | **no** |

The desktop build has neither an HTTP endpoint nor an origin private file
system, and how a Wails binding should hand a file of any size to the webview
is not settled. Until it is, a `DownloadFile` in a desktop app draws its button
and fails to save, with the reason on the console: use `DownloadButton` there,
or keep the component out of a page a desktop build serves.

## The token

The token names one file and grants nothing by itself. It is unguessable, it is
looked up in the state that offered it and nowhere else, and the fetch carries
it with whatever already identifies the connection — the state id on a server,
the session the bridge holds in a browser build. So a token taken from one page
fetches nothing on another, and a token alone fetches nothing at all.

It is also the run's, not the component's. A rerun that offers the same file
again keeps the token, so the pack does not change and a client holding one goes
on using it; a rerun that offers different bytes writes a new file under a new
token. What a token names is the output of the run that handed it out.

The token from the run before stays fetchable, because that is the button the
app user still has on screen until the replacement reaches them: a click in
that window saves what it was offering rather than failing. One run back and no
further, so a page that offers a new file on every run holds two of them per
button and not a history.

A download lives as long as the component offering it: clear it off the screen
and its bytes and its token go with it, and a state that goes away takes
whatever is left.
