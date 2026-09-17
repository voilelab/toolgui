package tclayout

import (
	"fmt"

	"github.com/voilelab/toolgui/toolgui/tgcomp/tcutil"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

var _ tgframe.Component = &dialogComponent{}

const dialogComponentName = "dialog_component"

// The widths a dialog may be given, as the client receives them.
const (
	DialogWidthSmall  = "small"
	DialogWidthMedium = "medium"
	DialogWidthLarge  = "large"
)

type dialogComponent struct {
	*tgframe.BaseComponent

	Title  string `json:"title"`
	Opened bool   `json:"opened"`

	Width       string `json:"width"`
	Dismissible bool   `json:"dismissible"`
}

func newDialogComponent(title string) *dialogComponent {
	return &dialogComponent{
		BaseComponent: &tgframe.BaseComponent{
			Name: dialogComponentName,
			// Whether the dialog is open is kept in the session state under
			// this id, so the dialog needs a name of its own to find it again
			// on the next run.
			ID: tcutil.NormalID(dialogComponentName, title),
		},
		Title: title,
	}
}

// dialogWidth settles the width an unset conf means, and rejects one that
// names nothing the client can render.
func dialogWidth(width string) string {
	switch width {
	case "":
		return DialogWidthSmall
	case DialogWidthSmall, DialogWidthMedium, DialogWidthLarge:
		return width
	}

	panic(fmt.Sprintf("toolgui: unsupported dialog width: %q", width))
}

// DialogConf is the configuration for the Dialog component.
type DialogConf struct {
	tgframe.Base

	// Width is DialogWidthSmall (default), DialogWidthMedium or
	// DialogWidthLarge. Anything else panics.
	Width string

	// Dismissible lets the app user close the dialog with X, ESC or a click
	// outside, default on. Set it with SetDismissible.
	//
	// It is a pointer, against the positive naming the other confs use for
	// their bools, because the default is on: a plain bool's zero value would
	// mean "not dismissible", which is the rarer of the two and not what a
	// dialog written without a conf should be. DataFrameConf.Sortable is
	// the same trade.
	Dismissible *bool
}

// SetDismissible sets Dismissible, which is a pointer so that leaving it out
// means dismissible rather than not.
func (c *DialogConf) SetDismissible(v bool) *DialogConf {
	c.Dismissible = &v
	return c
}

// DialogContainer is what [Dialog] hands out: the dialog it drew, to be
// opened, closed and filled while the page function runs.
type DialogContainer struct {
	comp *dialogComponent

	// owner is the container the dialog was written into, which is what an
	// Open or a Close is sent through; inner is the container the body of the
	// dialog goes into.
	owner *tgframe.Container
	inner *tgframe.Container

	open bool
}

// Open opens the dialog, from anywhere in the run: the client is told
// separately from the node itself, so a dialog already written this run opens
// without the page function having to draw it again.
//
// The body is not redrawn by this. Open before [DialogContainer.With], or the
// dialog opens on a body that this run left empty.
func (d *DialogContainer) Open() {
	d.setOpen(true)
}

// Close closes the dialog. The body this run already wrote stays where it is
// until the run ends; the next run does not write it, and the client drops it
// then.
func (d *DialogContainer) Close() {
	d.setOpen(false)
}

// IsOpen reports whether the dialog is open, as of this point in the run.
func (d *DialogContainer) IsOpen() bool {
	return d.open
}

func (d *DialogContainer) setOpen(open bool) {
	d.open = open
	d.comp.Opened = open

	if d.owner.State != nil {
		d.owner.State.Set(d.comp.ID, open)
	}

	d.owner.SendNotifyPack(tgframe.NewNotifyPackUpdate(d.comp))
}

// With writes the body of the dialog, and calls f only while the dialog is
// open. What a closed dialog would have shown is never computed, so the query
// behind a "delete these 3 rows?" does not run on every rerun.
func (d *DialogContainer) With(f func(c *tgframe.Container)) {
	if !d.open {
		return
	}

	f(d.inner)
}

// Dialog creates a dialog, closed until something opens it, and hands back a
// handle to open, close and fill it:
//
//	d := tgcomp.Dialog(p.Main, "Delete")
//
//	if tgcomp.Button(p.Main, "Delete") {
//		d.Open()
//	}
//
//	d.With(func(c *tgframe.Container) {
//		tgcomp.Text(c, "Delete "+name+"?")
//		if tgcomp.Button(c, "Yes") {
//			remove(name)
//			d.Close()
//		}
//	})
//
// Handle what opens the dialog before calling [DialogContainer.With], as
// above: With is what draws the body, so an Open after it opens a dialog this
// run left empty, and the body only appears on the run after.
//
// Whether the dialog is open is kept by the page rather than the client, so
// it claims an id derived from its title. Two dialogs with the same title
// collide; give one of them a Conf.ID.
//
// Widgets inside the body keep their state while the dialog is closed, as any
// component hidden behind an `if` does. Wrap the body in a
// [github.com/voilelab/toolgui/toolgui/tgcomp.Empty] slot and clear it to
// throw that state away.
func Dialog(c *tgframe.Container, title string, conf ...*DialogConf) *DialogContainer {
	cf := tgframe.OneConf("Dialog", conf)

	comp := newDialogComponent(title)
	comp.Width = dialogWidth(cf.Width)
	comp.Dismissible = boolOr(cf.Dismissible, true)
	tgframe.SetConfID(comp, cf)

	if c.State != nil {
		comp.Opened, _ = c.State.Get[bool](comp.ID)
	}

	// Sent whether or not it is open: an Open later in the run addresses the
	// node by its key, so the node has to be there for the key to exist. A
	// closed dialog renders as a portal that holds nothing and takes up no
	// room on the page.
	c.AddComponent(comp)

	return &DialogContainer{
		comp:  comp,
		owner: c,
		inner: c.AddContainerTo(comp, "inner", 0),
		open:  comp.Opened,
	}
}

// boolOr reports what an unset conf toggle means.
func boolOr(v *bool, def bool) bool {
	if v == nil {
		return def
	}
	return *v
}
