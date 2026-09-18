package demos

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// demoOrders is the fake order book the DataFrame demo pages through. It is
// generated rather than written out so that there is enough of it to sort,
// search and page.
func demoOrders() [][]string {
	regions := []string{"APAC", "EMEA", "LATAM", "NA"}
	items := []string{"Keyboard", "Monitor", "Mouse", "Laptop", "Dock"}

	const count = 2000
	ordered := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)

	rows := make([][]string, 0, count)
	for i := range count {
		rows = append(rows, []string{
			fmt.Sprintf("ORD-%04d", i+1),
			ordered.AddDate(0, 0, i%365).Format(time.RFC3339),
			regions[i%len(regions)],
			items[i%len(items)],
			strconv.Itoa(i + 1),
		})
	}

	return rows
}

func dataFrameDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	tgcomp.DataFrame(p.Main,
		[]string{"Order", "Ordered", "Region", "Item", "Amount"},
		demoOrders(),
		&tgcomp.DataFrameConf{
			ID:       "demo_orders",
			PageSize: 10,
			ColumnConf: []tgcomp.DataFrameColumnConf{
				{Width: "9rem"},
				{Type: tgcomp.ColumnTypeDatetime},
				{},
				{},
				{Type: tgcomp.ColumnTypeNumber},
			},
		})
	// ANCHOR_END: demo
	return nil
}

func dataFrameMultiDemo(p *tgframe.Params) error {
	// ANCHOR: multi
	hosts := [][]string{
		{"web-1", "APAC", "healthy"},
		{"web-2", "EMEA", "degraded"},
		{"db-1", "NA", "healthy"},
		{"db-2", "LATAM", "down"},
	}

	// The host name is what a row is, so the pick follows the host rather
	// than the position it happens to sit at this run.
	keys := make([]string, 0, len(hosts))
	for _, host := range hosts {
		keys = append(keys, host[0])
	}

	selected := tgcomp.DataFrame(p.Main,
		[]string{"Host", "Region", "Status"}, hosts,
		&tgcomp.DataFrameConf{
			ID:        "demo_hosts",
			Selection: tgcomp.SelectionModeMulti,
			RowKeys:   keys,
		})

	names := []string{}
	for _, idx := range selected {
		names = append(names, hosts[idx][0])
	}

	picked := "none"
	if len(names) != 0 {
		picked = strings.Join(names, ", ")
	}

	tgcomp.Text(p.Main, "Selected: "+picked,
		&tgcomp.TextConf{ID: "dataframe_multi_result"})
	// ANCHOR_END: multi
	return nil
}

func dataFrameSingleDemo(p *tgframe.Params) error {
	// ANCHOR: single
	builds := [][]string{
		{"#41", "Go", "passed"},
		{"#42", "Rust", "failed"},
		{"#43", "Python", "passed"},
	}

	selected := tgcomp.DataFrame(p.Main,
		[]string{"Build", "Language", "Result"}, builds,
		&tgcomp.DataFrameConf{
			ID:               "demo_builds",
			Selection:        tgcomp.SelectionModeSingle,
			DefaultSelection: []int{0},
		})

	detail := "none"
	if len(selected) != 0 {
		detail = strings.Join(builds[selected[0]], " / ")
	}

	tgcomp.Text(p.Main, "Build: "+detail,
		&tgcomp.TextConf{ID: "dataframe_single_result"})
	// ANCHOR_END: single
	return nil
}
