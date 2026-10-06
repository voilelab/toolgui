package tcdata

import (
	"math"
	"strconv"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

type cellKind int

const (
	cellText cellKind = iota
	cellNumber
	cellTime
	cellMissing
)

// Cell is one DataFrame cell: a value the column sorts by, and the string
// the table shows and searches. Build one with [TextCell], [NumberCell],
// [TimeCell] or [MissingCell].
type Cell struct {
	kind       cellKind
	text       string
	num        float64
	display    string
	hasDisplay bool
}

// TextCell is a cell holding s, read the way [DataFrame] reads its strings:
// a number or datetime column still parses it.
func TextCell(s string) Cell {
	return Cell{kind: cellText, text: s}
}

// NumberCell is a cell sorted by v. It is shown through the column's
// [NumberFormat] unless WithDisplay says otherwise. NaN is a missing cell.
// It belongs in a ColumnTypeNumber column.
func NumberCell(v float64) Cell {
	if math.IsNaN(v) {
		return MissingCell()
	}
	return Cell{kind: cellNumber, num: v}
}

// TimeCell is a cell sorted by the instant t, shown as RFC 3339 unless
// WithDisplay says otherwise. It belongs in a ColumnTypeDatetime column.
func TimeCell(t time.Time) Cell {
	return Cell{kind: cellTime, num: float64(t.UnixMilli()),
		text: t.Format(time.RFC3339)}
}

// MissingCell is a cell with no value. It sorts last whichever direction the
// column is sorted in, and shows empty unless WithDisplay says otherwise.
func MissingCell() Cell {
	return Cell{kind: cellMissing}
}

// WithDisplay returns the cell shown (and searched) as s, sorting as before.
func (c Cell) WithDisplay(s string) Cell {
	c.display = s
	c.hasDisplay = true
	return c
}

// NumberFormat is how a ColumnTypeNumber column shows a [NumberCell] that
// has no display of its own.
type NumberFormat struct {
	// Decimals is how many digits follow the point. Negative fails the run.
	Decimals int

	// Percent shows the value times 100 followed by "%": 0.2941 with
	// Decimals 2 shows as "29.41%".
	Percent bool
}

func (f *NumberFormat) format(v float64) string {
	if f == nil {
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	if f.Percent {
		return strconv.FormatFloat(v*100, 'f', f.Decimals, 64) + "%"
	}
	return strconv.FormatFloat(v, 'f', f.Decimals, 64)
}

// dataFrameCell is a typed cell on the wire. A plain text cell goes as a
// bare string instead, so a string-only DataFrame's wire is unchanged.
// Value is a float64 (a number, or epoch milliseconds for a time), the
// string of a text cell, or nil for a missing cell.
type dataFrameCell struct {
	Display string `json:"display"`
	Value   any    `json:"value"`
}

// wire settles the cell into what the client receives: a string or a
// dataFrameCell.
func (c Cell) wire(format *NumberFormat) any {
	var value any
	var display string

	switch c.kind {
	case cellText:
		if !c.hasDisplay {
			return c.text
		}
		value = c.text
	case cellNumber:
		value = c.num
		display = format.format(c.num)
	case cellTime:
		value = c.num
		display = c.text
	case cellMissing:
		value = nil
	}

	if c.hasDisplay {
		display = c.display
	}

	return dataFrameCell{Display: display, Value: value}
}

// check reports a cell its column cannot sort, which fails the run.
func (c Cell) check(row int, head string, t ColumnType) error {
	switch c.kind {
	case cellNumber:
		if math.IsInf(c.num, 0) {
			return tgutil.Errorf(
				"row %d column %q holds an infinite number", row, head)
		}
		if t != ColumnTypeNumber {
			return tgutil.Errorf(
				"row %d column %q holds a NumberCell, but the column is %s",
				row, head, t)
		}
	case cellTime:
		if t != ColumnTypeDatetime {
			return tgutil.Errorf(
				"row %d column %q holds a TimeCell, but the column is %s",
				row, head, t)
		}
	}

	return nil
}

// dataFrameRows is the rows either entry point takes.
type dataFrameRows interface {
	len() int
	check(head []string, cf *DataFrameConf) error
	wire(cf *DataFrameConf) any
}

// stringRows go on the wire as given, with no per-cell copy: the large
// string tables DataFrame is for pay nothing for typed cells.
type stringRows [][]string

func (r stringRows) len() int { return len(r) }

func (r stringRows) check(head []string, _ *DataFrameConf) error {
	for i, row := range r {
		if len(row) != len(head) {
			return tgutil.Errorf("len of row %d should equal to len of head", i)
		}
	}
	return nil
}

// wire is never nil, so the client can iterate it unguarded.
func (r stringRows) wire(*DataFrameConf) any {
	if r == nil {
		return [][]string{}
	}
	return [][]string(r)
}

type cellRows [][]Cell

func (r cellRows) len() int { return len(r) }

func (r cellRows) check(head []string, cf *DataFrameConf) error {
	for i, row := range r {
		if len(row) != len(head) {
			return tgutil.Errorf("len of row %d should equal to len of head", i)
		}

		for j, cell := range row {
			if err := cell.check(i, head[j], cf.columnConf(j).Type); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r cellRows) wire(cf *DataFrameConf) any {
	out := make([][]any, len(r))
	for i, row := range r {
		cells := make([]any, len(row))
		for j, cell := range row {
			cells[j] = cell.wire(cf.columnConf(j).Format)
		}
		out[i] = cells
	}
	return out
}
