import React, { Component } from "react"

import { App, AppConf, dispatchPack, UpdateEvent } from "@toolgui-web/lib"
import { Backend } from "./api/backend"

interface WasmAppState {
  appConf: AppConf | null
  pageName: string
  error: string | null
}

// pageNameFromHash reads the page off the URL. A static host cannot route
// paths, so the hash is the only place a page name can live.
function pageNameFromHash(appConf: AppConf): string {
  if (window.location.hash.startsWith('#/')) {
    return window.location.hash.substring(2)
  }

  return appConf.page_names.length > 0 ? appConf.page_names[0] : ''
}

export class WasmApp extends Component<{}, WasmAppState> {
  appEle: React.RefObject<App>
  backend: Backend

  constructor(props: {}) {
    super(props)
    this.state = {
      appConf: null,
      pageName: '',
      error: null,
    }
    this.appEle = React.createRef()

    // Listening before the first run: packs delivered with no listener are
    // lost.
    this.backend = new Backend((pack) => {
      const app = this.appEle.current
      if (!app) {
        return
      }

      dispatchPack(app, pack)
    })

    this.setup().catch((e) => { this.fail(e) })
  }

  async setup() {
    const appConf = await this.backend.appConf()

    // The wasm program stays loaded across pages, so navigation is a new
    // session rather than a page load.
    window.addEventListener('hashchange', () => {
      this.openPage(appConf, pageNameFromHash(appConf))
    })

    this.openPage(appConf, pageNameFromHash(appConf))
  }

  // openPage renders the page and asks the backend for a session on it. start
  // runs after the commit, so the ref the pack listener needs is set.
  openPage(appConf: AppConf, pageName: string) {
    this.appEle.current?.clearState()

    this.setState({ appConf, pageName }, () => {
      this.backend.start(pageName).catch((e) => { this.fail(e) })
    })
  }

  jumpToPage(pageName: string) {
    if (this.state.appConf.hash_page_name_mode) {
      // Let the URL drive, so a page stays linkable and Back works.
      window.location.hash = '#/' + pageName
      return
    }

    this.openPage(this.state.appConf, pageName)
  }

  fail(e: any) {
    console.error(e)
    this.setState({ error: String(e) })
  }

  render(): React.ReactNode {
    if (this.state.error) {
      return <p className="notification is-danger">{this.state.error}</p>
    }

    if (!this.state.appConf) {
      return <></>
    }

    return (
      // key remounts App on navigation, which resets its component tree the
      // way a page load does on the web.
      <App key={this.state.pageName}
        ref={this.appEle}
        appConf={this.state.appConf}
        pageName={this.state.pageName}
        onNavigate={(name) => { this.jumpToPage(name) }}
        update={(event: UpdateEvent) => {
          this.backend.update(event).catch((e) => { console.error(e) })
        }}
        upload={(file, id) => this.backend.uploadFile(file, id)} />
    )
  }
}
