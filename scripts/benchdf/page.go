// Command benchdf measures what a DataFrame costs on a rerun that leaves its
// rows unchanged (TG-94). Each page holds a Toggle and a DataFrame of the size
// in its name; flipping the Toggle reruns the page without touching the table.
package main

import (
	"fmt"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// rowCounts are the table sizes measured, 0 being the page without a table.
var rowCounts = []int{0, 1000, 3000, 10000}

func pageName(n int) string { return fmt.Sprintf("rows%d", n) }

func newApp() *tgframe.App {
	app := tgframe.NewApp()
	app.SetTitle("DataFrame bench")
	for _, n := range rowCounts {
		app.AddPage(pageName(n), fmt.Sprintf("%d rows", n), benchPage(n))
	}
	return app
}

// sampleRows is a sample table shaped like a real one: ids, a group, a
// timestamp, numbers and free text.
func sampleRows(n int) ([][]tgcomp.Cell, []string) {
	groups := []string{"alpha", "beta", "gamma", "delta", "epsilon"}
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	rows := make([][]tgcomp.Cell, n)
	keys := make([]string, n)
	for i := range n {
		keys[i] = fmt.Sprintf("sample-%06d", i)
		rows[i] = []tgcomp.Cell{
			tgcomp.TextCell(keys[i]),
			tgcomp.TextCell(groups[i%len(groups)]),
			tgcomp.TimeCell(base.Add(time.Duration(i) * time.Minute)),
			tgcomp.NumberCell(float64(i%997) * 1.25),
			tgcomp.NumberCell(float64(i%200) / 200).WithDisplay(
				fmt.Sprintf("%.1f%%", float64(i%200)/2)),
			tgcomp.TextCell(fmt.Sprintf("note for sample %d", i)),
		}
	}
	return rows, keys
}

var head = []string{"ID", "Group", "Time", "Value", "Ratio", "Note"}

func benchPage(n int) tgframe.RunFunc {
	rows, keys := sampleRows(n)
	return func(p *tgframe.Params) error {
		on := tgcomp.Toggle(p.Main, "Flag")
		tgcomp.Text(p.Main, fmt.Sprintf("flag: %v", on))

		if n == 0 {
			return nil
		}

		tgcomp.DataFrameCells(p.Main, head, rows, &tgcomp.DataFrameConf{
			Selection: tgcomp.SelectionModeMulti,
			RowKeys:   keys,
			ColumnConf: []tgcomp.DataFrameColumnConf{
				{}, {},
				{Type: tgcomp.ColumnTypeDatetime},
				{Type: tgcomp.ColumnTypeNumber},
				{Type: tgcomp.ColumnTypeNumber},
				{},
			},
		})
		return nil
	}
}
