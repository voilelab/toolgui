package demos

import (
	"fmt"
	"image/jpeg"
	"strings"

	"github.com/voilelab/toolgui/toolgui/tgcomp"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

func fileuploadDemo(p *tgframe.Params) error {
	// ANCHOR: demo
	fileObj := tgcomp.FileUpload(p.Main, "FileUpload", ".jpg,.png")
	if fileObj == nil {
		return nil
	}

	tgcomp.Text(p.Main, "FileUpload filename: "+fileObj.Name)
	tgcomp.Text(p.Main, fmt.Sprintf("FileUpload bytes length: %d", fileObj.Size))
	if strings.HasSuffix(fileObj.Name, ".jpg") {
		// Decoding reads the upload off disk, so the image never has to
		// be held twice.
		fp, err := fileObj.Open()
		if err != nil {
			return nil
		}
		defer fp.Close()

		img, err := jpeg.Decode(fp)
		if err == nil {
			tgcomp.Image(p.Main, img)
		}
	}
	// ANCHOR_END: demo
	return nil
}
