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

// syncResetKey drops the client's value for id when resetKey differs from the
// one last seen, so the input falls back to its default. Idempotent.
export function syncResetKey(id: string, resetKey: string | undefined) {
    const keyID = id + '#reset_key'
    const key = resetKey ?? ''
    if (keyID in stateValues && stateValues[keyID] !== key) {
        delete stateValues[id]
    }
    stateValues[keyID] = key
}
