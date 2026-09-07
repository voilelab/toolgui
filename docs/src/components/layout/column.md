# Column

Column provides columns layout.

## Usage

- Column create N columns, each as wide as its content needs.
- Column1 create 1 column.
- Column2 create 2 columns.
- Column3 create 3 columns.
- EqColumn create N columns with equal width.
- EqColumn1 create 1 column with equal width.
- EqColumn2 create 2 columns with equal width.
- EqColumn3 create 3 columns with equal width.
- EqColumn4 create 4 columns with equal width.
- EqColumn5 create 5 columns with equal width.

```go
func Column(c *tgframe.Container, n uint, conf ...*ColumnConf) []*tgframe.Container
func Column1(c *tgframe.Container, conf ...*ColumnConf) *tgframe.Container
func Column2(c *tgframe.Container, conf ...*ColumnConf) (*tgframe.Container, *tgframe.Container)
func Column3(c *tgframe.Container, conf ...*ColumnConf) (*tgframe.Container, *tgframe.Container, *tgframe.Container)

func EqColumn(c *tgframe.Container, n uint, conf ...*ColumnConf) []*tgframe.Container
func EqColumn1(c *tgframe.Container, conf ...*ColumnConf) *tgframe.Container
func EqColumn2(c *tgframe.Container, conf ...*ColumnConf) (*tgframe.Container, *tgframe.Container)
func EqColumn3(c *tgframe.Container, conf ...*ColumnConf) (*tgframe.Container, *tgframe.Container, *tgframe.Container)
func EqColumn4(c *tgframe.Container, conf ...*ColumnConf) (*tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container)
func EqColumn5(c *tgframe.Container, conf ...*ColumnConf) (*tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container)
```

* `c`: Parent container.
* `n`: Number of column.
* `conf`: Optional configuration, at most one.

```go
// ColumnConf is the configuration for the column components.
type ColumnConf struct {
	tgframe.Base // ID
}
```

The containers a column hands out derive their ids from its ID; give none and
they carry none, and the components inside are still placed by position.

## Example

```go
cols := tgcomp.Column(colCompCol, 3)
for i, col := range cols {
	tgcomp.Text(col, fmt.Sprintf("col-%d", i))
}
```

```go
left, right := tgcomp.EqColumn2(p.Main, &tgcomp.ColumnConf{ID: "summary"})
```
