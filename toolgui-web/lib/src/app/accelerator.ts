// A menu item's accelerator: the key combination that fires it without the
// menu being opened.
//
// On the desktop the OS dispatches it off the native menu item and none of
// this runs. In a browser nothing dispatches it, so the shell listens for the
// keystroke itself -- which is what this parses the declaration for.
//
// The spelling is the one the Go side normalized to: modifiers in a fixed
// order, then the key, joined by `+`. Everything here is a pure function of
// that string; the listening lives in AppMenuBar.

// Accelerator is a parsed declaration. The four modifiers are all named,
// present or not, because matching an event means checking the absent ones
// too: Ctrl+Shift+O must not fire the item that asked for Ctrl+O.
export interface Accelerator {
  cmdOrCtrl: boolean
  ctrl: boolean
  optionOrAlt: boolean
  shift: boolean

  // The key, lowercased: one printable ASCII character, or a named key.
  key: string
}

// namedKeys maps the names the declaration may use to the `KeyboardEvent.key`
// they arrive as. Anything not in here is a one character key, which arrives
// as itself.
const namedKeys: { [name: string]: string } = {
  'backspace': 'Backspace',
  'tab': 'Tab',
  'enter': 'Enter',
  'escape': 'Escape',
  'left': 'ArrowLeft',
  'right': 'ArrowRight',
  'up': 'ArrowUp',
  'down': 'ArrowDown',
  'space': ' ',
  'delete': 'Delete',
  'home': 'Home',
  'end': 'End',
  'page up': 'PageUp',
  'page down': 'PageDown',
  'plus': '+',
}

// isMac says which key CmdOrCtrl and OptionOrAlt stand for here. The platform
// APIs that answer this are all deprecated or lying, so this reads the least
// wrong of them and treats iPadOS -- which reports itself as a Mac -- as one
// too, since it carries a Mac's keyboard when it carries one at all.
export function isMac(): boolean {
  const nav = typeof navigator === 'undefined' ? null : navigator
  if (!nav) {
    return false
  }

  const platform = (nav as any).userAgentData?.platform || nav.platform || ''
  return /mac/i.test(platform)
}

// parseAccelerator reads a declaration. It returns null for one it does not
// understand, which the Go side should already have turned away -- a frontend
// newer or older than the app it is drawing is the case left.
export function parseAccelerator(accel: string): Accelerator | null {
  const parts = accel.split('+')
  const key = (parts.pop() || '').toLowerCase()
  if (!key) {
    return null
  }

  const out: Accelerator = {
    cmdOrCtrl: false, ctrl: false, optionOrAlt: false, shift: false, key,
  }

  for (const part of parts) {
    switch (part.toLowerCase()) {
      case 'cmdorctrl': out.cmdOrCtrl = true; break
      case 'ctrl': out.ctrl = true; break
      case 'optionoralt': out.optionOrAlt = true; break
      case 'shift': out.shift = true; break
      default: return null
    }
  }

  return out
}

// hasModifier reports whether the accelerator asks for a modifier that is not
// Shift. Shift alone does not count: Shift+A is typing, which is why an
// accelerator wearing no more than it stays out of the way of a text field.
export function hasModifier(accel: Accelerator): boolean {
  return accel.cmdOrCtrl || accel.ctrl || accel.optionOrAlt
}

// matchesKey reports whether the event's key is the accelerator's.
//
// `key` is the character the layout produced, so it is what a named key and a
// plain letter both arrive as. It is also what Option turns into something
// else entirely on a Mac -- Option+O is `ø` -- so `code`, which names the
// physical key, is taken as well for the letters and digits it covers.
function matchesKey(accel: Accelerator, e: KeyboardEvent): boolean {
  const want = namedKeys[accel.key]
  if (want !== undefined) {
    return e.key === want
  }

  if (e.key.toLowerCase() === accel.key) {
    return true
  }

  if (accel.key >= 'a' && accel.key <= 'z') {
    return e.code === 'Key' + accel.key.toUpperCase()
  }

  if (accel.key >= '0' && accel.key <= '9') {
    return e.code === 'Digit' + accel.key
  }

  return false
}

// matchesEvent reports whether e is the keystroke accel declared. Every
// modifier is checked, so a combination carrying one the item did not ask for
// belongs to whatever asked for that one instead.
export function matchesEvent(accel: Accelerator, e: KeyboardEvent,
  mac: boolean): boolean {

  // CmdOrCtrl is the platform's own menu modifier: Command on a Mac, Control
  // everywhere else. Ctrl is Control on all of them, which on a Mac makes the
  // two different keys -- and the Go side turns away an item asking for both,
  // since off a Mac they are one.
  const meta = mac && accel.cmdOrCtrl
  const control = accel.ctrl || (!mac && accel.cmdOrCtrl)

  return e.metaKey === meta && e.ctrlKey === control &&
    e.altKey === accel.optionOrAlt && e.shiftKey === accel.shift &&
    matchesKey(accel, e)
}

// keyLabels are the names a key is shown under, where that is not the key
// itself upper cased.
const keyLabels: { [name: string]: string } = {
  'backspace': 'Backspace',
  'tab': 'Tab',
  'enter': 'Enter',
  'escape': 'Esc',
  'left': '←',
  'right': '→',
  'up': '↑',
  'down': '↓',
  'space': 'Space',
  'delete': 'Del',
  'home': 'Home',
  'end': 'End',
  'page up': 'PgUp',
  'page down': 'PgDn',
  'plus': '+',
}

// formatAccelerator is what the item shows next to its label: the symbols a
// Mac menu uses, in the order a Mac menu writes them, and the spelled out
// names everywhere else.
export function formatAccelerator(accel: Accelerator, mac: boolean): string {
  const key = keyLabels[accel.key] || accel.key.toUpperCase()

  if (mac) {
    return (accel.ctrl ? '⌃' : '') + (accel.optionOrAlt ? '⌥' : '') +
      (accel.shift ? '⇧' : '') + (accel.cmdOrCtrl ? '⌘' : '') + key
  }

  const parts = []
  if (accel.cmdOrCtrl || accel.ctrl) parts.push('Ctrl')
  if (accel.optionOrAlt) parts.push('Alt')
  if (accel.shift) parts.push('Shift')
  parts.push(key)

  return parts.join('+')
}
