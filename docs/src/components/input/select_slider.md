# Select Slider

SelectSlider create a slider over a list of discrete items and return the index
of the one it sits on.

## API

### Interface

```go
func SelectSlider(c *tgframe.Container, label string, items []string,
	conf ...*SelectSliderConf) *int
```

### Parameters

* `c` is Parent container.
* `label` is the label of the slider.
* `items` is the list of options, drawn as marks along the track.
* `conf` is an optional configuration, at most one.
* Return the index of the item the slider sits on, 0-indexed as
  [Select](select.md)'s is. Never nil.

The handle is always on an item, so unlike `Select` there is no "nothing
selected" state and the returned pointer is never nil: it is the index the app
user left it at, else `Default`.

```go
// SelectSliderConf is the configuration for the SelectSlider component.
type SelectSliderConf struct {
	tgframe.Base // ID

	// Default is the index of the item the slider starts on. Defaults to 0.
	Default int
}
```

An empty `items`, or a `Default` outside it, is a mistake in the call rather
than a value to correct, and panics.

Like [Slider](slider.md), the index is sent to Go when the handle is released
rather than while it is being dragged.

## Example

```go
sizes := []string{"S", "M", "L"}
selIdx := tgcomp.SelectSlider(p.Main, "Size", sizes)

tgcomp.Text(p.Main, "Value: "+sizes[*selIdx],
	&tgcomp.TextConf{ID: "select_slider_result"})
```

Use this over [Select](select.md) when the options are ordered — sizes, tiers,
buckets — so that their order is part of what the control shows. For an
unordered list, a dropdown reads better.
