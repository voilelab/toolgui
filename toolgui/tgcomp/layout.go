package tgcomp

import "github.com/voilelab/toolgui/toolgui/tgcomp/tclayout"

// Box create a box container.
var Box = tclayout.Box

// BoxConf is the configuration for the Box component.
type BoxConf = tclayout.BoxConf

// Column create N columns.
var Column = tclayout.Column

// Column1 create 1 column.
var Column1 = tclayout.Column1

// Column2 create 2 columns.
var Column2 = tclayout.Column2

// Column3 create 3 columns.
var Column3 = tclayout.Column3

// ColumnConf is the configuration for the column components.
type ColumnConf = tclayout.ColumnConf

// EqColumn create N columns with same width.
var EqColumn = tclayout.EqColumn

// EqColumn1 create 1 columns with same width.
var EqColumn1 = tclayout.EqColumn1

// EqColumn2 create 2 columns with same width.
var EqColumn2 = tclayout.EqColumn2

// EqColumn3 create 3 columns with same width.
var EqColumn3 = tclayout.EqColumn3

// EqColumn4 create 4 columns with same width.
var EqColumn4 = tclayout.EqColumn4

// EqColumn5 create 5 columns with same width.
var EqColumn5 = tclayout.EqColumn5

// Tab create a tab component.
var Tab = tclayout.Tab

// Tab2 create 2 tabs.
var Tab2 = tclayout.Tab2

// Tab3 create 3 tabs.
var Tab3 = tclayout.Tab3

// Tab4 create 4 tabs.
var Tab4 = tclayout.Tab4

// Tab5 create 5 tabs.
var Tab5 = tclayout.Tab5

// TabConf is the configuration for the tab components.
type TabConf = tclayout.TabConf

// Expand create a expandable component.
var Expand = tclayout.Expand

// ExpandConf is the configuration for the Expand component.
type ExpandConf = tclayout.ExpandConf

// Popover create a button with a floating panel behind it.
var Popover = tclayout.Popover

// PopoverConf is the configuration for the Popover component.
type PopoverConf = tclayout.PopoverConf

// Dialog creates a dialog, closed until something opens it, and hands back a
// handle to open, close and fill it. The body is only computed while it is
// open, so handle what opens the dialog before calling DialogContainer.With.
var Dialog = tclayout.Dialog

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
var Empty = tclayout.Empty

// EmptyConf is the configuration for the Empty component.
type EmptyConf = tclayout.EmptyConf

// EmptySlot is what Empty hands out: a place in the page that can be written
// and written over.
type EmptySlot = tclayout.EmptySlot

// EmptyContainer is the old name of EmptySlot.
//
// Deprecated: use [EmptySlot].
type EmptyContainer = tclayout.EmptySlot
