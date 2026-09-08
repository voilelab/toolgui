export interface ClickEvent {
  type: "click"
  id: string
}

export interface InputEvent {
  type: "input"
  id: string
  value: any
}

// A SelectEvent carries either the one value a select or a radio has, or the
// values a multiselect has. The two are separate fields so that a server
// built before multi-selection existed still reads the single one.
export type SelectEvent = {
  type: "select"
  id: string
} & ({ value: number } | { values: number[] })

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
