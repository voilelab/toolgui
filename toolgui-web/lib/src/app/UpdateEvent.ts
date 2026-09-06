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

// IframeEvent carries an arbitrary value from inside an iframe component.
// The id is set by the host to the iframe's own id.
export interface IframeEvent {
  type: "iframe"
  id: string
  value: any
}

export interface FormEvent {
  type: "form"
  events: UpdateEvent[]
}

export type UpdateEvent = ClickEvent | InputEvent | SelectEvent | IframeEvent | FormEvent
