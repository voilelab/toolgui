package tgframe

// ErrorComponentName is the typename of the placeholder a failed component
// leaves behind.
var ErrorComponentName = "error_component"

var _ Component = &errorComponent{}

// errorComponent is what the page shows where a component could not be drawn.
// The message is the same one [App.Run] returns, so the app user reads on
// screen what the log line says.
type errorComponent struct {
	*BaseComponent
	Message string `json:"message"`
}

func newErrorComponent(message string) *errorComponent {
	return &errorComponent{
		BaseComponent: &BaseComponent{
			Name: ErrorComponentName,
		},
		Message: message,
	}
}

// Fail reports err as the failure of one component. The run records it and
// [App.Run] returns the first one, but the page function is not interrupted:
// everything after this still renders, and a visible error placeholder takes
// the failed component's place so the failure is on screen and not only in the
// server log.
//
// It is for a failure the run's data decides — a table whose rows do not match
// its head, an image that will not encode, a stored value that no longer
// parses. A call that cannot be right whatever the data is, such as
// Column(c, 0), still panics.
//
//	if len(row) != len(head) {
//		c.Fail(tgutil.Errorf("row %d does not match the head", i))
//		return
//	}
//
// A nil err is nothing to report and does nothing.
func (c *Container) Fail(err error) {
	if err == nil {
		return
	}

	if c.run != nil {
		c.run.fail(err)
	}

	c.AddComponent(newErrorComponent(err.Error()))
}
