package tccontent

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/voilelab/toolgui/toolgui/tgframe"
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
	// Width is the width of the image (e.g. "100px", "50%")
	Width string

	// Format is the format of the image, default is "png"
	Format ImageFormat

	// ID is the unique identifier for this image component
	ID string
}

// Image show an image.
func Image(c *tgframe.Container, img any) {
	ImageWithConf(c, img, nil)
}

// ImageWithConf show an image with a custom configuration.
func ImageWithConf(c *tgframe.Container, img any, conf *ImageConf) {
	if conf == nil {
		conf = &ImageConf{}
	}

	formatStr := ""
	switch conf.Format {
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
		uri = fmt.Sprintf("data:image/%s;base64,%s",
			formatStr, base64.StdEncoding.EncodeToString(v))
	case image.Image:
		var imageBuf bytes.Buffer
		switch conf.Format {
		case ImageFormatPNG:
			err := png.Encode(&imageBuf, v)
			if err != nil {
				panic(err)
			}
			formatStr = "png"
		case ImageFormatJPEG:
			err := jpeg.Encode(&imageBuf, v, nil)
			if err != nil {
				panic(err)
			}
			formatStr = "jpeg"
		default:
			err := fmt.Errorf("unsupported image format: %v", conf.Format)
			panic(err)
		}
		bs := imageBuf.Bytes()
		b64 := base64.StdEncoding.EncodeToString(bs)
		uri = fmt.Sprintf("data:image/%s;base64,%s",
			formatStr, b64)
	default:
		panic("unsupported image type")
	}

	comp := newImageComponent(uri)

	if conf.Width != "" {
		comp.Width = conf.Width
	}

	if conf.ID != "" {
		comp.SetID(conf.ID)
	}

	c.AddComponent(comp)
}
