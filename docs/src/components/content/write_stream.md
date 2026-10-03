# Write Stream

WriteStream writes a stream of text, such as a model's reply, into one
[Markdown](markdown.md) as it arrives, and returns the whole text once the
stream ends.

## API

### Interface

```go
func WriteStream(c *tgframe.Container, seq iter.Seq2[string, error],
	conf ...*WriteStreamConf) (string, error)
```

### Parameters

* `c`: Parent container.
* `seq`: The chunks to write. The first error stops the stream; it is returned
  with the text before it.
* `conf`: Optional configuration, at most one.

```go
// WriteStreamConf is the configuration for the WriteStream component.
type WriteStreamConf struct {
	tgframe.Base // ID

	// Interval is the least time between two sends of the text. Default is
	// 50ms.
	Interval time.Duration
}
```

The growing text is sent at most once per `Interval`, so a stream of one token
per chunk does not resend the whole text once per token.

### The stream

`seq` runs on a goroutine of its own, and must not draw components.
WriteStream waits for it to return before it does, so build it on
[`Params.Context`](../../app/page.md): a run cut by the next event then ends
the request too, instead of holding up the run after it.

### Keeping the text

Like everything else on the page, the streamed text is drawn again only if the
next run writes it again. Keep what WriteStream returns in the
[state](../../architecture/state-storage.md) and draw it with `Markdown` from
then on, as a chat keeps its history.

## Example

```go
{{#include ../../../demos/write_stream.go:demo}}
```

```go
{{#include ../../../demos/write_stream.go:reply}}
```

<div data-toolgui-demo="write_stream" data-toolgui-demo-height="300"></div>
