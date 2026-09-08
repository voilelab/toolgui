# Table

Table component display a table. It is static: the rows are drawn in the order
they are given, and there is nothing for the reader to click.

Reach for it when the rows are few and already in the order they should be
read in. When the rows are many, or when the reader — not the page function —
should decide the order, use [DataFrame](dataframe.md), which sorts, searches
and pages in the browser.

## API

```go
func Table(c *tgframe.Container, head []string, table [][]string, conf ...*TableConf)
```

* `c` is Parent container.
* `head` is the head of table.
* `table` is the body of table.
* `conf` is an optional configuration, at most one.

```go
// TableConf is the configuration for the Table component.
type TableConf struct {
	tgframe.Base // ID
}
```

## Example

```go
tgcomp.Table(p.Main, []string{"a", "b"}, [][]string{{"1", "2"}, {"3", "4"}})
```

![table component](table.png)
