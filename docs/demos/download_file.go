package demos

import (
	"time"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func downloadFileDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	// A megabyte, which is past what belongs in a data: URI, and a pattern
	// rather than noise so a byte anywhere in the file is known from its
	// offset alone.
	body := make([]byte, 1<<20)
	for i := range body {
		body[i] = byte(i % 251)
	}

	if tgcomp.DownloadFile(
		p.Main, "Save a megabyte", body,
		&tgcomp.DownloadFileConf{
			Filename: "pattern.bin",
			Color:    tcutil.ColorInfo,
		}) {
		tgcomp.Text(p.Main, "Megabyte saved!")
	}
	// ANCHOR_END: demo
	return nil
}

func downloadFileFuncDemo(p *tgframe.Params) error {
	// ANCHOR: func
	// Made on click, so a run that is not about the file never builds it.
	if tgcomp.DownloadFileFunc(
		p.Main, "Export a timestamp", func() ([]byte, error) {
			return []byte(time.Now().Format(time.RFC3339)), nil
		},
		&tgcomp.DownloadFileConf{
			Filename: "now.txt",
			MIME:     "text/plain",
		}) {
		tgcomp.Text(p.Main, "Timestamp exported!")
	}
	// ANCHOR_END: func
	return nil
}
