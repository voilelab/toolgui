export interface ClickEvent {
  type: "click"
  id: string
}

export interface InputEvent {
  type: "input"
  id: string
  value: any
}

// A SelectEvent carries the one value a select or a radio has, the values a
// multiselect has, or the keys a component that remembers what is picked by
// name has -- a DataFrame given row keys. Three separate fields so that a
// server built before either of the later two existed still reads the single
// value; a payload only ever carries one of them.
export type SelectEvent = {
  type: "select"
  id: string
} & ({ value: number } | { values: number[] } | { keys: string[] })

// CustomUpdateEvent carries an arbitrary value from a component that renders
// itself, an iframe or a plugin. The id is set by the host to the component's
// own id, so such a component can only write to its own state.
export interface CustomUpdateEvent {
  type: "custom"
  id: string
  value: any
}

export interface FormEvent {
  type: "form"
  events: UpdateEvent[]
}

export type UpdateEvent = ClickEvent | InputEvent | SelectEvent | CustomUpdateEvent | FormEvent
