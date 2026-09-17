# MultiSelect

MultiSelect create a dropdown list that takes more than one item and return
the indices of the selected ones.

## API

```go
func MultiSelect(c *tgframe.Container, label string, items []string, conf ...*MultiSelectConf) []int
```

* `c` is Parent container.
* `label` is the label for multiselect.
* `items` is the list of options.
* `conf` is an optional configuration, at most one.
* Return the indices of the selected items, 0-indexed. nil when nothing is
  selected.

The result is ordered by `items` rather than by the order the app user picked
them in, so the same selection always reads the same way.

Nothing selected is nil, the same "nothing" [Select](select.md) hands back, so
one check reads a single pick and a multiple one. Ranging over it is safe
either way — a nil slice has no elements — so the check is only needed where
the absence itself matters.

```go
// MultiSelectConf is the configuration for the MultiSelect component.
type MultiSelectConf struct {
	tgframe.Base // ID

	// Default is the selection the component starts with, as indices into
	// items. It is only read until the app user first touches the component.
	Default []int

	// MaxSelections is how many items may be selected at once. Zero, the
	// default, is no limit. At the limit the frontend disables the items that
	// are not selected, so the limit is never reached by a refusal.
	MaxSelections int

	// Placeholder is the text shown while nothing is selected.
	Placeholder string

	// Disabled is true if the multiselect is disabled.
	Disabled bool
}
```

## Example

```go
{{#include ../../../demos/multiselect.go:demo}}
```

Like a [Select](select.md), a multiselect derives its id from its label, so two
with the same label collide. Naming either of them is the way out:

```go
tgcomp.MultiSelect(p.Main, "Pick", items)
tgcomp.MultiSelect(p.Main, "Pick", items, &tgcomp.MultiSelectConf{ID: "second_pick"})
```

<div data-toolgui-demo="multiselect" data-toolgui-demo-height="640">

![multiselect component](multiselect.png)

</div>
