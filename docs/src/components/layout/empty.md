# Empty

Empty reserves a place in the page and hands back a slot to write it with.
Writing the slot again takes the previous contents off the screen instead of
adding to them, which is what lets a page function show progress and then
replace it with the result.

## Usage

```go
func Empty(c *tgframe.Container, conf ...*EmptyConf) *EmptySlot
```

* `c`: Parent container.
* `conf`: Optional configuration, at most one.

```go
// EmptyConf is the configuration for the Empty component.
type EmptyConf struct {
	tgframe.Base // ID
}
```

The container an empty hands out derives its id from the empty's; give none
and it carries none, and the components inside are still placed by position.

The slot has two methods:

```go
// With writes what f adds into the slot, over whatever it held.
func (s *EmptySlot) With(f func(c *tgframe.Container))

// Clear takes what the slot holds off the screen.
func (s *EmptySlot) Clear()
```

`EmptySlot` was called `EmptyContainer`; the old name is kept as a deprecated
alias. See [what a component hands
back](../../architecture/components.md#what-a-component-hands-back).

The container `With` hands over lives until the next `With` or `Clear`.
Writing into it after that puts components under a node the client no longer
has, so take it in the callback rather than keeping it.

## Example

```go
slot := tgcomp.Empty(p.Main)

slot.With(func(c *tgframe.Container) {
	tgcomp.Text(c, "Querying…")
})

head, rows := query()

slot.With(func(c *tgframe.Container) {
	tgcomp.Table(c, head, rows)
})
```

## Ids and state

A widget in a slot claims its id the same way it would anywhere else, and
gives it back when the slot is cleared. So the same widget may be written into
one slot as many times as the page likes without the run failing with
`duplicated component id`:

```go
slot := tgcomp.Empty(p.Main)
for range names {
	slot.With(func(c *tgframe.Container) {
		tgcomp.Textbox(c, "Name")
	})
}
```

A widget that is cleared and *not* written again is gone from the page, and
its state goes with it: the value it held is dropped at the end of the run,
rather than turning up in whatever lands on that id next run.

The slot also starts empty on every run, whatever the last run left in it.
