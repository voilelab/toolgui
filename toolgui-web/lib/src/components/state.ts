export var stateValues: { [key: string]: any } = {}

// stateGeneration rises every time the client's state is cleared, which is a
// new server session. Anything a component holds outside stateValues -- a
// form's queue of unsent events -- reads it to tell its own values from ones
// it collected against a session that is gone.
export var stateGeneration = 0

export function clearState() {
    stateValues = {}
    stateGeneration++
}
