package tcmisc

import (
	"fmt"
	"time"

	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &toastComponent{}
var toastComponentName = "toast_component"

type toastComponent struct {
	*tgframe.BaseComponent

	Text string `json:"text"`
	Icon string `json:"icon"`

	// DurationMS is how long the toast stays, in milliseconds. Zero leaves it
	// to the client's default.
	DurationMS int64 `json:"duration_ms"`

	// Seq is the serial of the run that wrote this toast.
	//
	// A toast is something that happened, and the page is a tree of nodes. The
	// client keys a node by where it sits, so the same Toast call in the same
	// place on the next run arrives as the node that is already there, with
	// the same text and the same id -- nothing tells the client a second toast
	// was asked for. The serial is the one prop that differs between runs, and
	// it is what the client fires on.
	Seq uint64 `json:"seq"`
}

func newToastComponent(text string) *toastComponent {
	return &toastComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: toastComponentName,
		},
		Text: text,
	}
}

// toastDuration turns a conf's duration into the milliseconds the client
// reads. Zero stays zero, which is the client's default; anything under a
// millisecond becomes one, so a duration that was asked for is never rounded
// away into meaning "default".
func toastDuration(d time.Duration) int64 {
	if d < 0 {
		panic(fmt.Sprintf("toolgui: Toast duration must not be negative: %v", d))
	}

	ms := d.Milliseconds()
	if d > 0 && ms == 0 {
		return 1
	}

	return ms
}

// ToastConf is the configuration for the Toast component.
type ToastConf struct {
	tgframe.Base

	// Icon is an emoji shown in front of the text. An emoji shortcode such as
	// `:tada:` works too.
	Icon string

	// Duration is how long the toast stays. Zero is the default. A negative
	// duration panics.
	Duration time.Duration
}

// Toast shows a one-off notification that takes itself off the screen again:
//
//	if tgcomp.Button(c, "Save") {
//		save()
//		tgcomp.Toast(c, "Saved", &tgcomp.ToastConf{Icon: "✅"})
//	}
//
// It leaves a node on the page that renders nothing, so it takes up no room
// and does not move what is around it. Toasts stack in the order they were
// written, and hovering one holds them all on screen.
//
// A toast is fired once per run. Every run that reaches the call fires it
// again, including a run the same button starts a second time, and a run that
// nothing on the page changed -- which is what makes it a notification rather
// than a [Message] that happens to fade. A run that never reaches the call,
// because an `if` above it was false or because the page function failed
// first, fires nothing.
//
// Whether the run that fired it finishes does not matter: a toast the client
// has already been told about lives out its duration there, and an
// interrupted run cannot take it back.
func Toast(c *tgframe.Container, text string, conf ...*ToastConf) {
	cf := tgframe.OneConf("Toast", conf)

	comp := newToastComponent(text)
	comp.Icon = cf.Icon
	comp.DurationMS = toastDuration(cf.Duration)
	comp.Seq = c.RunSeq()
	tgframe.SetConfID(comp, cf)

	c.AddComponent(comp)
}
