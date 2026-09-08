package tcmisc

import (
	"github.com/voilelab/toolgui/toolgui/tgcomp/tccontent"
	"github.com/voilelab/toolgui/toolgui/tgcomp/tclayout"
	"github.com/voilelab/toolgui/toolgui/tgframe"
)

// statusState is how the work a status reports is going. It is what the icon
// in front of the label says.
type statusState string

const (
	statusRunning  statusState = "⏳"
	statusComplete statusState = "✅"
	statusError    statusState = "❌"
)

// StatusConf is the configuration for the Status component.
type StatusConf struct {
	tgframe.Base

	// Expanded is whether the status starts open. It is closed by default:
	// the label says how the work is going, and the lines are the detail.
	Expanded bool
}

// StatusHandle is what [Status] hands out. It is an expander whose label
// carries the state of the work, and whose contents are the lines written to
// it so far.
//
// It holds an [tclayout.Empty] slot and rewrites it on every change, so the
// label and the lines are always what the last call left.
type StatusHandle struct {
	slot     *tclayout.EmptySlot
	id       string
	label    string
	expanded bool
	state    statusState
	lines    []string
}

// StatusContainer is the old name of [StatusHandle].
//
// Deprecated: use [StatusHandle]. A status is a handle on work being
// reported, not a container components can be added to.
type StatusContainer = StatusHandle

// Status reports a piece of work while the page function does it. Write the
// lines as they come, and close it with [StatusHandle.Complete] or
// [StatusHandle.Fail]:
//
//	s := tgcomp.Status(c, "Importing…")
//	for _, f := range files {
//		s.Write(f)
//		importFile(f)
//	}
//	s.Complete("Imported")
//
// A status that is never closed stays in its running state, which is what the
// page should show when the work did not get that far.
func Status(c *tgframe.Container, label string, conf ...*StatusConf) *StatusHandle {
	cf := tgframe.OneConf("Status", conf)

	// The status is built out of an expander, which is rewritten on every
	// change. The id has to outlive those rewrites, so it comes from the label
	// rather than from the label plus the icon in front of it.
	id := cf.ID
	if id == "" {
		id = label
	}

	s := &StatusHandle{
		slot:     tclayout.Empty(c, &tclayout.EmptyConf{ID: cf.ID}),
		id:       id,
		label:    label,
		expanded: cf.Expanded,
		state:    statusRunning,
	}

	s.render()
	return s
}

// Write appends a line to the status.
func (s *StatusHandle) Write(text string) {
	s.lines = append(s.lines, text)
	s.render()
}

// Update replaces the label, leaving the state and the lines alone.
func (s *StatusHandle) Update(label string) {
	s.label = label
	s.render()
}

// Complete closes the status as a success, taking a new label at most one.
// Two labels or more is a mistake rather than something to join, and panics,
// as passing two confs to a component does.
func (s *StatusHandle) Complete(label ...string) {
	s.finish(statusComplete, label)
}

// Fail closes the status as a failure, taking a new label at most one.
func (s *StatusHandle) Fail(label ...string) {
	s.finish(statusError, label)
}

// Error closes the status as a failure, taking a new label at most one.
//
// Deprecated: use [StatusHandle.Fail]. This name reads like the [error]
// interface, which a status does not implement.
func (s *StatusHandle) Error(label ...string) {
	s.finish(statusError, label)
}

func (s *StatusHandle) finish(state statusState, label []string) {
	if len(label) > 1 {
		panic("toolgui: Status takes at most one closing label")
	}

	if len(label) == 1 {
		s.label = label[0]
	}

	s.state = state
	s.render()
}

// render draws the status as it now stands, over what was there before.
func (s *StatusHandle) render() {
	s.slot.With(func(c *tgframe.Container) {
		inner := tclayout.Expand(c, string(s.state)+" "+s.label, s.expanded,
			&tclayout.ExpandConf{ID: s.id})

		for _, line := range s.lines {
			tccontent.Text(inner, line)
		}
	})
}
