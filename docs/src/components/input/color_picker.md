# Color Picker

ColorPicker create a color picker and return the picked color.

## API

### Interface

```go
func ColorPicker(c *tgframe.Container, label string, conf ...*ColorPickerConf) string
```

### Parameters

* `c` is Parent container.
* `label` is the label of the color picker.
* `conf` is an optional configuration, at most one.
* Return the picked color as a lowercase `#rrggbb` string. Before anything is
  picked, that is the conf's `Default`.

```go
// ColorPickerConf is the configuration for a color picker.
type ColorPickerConf struct {
	tgframe.Base // ID

	// Default is the color the picker starts on, as "#rrggbb". Defaults to
	// black.
	Default string

	// Disabled is true if the color picker is disabled.
	Disabled bool
}
```

The return is always seven characters, always lowercase, and never empty, so it
can go straight into a style or a chart color. A `Default` that is not an
`#rrggbb` color is a mistake in the call rather than a value to correct, and
panics — `"#fff"` and `"red"` included.

The color is sent to Go when the pointer is released, not while it is being
dragged over the picker.

## Example

```go
color := tgcomp.ColorPicker(p.Main, "ColorPicker",
	&tgcomp.ColorPickerConf{Default: "#ff3860"})

tgcomp.Text(p.Main, "Value: "+color,
	&tgcomp.TextConf{ID: "color_picker_result"})
```
