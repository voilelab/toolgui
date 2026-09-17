package tcinput_test

import (
	"testing"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcinput"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// TestDownloadFileClicked pins the click reporting, whose ids come from the
// component's own name.
func TestDownloadFileClicked(t *testing.T) {
	newDownloadFileRunner := func(t *testing.T) *runner {
		t.Helper()

		return newRunner(t, func(p *tgframe.Params, seen *clickRun) {
			conf := &tcinput.DownloadFileConf{ID: "report"}
			seen.before = tcinput.DownloadFileClicked(p.Main, "Report", conf)
			seen.drawn = tcinput.DownloadFile(p.Main, "Report",
				[]byte("body"), conf)
		})
	}

	t.Run("clicked", func(t *testing.T) {
		r := newDownloadFileRunner(t)

		got := r.run(&tgframe.EventClick{ID: "download_file_component_report"})
		if !got.before {
			t.Error("DownloadFileClicked = false before the draw, want true")
		}

		if got.before != got.drawn {
			t.Errorf("DownloadFileClicked = %v, DownloadFile = %v,"+
				" want the same", got.before, got.drawn)
		}
	})

	t.Run("another button", func(t *testing.T) {
		r := newDownloadFileRunner(t)

		got := r.run(&tgframe.EventClick{ID: "download_file_component_other"})
		if got.before || got.drawn {
			t.Errorf("DownloadFileClicked = %v, DownloadFile = %v,"+
				" want both false", got.before, got.drawn)
		}
	})

	t.Run("never drawn", func(t *testing.T) {
		r := newRunner(t, func(p *tgframe.Params, seen *clickRun) {
			seen.before = tcinput.DownloadFileClicked(p.Main, "Secret",
				&tcinput.DownloadFileConf{ID: "secret"})
		})

		got := r.run(&tgframe.EventClick{ID: "download_file_component_secret"})
		if got.before {
			t.Error("DownloadFileClicked = true for a button the page never" +
				" drew, want false")
		}
	})
}
