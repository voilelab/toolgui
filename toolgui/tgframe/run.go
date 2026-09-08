package tgframe

import (
	"errors"

	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// ErrDuplicatedID is the error that two components in one run claim the same
// id. Components are placed by position, so duplicates of a component that has
// no id are fine; an id is a name for state, and two things cannot share one.
var ErrDuplicatedID = errors.New("duplicated component id")

// runState is the bookkeeping shared by every container of a single run.
type runState struct {
	ids map[string]bool

	// released holds the ids the run has taken back off the screen, the
	// contents of a slot it cleared. An id claimed again leaves the set; what
	// stays is state no component reads any more, and [App.Run] drops it.
	released map[string]bool

	err error
}

func newRunState() *runState {
	return &runState{ids: map[string]bool{}, released: map[string]bool{}}
}

// registerID claims comp's id for this run. The first collision is kept and
// reported by [App.Run]; the run itself carries on, so the page still renders
// and the user can see what collided.
func (r *runState) registerID(comp Component) {
	id := comp.GetID()
	if id == "" {
		return
	}

	delete(r.released, id)

	if r.ids[id] {
		if r.err == nil {
			r.err = tgutil.Errorf(
				"%w: `%s`. Two components cannot share an id; give one of them"+
					" its own through the component's Conf.ID.",
				ErrDuplicatedID, id)
		}
		return
	}

	r.ids[id] = true
}

// unregisterID gives comp's id back, so this run may claim it again. The state
// under it is released at the end of the run unless something claims it first:
// a widget that left the screen should not hand its old value to whatever
// lands on its id next run.
func (r *runState) unregisterID(comp Component) {
	id := comp.GetID()
	if id == "" {
		return
	}

	delete(r.ids, id)
	r.released[id] = true
}
