package tcmisc

import (
	"github.com/voilelab/toolgui/toolgui/tgframe"
	"github.com/voilelab/toolgui/toolgui/tgutil"
)

// frameValue reads the latest value the sandboxed frame with the given id sent
// through window.toolgui.update.
//
// It reads nil when there is no value, which a value that happens to be the
// zero T does not: before the frame's first update, and equally when the frame
// sent null, which is how it says it has none. A value that does not fit T is
// the run's failure rather than a silent zero.
func frameValue[T any](c *tgframe.Container, kind, id string) *T {
	var out *T

	if err := c.State.GetObject(id, &out); err != nil {
		c.Fail(tgutil.Errorf("%s %q sent a value that does not parse: %w",
			kind, id, err))
		return nil
	}

	return out
}
