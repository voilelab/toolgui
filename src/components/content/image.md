# Image

Image component display an image.

## API

### Interface

```go
func Image(c *tgframe.Container, img any)
func ImageWithConf(c *tgframe.Container, img any, conf *ImageConf)
```

### Parameters

* `c` is Parent container.
* `img` is the image.
  * `image.Image`: an image.Image
  * `[]byte`: a byte array
  * `string`: a url or base64 encoded image
    * example:
      * url: `https://http.cat/100`
      * base64 uri: `data:image/png;base64,...`
* `conf` is the configuration of the image.

```go
// ImageConf is the configuration for the Image component
type ImageConf struct {
	// Width is the width of the image (e.g. "100px", "50%")
	Width string

	// Format is the format of the image, default is "png"
	Format ImageFormat

	// ID is the unique identifier for this image component
	ID string
}
```

## Example

```go
tgcomp.Image(p.Main, "https://http.cat/100")
```

![image component](image.png)
