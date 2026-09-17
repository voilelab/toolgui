import React, { Component } from "react";
import { ActionIcon, Burger, Button, Divider, Group, Loader, NavLink, Text } from "@mantine/core";
import { IconLayoutSidebarLeftCollapse, IconLayoutSidebarLeftExpand } from "@tabler/icons-react";

import { ThemeModeButton } from './ThemeModeButton';
import { AppConf } from "./AppConf";
import { Forest } from "./Nodes";
import { TComponent } from "../components/factory";
import { UpdateEvent } from "./UpdateEvent";
import { UploadFunc } from "./Upload";
import { ThemeMode } from "../util/theme";
import {
  NAV_DEFAULT_WIDTH, NAV_MAX_WIDTH, NAV_MIN_WIDTH, NAV_WIDTH_STEP,
  clampNavWidth, initialNavCollapsed, initialNavWidth,
  storeNavCollapsed, storeNavWidth,
} from "../util/sidenav";
import { emojize } from "../util/emoji";

import '@toolgui-web/lib/src/assets/css/shell.css'

interface AppSideNavProps {
  appConf: AppConf
  forest: Forest
  running: boolean
  pageFound: boolean
  pageName: string
  onNavigate?: (name: string) => void
  rerun: () => void
  update: (e: UpdateEvent) => void
  upload: UploadFunc
  themeMode: ThemeMode
}

interface AppSideNavState {
  // Whether the collapsed mobile menu is open. Ignored on wider screens,
  // where the burger is hidden.
  open: boolean
  // Whether the column is collapsed to its handle. Ignored below the mobile
  // breakpoint, where the burger owns the toggle instead. Outlives the page:
  // stored, and read back before the first paint.
  collapsed: boolean
  // The column's width in px, stored and read back the same way. Ignored
  // while collapsed and below the mobile breakpoint, where the CSS sizes the
  // column itself.
  width: number
  // Whether a drag is in flight, for the handle's own highlight.
  dragging: boolean
}

// The nav body both toggles point at.
const navBodyID = 'toolgui-nav-body'

// Put on <body> for the length of a drag, so the pointer cannot select the
// page it sweeps across.
const resizingClass = 'toolgui-resizing'

// AppSideNav is the left column: the page list on top, the page's own sidebar
// below it, and the app controls at the bottom.
export class AppSideNav extends Component<AppSideNavProps, AppSideNavState> {
  constructor(props: AppSideNavProps) {
    super(props)
    this.state = {
      open: false,
      collapsed: initialNavCollapsed(),
      width: initialNavWidth(),
      dragging: false,
    }
  }

  // Where the drag started. Moving the edge by the delta, rather than to the
  // pointer, keeps the column from jumping when the handle is grabbed off
  // centre.
  private dragOriginX = 0
  private dragOriginWidth = 0

  // Whether a drag is in flight, as a field rather than off state: down,
  // move and up can all land in one tick, and a state update would not have
  // flushed in time for the next one. state.dragging only drives the class.
  private dragging = false

  componentWillUnmount() {
    // A drag interrupted by a page change would otherwise leave the whole
    // document unselectable.
    document.body.classList.remove(resizingClass)
  }

  // store is off mid-drag: that is hundreds of events, and only where the
  // edge comes to rest is worth keeping.
  setWidth(px: number, store: boolean) {
    const width = clampNavWidth(px)
    this.setState({ width })

    if (store) {
      storeNavWidth(width)
    }
  }

  // The width the edge is at for this pointer position.
  widthAt(e: React.PointerEvent<HTMLDivElement>): number {
    return clampNavWidth(this.dragOriginWidth + e.clientX - this.dragOriginX)
  }

  startDrag(e: React.PointerEvent<HTMLDivElement>) {
    if (this.dragging) {
      return
    }

    // Capture keeps the moves coming to the handle once the pointer outruns
    // its 6px, and up still lands here if it happens outside the window.
    // Synthesized events have no live pointer to capture, and throw.
    try {
      e.currentTarget.setPointerCapture(e.pointerId)
    } catch (err) {
      // Without capture the drag still works, just not past the window edge.
    }

    this.dragOriginX = e.clientX
    this.dragOriginWidth = this.state.width
    this.dragging = true
    document.body.classList.add(resizingClass)
    this.setState({ dragging: true })
  }

  drag(e: React.PointerEvent<HTMLDivElement>) {
    if (!this.dragging) {
      return
    }

    this.setWidth(this.widthAt(e), false)
  }

  endDrag(e: React.PointerEvent<HTMLDivElement>) {
    if (!this.dragging) {
      return
    }

    this.dragging = false
    // No release: the browser drops the capture itself once pointerup or
    // pointercancel has been dispatched.
    document.body.classList.remove(resizingClass)
    this.setState({ dragging: false })
    // Off the event rather than off state, which the last move may not have
    // flushed into yet.
    this.setWidth(this.widthAt(e), true)
  }

  // The handle is a real separator, so it answers the keys one is expected
  // to: the arrows step, Home and End go to the bounds, Enter resets -- the
  // same thing a double click does for a pointer.
  resizeKey(e: React.KeyboardEvent<HTMLDivElement>) {
    let width = this.state.width

    switch (e.key) {
      case 'ArrowLeft':
        width -= NAV_WIDTH_STEP
        break
      case 'ArrowRight':
        width += NAV_WIDTH_STEP
        break
      case 'Home':
        width = NAV_MIN_WIDTH
        break
      case 'End':
        width = NAV_MAX_WIDTH
        break
      case 'Enter':
        width = NAV_DEFAULT_WIDTH
        break
      default:
        return
    }

    e.preventDefault()
    this.setWidth(width, true)
  }

  toggleCollapsed() {
    const collapsed = !this.state.collapsed
    storeNavCollapsed(collapsed)
    this.setState({ collapsed })
  }

  jumpToPage(name: string) {
    this.setState({ open: false })

    if (this.props.onNavigate) {
      this.props.onNavigate(name)
      return
    }

    if (this.props.appConf.hash_page_name_mode) {
      window.location.href = '#/' + name
      window.location.reload();
    } else {
      window.location.href = '/' + name
    }
  }

  // pageHref keeps the anchors real links. A transport with its own
  // navigation has no URL to point at, so it gets the hash form; the click
  // handler cancels the default either way.
  pageHref(name: string) {
    if (this.props.onNavigate || this.props.appConf.hash_page_name_mode) {
      return '#/' + name
    }

    return '/' + name
  }

  sidebarNode() {
    return this.props.forest.nodes[this.props.appConf.sidebar_container_id]
  }

  render() {
    const sidebarNode = this.sidebarNode()
    const hasSidebar = sidebarNode.children.length > 0

    const collapsed = this.state.collapsed

    return <>
      <aside className={`toolgui-nav ${collapsed ? 'is-collapsed' : ''}`}
        style={{ '--tg-nav-width': `${this.state.width}px` } as React.CSSProperties}>
        <Burger className="toolgui-nav-burger"
          aria-label="menu"
          aria-controls={navBodyID}
          opened={this.state.open}
          onClick={() => { this.setState((prev) => ({ open: !prev.open })) }} />

        {/* Stays in the column when collapsed, so the strip keeps a handle. */}
        <ActionIcon className="toolgui-nav-collapse"
          variant="subtle" size="lg"
          aria-label={collapsed ? 'Expand the side column' : 'Collapse the side column'}
          aria-expanded={!collapsed}
          aria-controls={navBodyID}
          onClick={() => { this.toggleCollapsed() }}>
          {collapsed ?
            <IconLayoutSidebarLeftExpand size={18} /> :
            <IconLayoutSidebarLeftCollapse size={18} />}
        </ActionIcon>

        <div id={navBodyID}
          className={`toolgui-nav-body ${this.state.open ? 'is-open' : ''}`}>
          <nav className="toolgui-nav-list" aria-label="main navigation">
            {
              this.props.appConf.page_names.map(name => {
                const active = name === this.props.pageName
                return (
                  <NavLink key={name}
                    component="a"
                    href={this.pageHref(name)}
                    active={active}
                    variant="filled"
                    aria-current={active ? 'page' : undefined}
                    label={<>
                      {emojize(this.props.appConf.page_confs[name].emoji || '')}
                      {this.props.appConf.page_confs[name].title}
                    </>}
                    onClick={(e) => { e.preventDefault(); this.jumpToPage(name) }} />
                )
              })
            }
          </nav>

          {hasSidebar ?
            <div>
              <Divider my="sm" />
              <TComponent node={sidebarNode}
                update={(e) => { this.props.update(e) }}
                upload={async (f, id) => await this.props.upload(f, id)}
                theme={this.props.themeMode} />
            </div> : ''}

          <Group className="toolgui-nav-foot" gap="xs">
            {this.props.running ? <Loader size="sm" /> : ''}
            {this.props.pageFound ?
              <Button variant="default" onClick={() => { this.props.rerun() }}>
                Rerun
              </Button> : ''}
            <ThemeModeButton />
          </Group>

          {this.props.appConf.show_version ?
            <Text className="toolgui-nav-version" size="xs" c="dimmed">
              toolgui {this.props.appConf.version}
            </Text> : ''}
        </div>
      </aside>

      {/* A sibling of the column, not a child: the column scrolls its own
          overflow, and a handle inside it would scroll away with the page
          list. The CSS hides it wherever the width is not the column's to
          keep -- collapsed, or on a narrow viewport. */}
      <div className={`toolgui-nav-resizer ${this.state.dragging ? 'is-dragging' : ''}`}
        role="separator"
        aria-orientation="vertical"
        aria-label="Resize the side column"
        aria-controls={navBodyID}
        aria-valuenow={this.state.width}
        aria-valuemin={NAV_MIN_WIDTH}
        aria-valuemax={NAV_MAX_WIDTH}
        tabIndex={0}
        onPointerDown={(e) => { this.startDrag(e) }}
        onPointerMove={(e) => { this.drag(e) }}
        onPointerUp={(e) => { this.endDrag(e) }}
        onPointerCancel={(e) => { this.endDrag(e) }}
        onDoubleClick={() => { this.setWidth(NAV_DEFAULT_WIDTH, true) }}
        onKeyDown={(e) => { this.resizeKey(e) }} />
    </>
  }
}
