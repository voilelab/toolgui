import React, { Component } from "react"
import { Button, Divider, Menu } from "@mantine/core"

import { MenuNode } from "./AppConf"
import { UpdateEvent } from "./UpdateEvent"

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

// hasItems reports whether a menu holds anything worth opening a dropdown for.
// A submenu of nothing but separators draws nothing, so it stays a plain
// label rather than an empty panel.
function hasItems(nodes: MenuNode[]): boolean {
  return nodes.some(node => node.type !== 'separator')
}

// AppMenuBar is the row above the app: one entry per top level node of the
// tree the App declared. It is only rendered for an app that declared one.
export class AppMenuBar extends Component<AppMenuBarProps, AppMenuBarState> {
  constructor(props: AppMenuBarProps) {
    super(props)
    this.state = { open: null }
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
