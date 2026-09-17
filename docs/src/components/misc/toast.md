# Toast

Toast shows a one-off notification that takes itself off the screen again.

```go
if tgcomp.Button(c, "Save") {
	save()
	tgcomp.Toast(c, "Saved", &tgcomp.ToastConf{Icon: "✅"})
}
```

## API

### Interface

```go
func Toast(c *tgframe.Container, text string, conf ...*ToastConf)
```

### Parameters

* `c`: Parent container.
* `text`: Text of the notification. Emoji shortcodes such as `:tada:` work.
* `conf`: Optional configuration, at most one.

```go
// ToastConf is the configuration for the Toast component.
type ToastConf struct {
	tgframe.Base // ID

	// Icon is an emoji shown in front of the text.
	Icon string

	// Duration is how long the toast stays. Zero is the default.
	Duration time.Duration
}
```

A negative `Duration` panics. There is no way to ask for a toast that never
goes away; a notification the user has to dismiss is a
[Message](message.md).

Toast hands nothing back. It is a notification that has been sent, not
something the page keeps: it cannot be updated or taken down again. See [what
a component hands back](../../architecture/components.md#what-a-component-hands-back).

## One toast per run

**Every run that reaches the call fires the toast again.** A second click of
the same button shows it a second time, and so does a rerun that nothing on
the page caused. This is the same rule Streamlit's `st.toast` follows, and it
is what makes a toast a notification rather than a `Message` that happens to
fade.

A run that never reaches the call, because an `if` above it was false or
because the page function failed first, fires nothing.

So put a Toast behind whatever it is reporting, not beside it:

```go
// Fires once, when the button is clicked.
if tgcomp.Button(c, "Save") {
	save()
	tgcomp.Toast(c, "Saved", &tgcomp.ToastConf{Icon: "✅"})
}

// Fires on every run, including the ones the user did not start.
tgcomp.Toast(c, "Saved", &tgcomp.ToastConf{Icon: "✅"})
```

Once a toast is out it is out. A run that is interrupted partway through —
the user clicked something else while the page function was still working —
does not take back a toast it already sent.

## On the page

A toast takes up no room. The node it leaves where the page function wrote it
renders nothing, and the notification itself is drawn over the page, so
nothing around the call moves when one fires.

Toasts stack in the order they were fired, newest nearest the corner, and five
are shown at a time; more than that wait their turn. Hovering any of them, or
moving the keyboard focus into one, holds them all on screen until the pointer
leaves again.

## Example

Two toasts from one run, one of them left up longer than the default:

```go
if tgcomp.Button(c, "Save") {
	tgcomp.Toast(c, "Saved to disk", &tgcomp.ToastConf{Icon: "✅"})
	tgcomp.Toast(c, "Two rows changed", &tgcomp.ToastConf{
		Icon:     "📝",
		Duration: 10 * time.Second,
	})
}
```
