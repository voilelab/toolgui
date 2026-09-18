package tgcomp

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tclayout"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// Box create a box container.
func Box(c *tgframe.Container, conf ...*BoxConf) *tgframe.Container {
	return tclayout.Box(c, conf...)
}

// BoxConf is the configuration for the Box component.
type BoxConf = tclayout.BoxConf

// Column create N columns.
func Column(
	c *tgframe.Container, n uint, conf ...*ColumnConf) []*tgframe.Container {

	return tclayout.Column(c, n, conf...)
}

// Column1 create 1 column.
func Column1(c *tgframe.Container, conf ...*ColumnConf) *tgframe.Container {
	return tclayout.Column1(c, conf...)
}

// Column2 create 2 columns.
func Column2(c *tgframe.Container, conf ...*ColumnConf) (
	*tgframe.Container, *tgframe.Container) {

	return tclayout.Column2(c, conf...)
}

// Column3 create 3 columns.
func Column3(c *tgframe.Container, conf ...*ColumnConf) (
	*tgframe.Container, *tgframe.Container, *tgframe.Container) {

	return tclayout.Column3(c, conf...)
}

// ColumnConf is the configuration for the column components.
type ColumnConf = tclayout.ColumnConf

// EqColumn create N columns with same width.
func EqColumn(
	c *tgframe.Container, n uint, conf ...*ColumnConf) []*tgframe.Container {

	return tclayout.EqColumn(c, n, conf...)
}

// EqColumn1 create 1 columns with same width.
func EqColumn1(c *tgframe.Container, conf ...*ColumnConf) *tgframe.Container {
	return tclayout.EqColumn1(c, conf...)
}

// EqColumn2 create 2 columns with same width.
func EqColumn2(c *tgframe.Container, conf ...*ColumnConf) (
	*tgframe.Container, *tgframe.Container) {

	return tclayout.EqColumn2(c, conf...)
}

// EqColumn3 create 3 columns with same width.
func EqColumn3(c *tgframe.Container, conf ...*ColumnConf) (
	*tgframe.Container, *tgframe.Container, *tgframe.Container) {

	return tclayout.EqColumn3(c, conf...)
}

// EqColumn4 create 4 columns with same width.
func EqColumn4(c *tgframe.Container, conf ...*ColumnConf) (
	*tgframe.Container, *tgframe.Container, *tgframe.Container,
	*tgframe.Container) {

	return tclayout.EqColumn4(c, conf...)
}

// EqColumn5 create 5 columns with same width.
func EqColumn5(c *tgframe.Container, conf ...*ColumnConf) (
	*tgframe.Container, *tgframe.Container, *tgframe.Container,
	*tgframe.Container, *tgframe.Container) {

	return tclayout.EqColumn5(c, conf...)
}

// Toolbar creates a row of controls and returns the container they go into.
// What is written inside lines up horizontally rather than taking a row of the
// page each.
func Toolbar(c *tgframe.Container, conf ...*ToolbarConf) *tgframe.Container {
	return tclayout.Toolbar(c, conf...)
}

// ToolbarConf is the configuration for the Toolbar component.
type ToolbarConf = tclayout.ToolbarConf

// The ways a Toolbar may line its items up.
const (
	ToolbarJustifyStart   = tclayout.ToolbarJustifyStart
	ToolbarJustifyEnd     = tclayout.ToolbarJustifyEnd
	ToolbarJustifyBetween = tclayout.ToolbarJustifyBetween
)

// Tab create a tab component.
func Tab(
	c *tgframe.Container, tabs []string,
	conf ...*TabConf) []*tgframe.Container {

	return tclayout.Tab(c, tabs, conf...)
}

// Tab2 create 2 tabs.
func Tab2(c *tgframe.Container, tab1, tab2 string, conf ...*TabConf) (
	*tgframe.Container, *tgframe.Container) {

	return tclayout.Tab2(c, tab1, tab2, conf...)
}

// Tab3 create 3 tabs.
func Tab3(c *tgframe.Container, tab1, tab2, tab3 string, conf ...*TabConf) (
	*tgframe.Container, *tgframe.Container, *tgframe.Container) {

	return tclayout.Tab3(c, tab1, tab2, tab3, conf...)
}

// Tab4 create 4 tabs.
func Tab4(
	c *tgframe.Container, tab1, tab2, tab3, tab4 string, conf ...*TabConf) (
	*tgframe.Container, *tgframe.Container, *tgframe.Container,
	*tgframe.Container) {

	return tclayout.Tab4(c, tab1, tab2, tab3, tab4, conf...)
}

// Tab5 create 5 tabs.
func Tab5(
	c *tgframe.Container, tab1, tab2, tab3, tab4, tab5 string,
	conf ...*TabConf) (
	*tgframe.Container, *tgframe.Container, *tgframe.Container,
	*tgframe.Container, *tgframe.Container) {

	return tclayout.Tab5(c, tab1, tab2, tab3, tab4, tab5, conf...)
}

// TabConf is the configuration for the tab components.
type TabConf = tclayout.TabConf

// Expand create a expandable component.
func Expand(
	c *tgframe.Container, title string, expanded bool,
	conf ...*ExpandConf) *tgframe.Container {

	return tclayout.Expand(c, title, expanded, conf...)
}

// ExpandConf is the configuration for the Expand component.
type ExpandConf = tclayout.ExpandConf

// Popover create a button with a floating panel behind it.
func Popover(
	c *tgframe.Container, label string,
	conf ...*PopoverConf) *tgframe.Container {

	return tclayout.Popover(c, label, conf...)
}

// PopoverConf is the configuration for the Popover component.
type PopoverConf = tclayout.PopoverConf

// Dialog creates a dialog, closed until something opens it, and hands back a
// handle to open, close and fill it. The body is only computed while it is
// open, so handle what opens the dialog before calling DialogContainer.With.
func Dialog(
	c *tgframe.Container, title string,
	conf ...*DialogConf) *DialogContainer {

	return tclayout.Dialog(c, title, conf...)
}

// DialogConf is the configuration for the Dialog component.
type DialogConf = tclayout.DialogConf

// DialogContainer is what Dialog hands out: the dialog it drew, to be opened,
// closed and filled while the page function runs.
type DialogContainer = tclayout.DialogContainer

// The widths a Dialog may be given.
const (
	DialogWidthSmall  = tclayout.DialogWidthSmall
	DialogWidthMedium = tclayout.DialogWidthMedium
	DialogWidthLarge  = tclayout.DialogWidthLarge
)

// Empty reserves a place in the page and hands back a slot to write it with.
// Writing the slot again takes the previous contents off the screen instead of
// adding to them.
func Empty(c *tgframe.Container, conf ...*EmptyConf) *EmptySlot {
	return tclayout.Empty(c, conf...)
}

// EmptyConf is the configuration for the Empty component.
type EmptyConf = tclayout.EmptyConf

// EmptySlot is what Empty hands out: a place in the page that can be written
// and written over.
type EmptySlot = tclayout.EmptySlot

// EmptyContainer is the old name of EmptySlot.
//
// Deprecated: use [EmptySlot].
type EmptyContainer = tclayout.EmptySlot
