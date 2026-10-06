package tccontent

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"strings"

	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgutil"
)

var _ tgframe.Component = &imageComponent{}
var imageComponentName = "image_component"

type imageComponent struct {
	*tgframe.BaseComponent
	Src   string `json:"src"`
	Width string `json:"width"`
}

func newImageComponent(src string) *imageComponent {
	return &imageComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: imageComponentName,
		},
		Src: src,
	}
}

// ImageFormat is the format of the image
type ImageFormat int

const (
	ImageFormatPNG ImageFormat = iota
	ImageFormatJPEG
)

// ImageConf is the configuration for the Image component
type ImageConf struct {
	tgframe.Base

	// Width is the width of the image (e.g. "100px", "50%")
	Width string

	// Format is the format of the image, default is "png".
	// For []byte, the MIME is detected from magic bytes; Format is the fallback.
	Format ImageFormat
}

// Image show an image.
func Image(c *tgframe.Container, img any, conf ...*ImageConf) {
	cf := tgframe.OneConf("Image", conf)

	formatStr := ""
	switch cf.Format {
	case ImageFormatPNG:
		formatStr = "png"
	case ImageFormatJPEG:
		formatStr = "jpeg"
	default:
		panic("unsupported image format")
	}

	uri := ""
	switch v := img.(type) {
	case string:
		uri = v
	case []byte:
		uri = fmt.Sprintf("data:%s;base64,%s",
			detectImageMIME(v, "image/"+formatStr),
			base64.StdEncoding.EncodeToString(v))
	case image.Image:
		var imageBuf bytes.Buffer
		switch cf.Format {
		case ImageFormatPNG:
			err := png.Encode(&imageBuf, v)
			if err != nil {
				c.Fail(tgutil.Errorf("encode png: %w", err))
				return
			}
			formatStr = "png"
		case ImageFormatJPEG:
			err := jpeg.Encode(&imageBuf, v, nil)
			if err != nil {
				c.Fail(tgutil.Errorf("encode jpeg: %w", err))
				return
			}
			formatStr = "jpeg"
		default:
			panic(fmt.Sprintf("unsupported image format: %v", cf.Format))
		}
		bs := imageBuf.Bytes()
		b64 := base64.StdEncoding.EncodeToString(bs)
		uri = fmt.Sprintf("data:image/%s;base64,%s",
			formatStr, b64)
	default:
		panic("unsupported image type")
	}

	comp := newImageComponent(uri)

	if cf.Width != "" {
		comp.Width = cf.Width
	}

	tgframe.SetConfIDIn(c, comp, cf)

	c.AddComponent(comp)
}

// detectImageMIME sniffs the image MIME from magic bytes, or returns fallback.
func detectImageMIME(bs []byte, fallback string) string {
	mime := http.DetectContentType(bs)
	if strings.HasPrefix(mime, "image/") {
		return mime
	}
	if isSVG(bs) {
		return "image/svg+xml"
	}
	return fallback
}

// isSVG reports whether bs starts with an <svg> root, skipping BOM,
// whitespace, XML declaration, comments and DOCTYPE.
func isSVG(bs []byte) bool {
	bs = bytes.TrimPrefix(bs, []byte("\xEF\xBB\xBF"))
	for {
		bs = bytes.TrimLeft(bs, " \t\r\n")
		var end []byte
		switch {
		case bytes.HasPrefix(bs, []byte("<?")):
			end = []byte("?>")
		case bytes.HasPrefix(bs, []byte("<!--")):
			end = []byte("-->")
		case bytes.HasPrefix(bs, []byte("<!")):
			end = []byte(">")
		default:
			return bytes.HasPrefix(bs, []byte("<svg")) && len(bs) > 4 &&
				strings.ContainsRune(" \t\r\n>/", rune(bs[4]))
		}
		i := bytes.Index(bs, end)
		if i < 0 {
			return false
		}
		bs = bs[i+len(end):]
	}
}
