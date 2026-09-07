export interface ClickEvent {
  type: "click"
  id: string
}

export interface InputEvent {
  type: "input"
  id: string
  value: any
}

export interface SelectEvent {
  type: "select"
  id: string
  value: number
}

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
