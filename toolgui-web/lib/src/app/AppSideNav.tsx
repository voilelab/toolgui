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
  // breakpoint, where the burger owns the toggle instead.
  collapsed: boolean
}

// The nav body both toggles point at.
const navBodyID = 'toolgui-nav-body'

// AppSideNav is the left column: the page list on top, the page's own sidebar
// below it, and the app controls at the bottom.
export class AppSideNav extends Component<AppSideNavProps, AppSideNavState> {
  constructor(props: AppSideNavProps) {
    super(props)
    this.state = {
      open: false,
      collapsed: false,
    }
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

    return <aside className={`toolgui-nav ${collapsed ? 'is-collapsed' : ''}`}>
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
        onClick={() => { this.setState((prev) => ({ collapsed: !prev.collapsed })) }}>
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
  }
}
