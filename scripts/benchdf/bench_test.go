package main

import (
	"bytes"
	"flag"
	"fmt"
	"testing"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgjson"
)

var report = flag.Bool("report", false, "print the TG-94 report")

// runStats is one run of a page as the transport sees it: every pack
// marshalled the way tgexec and tgwasm marshal them.
type runStats struct {
	total     time.Duration // page function plus marshalling
	marshal   time.Duration // marshalling alone
	bytes     int
	dfBytes   int
	dfMarshal time.Duration
}

func runOnce(tb testing.TB, app *tgframe.App, name string, state *tgframe.State) runStats {
	var st runStats
	start := time.Now()
	err := app.Run(name, state, func(pack tgframe.NotifyPack) {
		t := time.Now()
		bs, err := tgjson.Marshal(pack)
		d := time.Since(t)
		if err != nil {
			tb.Fatal(err)
		}
		st.marshal += d
		st.bytes += len(bs)
		if isDataFrame(bs) {
			st.dfBytes += len(bs)
			st.dfMarshal += d
		}
	})
	st.total = time.Since(start)
	if err != nil {
		tb.Fatal(err)
	}
	return st
}

// isDataFrame tells the table's pack apart by its wire name, since the pack
// types are unexported.
func isDataFrame(bs []byte) bool {
	return bytes.Contains(bs, []byte(`"name":"dataframe_component"`))
}

// TestReport prints the numbers TG-94 asks for. A rerun after the first is
// what flipping the Toggle costs: the table is unchanged but sent again.
//
// Opt-in by a flag rather than an env var, since the wasm runner hands the
// test no environment:
//
//	go test ./scripts/benchdf -run TestReport -v -args -report
func TestReport(t *testing.T) {
	if !*report {
		t.Skip("pass -args -report to measure")
	}

	const reps = 20

	app := newApp()
	fmt.Printf("%-6s %10s %10s %6s %12s %12s %12s\n",
		"rows", "run bytes", "df bytes", "df %", "run time", "marshal", "df marshal")
	for _, n := range rowCounts {
		state := tgframe.NewState()
		runOnce(t, app, pageName(n), state) // first load, warms caches

		var sum runStats
		var last runStats
		for range reps {
			last = runOnce(t, app, pageName(n), state)
			sum.total += last.total
			sum.marshal += last.marshal
			sum.dfMarshal += last.dfMarshal
		}

		pct := 0.0
		if last.bytes > 0 {
			pct = 100 * float64(last.dfBytes) / float64(last.bytes)
		}
		fmt.Printf("%-6d %10d %10d %5.1f%% %12v %12v %12v\n",
			n, last.bytes, last.dfBytes, pct,
			sum.total/reps, sum.marshal/reps, sum.dfMarshal/reps)
	}
}

func BenchmarkRerun(b *testing.B) {
	app := newApp()
	for _, n := range rowCounts {
		b.Run(pageName(n), func(b *testing.B) {
			state := tgframe.NewState()
			runOnce(b, app, pageName(n), state)
			b.ReportAllocs()
			for b.Loop() {
				runOnce(b, app, pageName(n), state)
			}
		})
	}
}
