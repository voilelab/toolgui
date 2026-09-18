import React, { Component } from "react"
import { Button, Divider, Menu } from "@mantine/core"

import { MenuNode } from "./AppConf"
import { UpdateEvent } from "./UpdateEvent"
import {
  Accelerator, formatAccelerator, hasModifier, isMac, matchesEvent,
  parseAccelerator,
} from "./accelerator"

import '@toolgui-web/lib/src/assets/css/shell.css'

interface AppMenuBarProps {
  menu: MenuNode[]
  update: (e: UpdateEvent) => void
}

interface AppMenuBarState {
  // Which top level entry is open, by index, and null when none is. One
  // number for the whole row rather than a flag per entry: a menubar has at
  // most one dropdown open, and it is also what says the row is armed.
  open: number | null
}

// Bound is one item's accelerator, with the click id it fires.
interface Bound {
  accel: Accelerator
  id: string
}

// bindings walks the tree for every accelerator declared in it. An app that
// declared none gets an empty list, which is what keeps the listener off the
// document entirely.
function bindings(nodes: MenuNode[]): Bound[] {
  const out: Bound[] = []

  for (const node of nodes) {
    if (node.type === 'submenu') {
      out.push(...bindings(node.children || []))
      continue
    }

    if (node.type !== 'text' || !node.accelerator) {
      continue
    }

    const accel = parseAccelerator(node.accelerator)
    if (accel) {
      out.push({ accel, id: node.id })
    }
  }

  return out
}

// isAltGraph reports whether the keystroke carries AltGr, the modifier that
// puts a third character on a key.
//
// Windows, and some layouts elsewhere, report AltGr as Control and Alt held
// together, so a Ctrl+OptionOrAlt chord and the character AltGr+that key
// produces arrive as the same event. There is no telling them apart, so the
// character wins: firing the item would eat a keystroke the visitor meant to
// type, and the item is still in the menu, while a shortcut that silently
// swallows text is not something to hand a visitor on a German keyboard.
function isAltGraph(e: KeyboardEvent): boolean {
  return typeof e.getModifierState === 'function' &&
    e.getModifierState('AltGraph')
}

// isEditable reports whether the keystroke landed in something the visitor is
// typing into. Mantine's own inputs are all one of these three.
function isEditable(target: EventTarget | null): boolean {
  const el = target as HTMLElement | null
  if (!el || !el.tagName) {
    return false
  }

  const tag = el.tagName.toLowerCase()
  return tag === 'input' || tag === 'textarea' || tag === 'select' ||
    el.isContentEditable
}

// hasItems reports whether a menu holds anything worth opening a dropdown for.
// A submenu of nothing but separators draws nothing, so it stays a plain
// label rather than an empty panel.
function hasItems(nodes: MenuNode[]): boolean {
  return nodes.some(node => node.type !== 'separator')
}

// AppMenuBar is the row above the app: one entry per top level node of the
// tree the App declared. It is only rendered for an app that declared one.
export class AppMenuBar extends Component<AppMenuBarProps, AppMenuBarState> {
  // The accelerators of the tree this bar was given, and which key
  // CmdOrCtrl stands for here. Both are read once: the tree is the app's and
  // does not change between runs, and neither does the keyboard.
  private bindings: Bound[]
  private mac: boolean

  private onKeyDown = (e: KeyboardEvent) => { this.keyDown(e) }

  constructor(props: AppMenuBarProps) {
    super(props)
    this.state = { open: null }
    this.bindings = bindings(props.menu)
    this.mac = isMac()
  }

  // Nothing is listened for on behalf of a menu that declared no accelerator:
  // the listener is only added once there is something for it to match.
  componentDidMount() {
    if (this.bindings.length > 0) {
      document.addEventListener('keydown', this.onKeyDown)
    }
  }

  componentWillUnmount() {
    document.removeEventListener('keydown', this.onKeyDown)
  }

  // The browser's half of the feature. On the desktop the OS dispatches an
  // accelerator off the native menu item and none of this runs.
  //
  // A hit is preventDefault'd, which is all a page can do about a combination
  // the browser has already taken -- and for some of them it is not enough.
  // The menu documentation says which ones are not reliably the app's.
  keyDown(e: KeyboardEvent) {
    // A keystroke the visitor is still composing is not a keystroke yet: an
    // IME reports the whole composition as one keydown of its own.
    if (e.isComposing || e.repeat || isAltGraph(e)) {
      return
    }

    const typing = isEditable(e.target)

    for (const bound of this.bindings) {
      // In a text field, only a real chord is an accelerator. A bare key --
      // and Shift plus one, which is the same key shifted -- is what the
      // visitor is typing.
      if (typing && !hasModifier(bound.accel)) {
        continue
      }

      if (matchesEvent(bound.accel, e, this.mac)) {
        e.preventDefault()
        this.click(bound.id)
        return
      }
    }
  }

  // A menu item sends the same click event a Button does; the Go side tells
  // the two apart by the item id's reserved prefix.
  click(id: string) {
    this.setState({ open: null })
    this.props.update({ type: 'click', id })
  }

  // Moving along the row once something is open moves the dropdown with the
  // pointer, which is what a menubar does. Before anything is open it does
  // nothing: a pointer crossing the row on its way elsewhere should not pop
  // a menu open.
  hover(at: number | null) {
    if (this.state.open === null) {
      return
    }

    this.setState({ open: at })
  }

  // The contents of a dropdown. Nested submenus go through Menu.Sub, whose
  // dropdown has to stay a DOM descendant of the one it opened from.
  dropdownItems(nodes: MenuNode[], path: string) {
    return nodes.map((node, i) => {
      const key = `${path}/${i}`

      switch (node.type) {
        case 'separator':
          return <Menu.Divider key={key} />
        case 'text':
          return (
            <Menu.Item key={key} id={node.id}
              rightSection={this.accelLabel(node.accelerator)}
              onClick={() => { this.click(node.id) }}>
              {node.label}
            </Menu.Item>
          )
        case 'submenu': {
          const children = node.children || []
          if (!hasItems(children)) {
            return <Menu.Item key={key} disabled>{node.label}</Menu.Item>
          }

          return (
            <Menu.Sub key={key}>
              <Menu.Sub.Target>
                <Menu.Sub.Item>{node.label}</Menu.Sub.Item>
              </Menu.Sub.Target>
              <Menu.Sub.Dropdown>
                {this.dropdownItems(children, key)}
              </Menu.Sub.Dropdown>
            </Menu.Sub>
          )
        }
      }
    })
  }

  // What an item shows next to its label, and nothing for an item that
  // declared no accelerator -- or one this frontend does not understand,
  // which is nothing to draw a gap for.
  accelLabel(accel?: string) {
    const parsed = accel ? parseAccelerator(accel) : null
    if (!parsed) {
      return undefined
    }

    return (
      <span className="toolgui-menubar-accel">
        {formatAccelerator(parsed, this.mac)}
      </span>
    )
  }

  // The button every top level entry wears, submenu or not.
  entryButton(label: string, at: number, disabled?: boolean) {
    return (
      <Button className="toolgui-menubar-button"
        variant="subtle" color="gray" size="compact-sm"
        disabled={disabled}
        onMouseEnter={() => { this.hover(at) }}>
        {label}
      </Button>
    )
  }

  // One top level node. A submenu opens a dropdown; a text item is a button
  // that sends its click straight away, and a separator divides the row.
  topNode(node: MenuNode, at: number) {
    const key = `${at}`

    switch (node.type) {
      case 'separator':
        return <Divider key={key} orientation="vertical" my={4} />
      case 'text':
        return (
          // Crossing a plain entry closes whatever was open: it has no
          // dropdown to move to, and leaving the last one up would leave a
          // dropdown belonging to an entry the pointer has left.
          <Button key={key} className="toolgui-menubar-button"
            id={node.id} variant="subtle" color="gray" size="compact-sm"
            onMouseEnter={() => { this.hover(null) }}
            onClick={() => { this.click(node.id) }}>
            {node.label}
          </Button>
        )
      case 'submenu': {
        const children = node.children || []
        if (!hasItems(children)) {
          return <React.Fragment key={key}>
            {this.entryButton(node.label, at, true)}
          </React.Fragment>
        }

        return (
          // Controlled, and one entry at a time: left to their own state the
          // dropdowns do not know about each other, and the one a click
          // opened would stay up while hovering opened the next.
          <Menu key={key} trigger="click" position="bottom-start"
            opened={this.state.open === at}
            onChange={(opened) => { this.setState({ open: opened ? at : null }) }}
            withinPortal>
            <Menu.Target>
              {this.entryButton(node.label, at)}
            </Menu.Target>
            <Menu.Dropdown>
              {this.dropdownItems(children, key)}
            </Menu.Dropdown>
          </Menu>
        )
      }
    }
  }

  render() {
    return (
      // A row of buttons, each of which says what it is: no role="menubar",
      // which would promise roving arrow keys and menuitem children that
      // Mantine's own dropdowns do not render.
      <div className="toolgui-menubar">
        {this.props.menu.map((node, i) => this.topNode(node, i))}
      </div>
    )
  }
}
