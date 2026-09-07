# Tab

Tab component is used to create a tabbed interface.

## API

```go
func Tab(c *tgframe.Container, tabs []string, conf ...*TabConf) []*tgframe.Container
func Tab2(c *tgframe.Container, tab1, tab2 string, conf ...*TabConf) (*tgframe.Container, *tgframe.Container)
func Tab3(c *tgframe.Container, tab1, tab2, tab3 string, conf ...*TabConf) (*tgframe.Container, *tgframe.Container, *tgframe.Container)
func Tab4(c *tgframe.Container, tab1, tab2, tab3, tab4 string, conf ...*TabConf) (*tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container)
func Tab5(c *tgframe.Container, tab1, tab2, tab3, tab4, tab5 string, conf ...*TabConf) (*tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container)
```

* `c` is the container to add the tab component to.
* `tabs` is the list of tab titles.
* `tab1`, `tab2`, `tab3`, `tab4`, `tab5` are the tab titles.
* `conf` is an optional configuration, at most one.

```go
// TabConf is the configuration for the tab components.
type TabConf struct {
	tgframe.Base // ID
}
```

The client keeps which tab is open, so the component claims an id derived from
the tab titles. Two tab groups with the same titles collide; give one of them
a `Conf.ID`.

## Example

```go
tabs := tgcomp.Tab(c, []string{"Tab 1", "Tab 2", "Tab 3"})
for _, tab := range tabs {
    tgcomp.Text(tab, "Hello World")
}
```

```go
one, two := tgcomp.Tab2(c, "one", "two", &tgcomp.TabConf{ID: "lower"})
```

## Rendering

Every tab is rendered and the inactive ones are hidden, so a tab keeps what it
holds while you are on another one: an `Iframe` is not reloaded, a `Chart` is
not rebuilt, and what was typed into an input is still there when you come back.
