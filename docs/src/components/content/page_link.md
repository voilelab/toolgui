# Page Link

Page Link component displays a link to another page of the app, opened with
a page query. The page it opens reads the query as
[`Params.Query`](../../app/page.md#page-query).

It is a real link: it opens in a new tab on a middle click, and its url can be
copied. It takes a page name and values, never a url, so it cannot point off
the app. The url follows the app's page mode: `/detail?group=a`, or
`#/detail?group=a` in hash mode and in the browser build. On the desktop it
switches page the way the side nav does.

The text supports [emoji shortcodes](emoji.md).

## API

```go
func PageLink(c *tgframe.Container, text, page string,
	query url.Values, conf ...*PageLinkConf)
```

* `c` is Parent container.
* `text` is the link text.
* `page` is the name of the page to open. A page the app does not have fails
  the component: the run reports the error and the link is not drawn.
* `query` is the page query. `nil` opens the page without one.
* `conf` is an optional configuration, at most one.

```go
// PageLinkConf is the configuration for the PageLink component.
type PageLinkConf struct {
	tgframe.Base // ID
}
```

## Example

```go
{{#include ../../../demos/page_link.go:demo}}
```

<div data-toolgui-demo="page_link"></div>
