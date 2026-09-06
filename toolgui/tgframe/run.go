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
	err error
}

func newRunState() *runState {
	return &runState{ids: map[string]bool{}}
}

// registerID claims comp's id for this run. The first collision is kept and
// reported by [App.Run]; the run itself carries on, so the page still renders
// and the user can see what collided.
func (r *runState) registerID(comp Component) {
	id := comp.GetID()
	if id == "" {
		return
	}

	if r.ids[id] {
		if r.err == nil {
			r.err = tgutil.Errorf(
				"%w: `%s`. Two components cannot share an id; give one of them"+
					" its own with the component's WithID variant or Conf.ID.",
				ErrDuplicatedID, id)
		}
		return
	}

	r.ids[id] = true
}
