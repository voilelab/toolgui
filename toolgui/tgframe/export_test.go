package tgframe

import "context"

// RunWithHandlingPanic is [App.Run] with panics recovered as a session does,
// for the external tests.
func RunWithHandlingPanic(app *App, name string, state *State, notifyFunc SendNotifyPackFunc) error {
	return app.runContextWithHandlingPanic(context.Background(), name, nil, state, notifyFunc, nil)
}
