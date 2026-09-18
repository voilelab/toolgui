import React, { Component } from "react"
import { Button, Divider, Menu } from "@mantine/core"

import { MenuNode } from "./AppConf"
import { UpdateEvent } from "./UpdateEvent"

import '@toolgui-web/lib/src/assets/css/shell.css'

interface AppMenuBarProps {
  menu: MenuNode[]
  update: (e: UpdateEvent) => void
}

// hasItems reports whether a menu holds anything worth opening a dropdown for.
// A submenu of nothing but separators draws nothing, so it stays a plain
// label rather than an empty panel.
function hasItems(nodes: MenuNode[]): boolean {
  return nodes.some(node => node.type !== 'separator')
}

// AppMenuBar is the row above the app: one entry per top level node of the
// tree the App declared. It is only rendered for an app that declared one.
export class AppMenuBar extends Component<AppMenuBarProps> {
  // A menu item sends the same click event a Button does; the Go side tells
  // the two apart by the item id's reserved prefix.
  click(id: string) {
    this.props.update({ type: 'click', id })
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

  // One top level node. A submenu opens a dropdown; a text item is a button
  // that sends its click straight away, and a separator divides the row.
  topNode(node: MenuNode, key: string) {
    switch (node.type) {
      case 'separator':
        return <Divider key={key} orientation="vertical" my={4} />
      case 'text':
        return (
          <Button key={key} className="toolgui-menubar-button"
            id={node.id} variant="subtle" color="gray" size="compact-sm"
            onClick={() => { this.click(node.id) }}>
            {node.label}
          </Button>
        )
      case 'submenu': {
        const children = node.children || []
        if (!hasItems(children)) {
          return (
            <Button key={key} className="toolgui-menubar-button"
              variant="subtle" color="gray" size="compact-sm" disabled>
              {node.label}
            </Button>
          )
        }

        return (
          // trigger="click-hover" is what makes a menubar feel like one: the
          // first entry is opened with a click, and moving along the row then
          // opens the rest without clicking again.
          <Menu key={key} trigger="click-hover" position="bottom-start"
            openDelay={0} closeDelay={80} withinPortal>
            <Menu.Target>
              <Button className="toolgui-menubar-button"
                variant="subtle" color="gray" size="compact-sm">
                {node.label}
              </Button>
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
        {this.props.menu.map((node, i) => this.topNode(node, `${i}`))}
      </div>
    )
  }
}
