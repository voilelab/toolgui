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

// runStats is one run of a page as the transport sees it: every pack a
// [tgframe.Session] sends, marshalled the way tgexec and tgwasm marshal them.
type runStats struct {
	total     time.Duration // event to result, marshalling included
	marshal   time.Duration // transport marshalling alone
	bytes     int
	dfBytes   int
	dfMarshal time.Duration
}

// sessionRunner runs a page through a Session, so a rerun gets the keep packs
// a real client would (TG-99).
type sessionRunner struct {
	session *tgframe.Session
	st      runStats
	done    chan struct{}
}

func newSessionRunner(tb testing.TB, app *tgframe.App, name string) *sessionRunner {
	r := &sessionRunner{done: make(chan struct{}, 1)}
	s, err := tgframe.NewSession(app, name, nil, tgframe.NewState(), func(pack any) error {
		t := time.Now()
		bs, err := tgjson.Marshal(pack)
		d := time.Since(t)
		if err != nil {
			return err
		}

		if res, ok := pack.(*tgframe.ResultPack); ok {
			if !res.Success {
				tb.Error(res.Error)
			}
			r.done <- struct{}{}
			return nil
		}
		if _, ok := pack.(tgframe.NotifyPack); !ok {
			return nil
		}

		r.st.marshal += d
		r.st.bytes += len(bs)
		if isDataFrame(bs) {
			r.st.dfBytes += len(bs)
			r.st.dfMarshal += d
		}
		return nil
	})
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(s.Close)

	r.session = s
	return r
}

func (r *sessionRunner) run() runStats {
	r.st = runStats{}
	start := time.Now()
	r.session.HandleEvent(&tgframe.EventEmpty{})
	<-r.done
	r.st.total = time.Since(start)
	return r.st
}

// isDataFrame tells the table's pack apart by its wire name, since the pack
// types are unexported.
func isDataFrame(bs []byte) bool {
	return bytes.Contains(bs, []byte(`"name":"dataframe_component"`))
}

// TestReport prints the numbers TG-94 asks for. A rerun after the first is
// what flipping the Toggle costs: the table is unchanged, so it goes out as a
// keep pack.
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
	fmt.Printf("%-6s %10s %10s %10s %6s %12s %12s %12s\n",
		"rows", "first", "run bytes", "df bytes", "df %", "run time", "marshal", "df marshal")
	for _, n := range rowCounts {
		r := newSessionRunner(t, app, pageName(n))
		first := r.run() // first load, warms caches

		var sum runStats
		var last runStats
		for range reps {
			last = r.run()
			sum.total += last.total
			sum.marshal += last.marshal
			sum.dfMarshal += last.dfMarshal
		}

		pct := 0.0
		if last.bytes > 0 {
			pct = 100 * float64(last.dfBytes) / float64(last.bytes)
		}
		fmt.Printf("%-6d %10d %10d %10d %5.1f%% %12v %12v %12v\n",
			n, first.bytes, last.bytes, last.dfBytes, pct,
			sum.total/reps, sum.marshal/reps, sum.dfMarshal/reps)
	}
}

func BenchmarkRerun(b *testing.B) {
	app := newApp()
	for _, n := range rowCounts {
		b.Run(pageName(n), func(b *testing.B) {
			r := newSessionRunner(b, app, pageName(n))
			r.run()
			b.ReportAllocs()
			for b.Loop() {
				r.run()
			}
		})
	}
}
