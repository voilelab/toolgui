import React, { Component } from "react";

import { ThemeModeButton } from './ThemeModeButton';
import { AppConf } from "./AppConf";
import { Forest } from "./Nodes";
import { TComponent } from "../components/factory";
import { UpdateEvent } from "./UpdateEvent";
import { UploadFunc } from "./Upload";

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
  darkMode: string
  onChange: (darkMode: string) => void
}

interface AppSideNavState {
  // Whether the collapsed mobile menu is open. Ignored on wider screens,
  // where the column is always shown.
  open: boolean
}

// AppSideNav is the left column: the page list on top, the page's own sidebar
// below it, and the app controls at the bottom.
export class AppSideNav extends Component<AppSideNavProps, AppSideNavState> {
  constructor(props: AppSideNavProps) {
    super(props)
    this.state = {
      open: false,
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

    return <aside className="toolgui-nav">
      <button className="button is-ghost toolgui-nav-burger"
        aria-label="menu"
        aria-expanded={this.state.open}
        onClick={() => { this.setState((prev) => ({ open: !prev.open })) }}>
        <span className="icon">
          <i className="fas fa-bars"></i>
        </span>
        <span>Menu</span>
      </button>

      <div className={`toolgui-nav-body ${this.state.open ? 'is-open' : ''}`}>
        <nav className="menu" aria-label="main navigation">
          <ul className="menu-list">
            {
              this.props.appConf.page_names.map(name =>
                <li key={name}>
                  <a className={name === this.props.pageName ? 'is-active' : ''}
                    href={this.pageHref(name)}
                    onClick={(e) => { e.preventDefault(); this.jumpToPage(name) }}>
                    {this.props.appConf.page_confs[name].emoji}
                    {this.props.appConf.page_confs[name].title}
                  </a>
                </li>
              )
            }
          </ul>
        </nav>

        {hasSidebar ?
          <div>
            <hr />
            <TComponent node={sidebarNode}
              update={(e) => { this.props.update(e) }}
              upload={async (f) => await this.props.upload(f)}
              theme={this.props.darkMode} />
          </div> : ''}

        <div className="toolgui-nav-foot buttons">
          {this.props.running ?
            <span className="icon">
              <i className="fas fa-spinner fa-pulse"></i>
            </span> : ''}
          {this.props.pageFound ?
            <button className="button" onClick={() => { this.props.rerun() }}>
              Rerun
            </button> : ''}
          <ThemeModeButton onChange={(darkMode) => {
            this.props.onChange(darkMode)
          }} />
        </div>

        {this.props.appConf.show_version ?
          <p className="toolgui-nav-version">
            toolgui {this.props.appConf.version}
          </p> : ''}
      </div>
    </aside>
  }
}
