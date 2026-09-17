import React from "react"

// FormSubmitContext carries the enclosing form's submit, so a component can
// send the form it is written in without knowing where in the tree it sits.
// It is null outside a form.
//
// A form already reaches its children through the update it hands them, but
// that channel cannot say which of them asked: a container in between passes
// it along unchanged, and every component that reports a press sends the same
// click event. Only a Button means "send this form", so only a Button reads
// this.
export const FormSubmitContext = React.createContext<(() => void) | null>(null)
