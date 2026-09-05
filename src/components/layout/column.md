# Column

Column provides columns layout.

## Usage

- Column create N columns.
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
func Column(c *tgframe.Container, id string, n uint) []*tgframe.Container
func Column1(c *tgframe.Container, id string) *tgframe.Container
func Column2(c *tgframe.Container, id string) (*tgframe.Container, *tgframe.Container)
func Column3(c *tgframe.Container, id string) (*tgframe.Container, *tgframe.Container, *tgframe.Container)

func EqColumn(c *tgframe.Container, id string, n uint) []*tgframe.Container
func EqColumn1(c *tgframe.Container, id string) *tgframe.Container
func EqColumn2(c *tgframe.Container, id string) (*tgframe.Container, *tgframe.Container)
func EqColumn3(c *tgframe.Container, id string) (*tgframe.Container, *tgframe.Container, *tgframe.Container)
func EqColumn4(c *tgframe.Container, id string) (*tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container)
func EqColumn5(c *tgframe.Container, id string) (*tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container, *tgframe.Container)
```

* `c`: Parent container.
* `id`: Unique component ID.
* `n`: Number of column.

## Example

```go
cols := tgcomp.Column(colCompCol, "cols", 3)
for i, col := range cols {
	tgcomp.Text(col, fmt.Sprintf("col-%d", i))
}
```
