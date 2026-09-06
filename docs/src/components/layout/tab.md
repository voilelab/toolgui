# Tab

Tab component is used to create a tabbed interface.

## API

```go
func Tab(c *tgframe.Container, tabs []string) []*tgframe.Container
func Tab2(c *tgframe.Container, tab1, tab2 string) (*tgframe.Container, *tgframe.Container)
func Tab3(c *tgframe.Container, tab1, tab2, tab3 string) (*tgframe.Container, *tgframe.Container, *tgframe.Container)
func Tab4(c *tgframe.Container, tab1, tab2, tab3, tab4 string) (*tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container)
func Tab5(c *tgframe.Container, tab1, tab2, tab3, tab4, tab5 string) (*tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container)
```

* `c` is the container to add the tab component to.
* `tabs` is the list of tab titles.
* `tab1`, `tab2`, `tab3`, `tab4`, `tab5` are the tab titles.

## Example

```go
tabs := tgcomp.Tab(c, []string{"Tab 1", "Tab 2", "Tab 3"})
for _, tab := range tabs {
    tgcomp.Text(tab, "Hello World")
}
```
