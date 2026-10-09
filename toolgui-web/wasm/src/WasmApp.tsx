import React, { Component } from "react"

import {
  App, AppConf, dispatchPack, PageLocation, pageFromLocation, pageHref,
  UpdateEvent,
} from "@toolgui-web/lib"
import { Backend, LoadProgress } from "./api/backend"

interface WasmAppState {
  appConf: AppConf | null
  pageName: string
  query: string
  error: string | null
  progress: LoadProgress | null
}

// pageFromHash reads the page off the URL. A static host cannot route paths,
// so the hash is the only place a page name can live: `#/detail?group=a`.
function pageFromHash(appConf: AppConf): PageLocation {
  return pageFromLocation(window.location, true, appConf.page_names)
}

// embedFromSearch reads the embed flag off the URL. embed is a display mode of
// the front end rather than something the app declares, so it lives in the
// address: the same binary serves a site and an iframe.
function embedFromSearch(search: string): boolean {
  const value = new URLSearchParams(search).get('embed')

  // Bare `?embed` counts. The two spellings that do not are the ones a caller
  // building the URL from a boolean would produce for "no".
  return value !== null && value !== '0' && value !== 'false'
}

export class WasmApp extends Component<{}, WasmAppState> {
  appEle: React.RefObject<App>
  backend: Backend

  // Read once: the query string cannot change without a page load, while the
  // hash changes on every page.
  embed: boolean = embedFromSearch(window.location.search)

  constructor(props: {}) {
    super(props)
    this.state = {
      appConf: null,
      pageName: '',
      query: '',
      error: null,
      progress: null,
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
    }, this.embed, (progress) => { this.setState({ progress }) },
      window.location.search)

    this.setup().catch((e) => { this.fail(e) })
  }

  async setup() {
    const appConf = await this.backend.appConf()

    // The wasm program stays loaded across pages, so navigation is a new
    // session rather than a page load.
    window.addEventListener('hashchange', () => {
      this.openPage(appConf, pageFromHash(appConf))
    })

    this.openPage(appConf, pageFromHash(appConf))
  }

  // openPage renders the page and asks the backend for a session on it. start
  // runs after the commit, so the ref the pack listener needs is set.
  openPage(appConf: AppConf, { name, query }: PageLocation) {
    this.appEle.current?.clearState()

    this.setState({ appConf, pageName: name, query }, () => {
      this.backend.start(name, query).catch((e) => { this.fail(e) })
    })
  }

  jumpToPage(name: string, query: string) {
    if (this.state.appConf.hash_page_name_mode) {
      // Let the URL drive, so a page stays linkable and Back works.
      window.location.hash = pageHref(name, query, true)
      return
    }

    this.openPage(this.state.appConf, { name, query })
  }

  // replaceQuery keeps the address bar in step with Params.ReplaceQuery. Only
  // hash mode has the page in the URL. replaceState fires no hashchange, so
  // the session stays.
  replaceQuery(query: string) {
    if (this.state.appConf.hash_page_name_mode) {
      window.history.replaceState(window.history.state, '',
        pageHref(this.state.pageName, query, true))
    }
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
      return <Loading progress={this.state.progress} />
    }

    return (
      // key remounts App on navigation, which resets its component tree the
      // way a page load does on the web.
      // The query is in the key too: a link to the same page with other
      // parameters is a new page load.
      <App key={`${this.state.pageName}?${this.state.query}`}
        ref={this.appEle}
        appConf={this.state.appConf}
        pageName={this.state.pageName}
        query={this.state.query}
        embed={this.embed}
        onNavigate={(name, query) => { this.jumpToPage(name, query) }}
        onReplaceQuery={(query) => { this.replaceQuery(query) }}
        update={(event: UpdateEvent) => {
          this.backend.update(event).catch((e) => { console.error(e) })
        }}
        upload={(file, id) => this.backend.uploadFile(file, id)}
        download={(token) => this.backend.downloadFile(token)} />
    )
  }
}

// Loading shows the download of app.wasm, which is most of a first visit.
function Loading({ progress }: { progress: LoadProgress | null }) {
  if (!progress) {
    return <div className="tg-loading"><progress /></div>
  }

  if (progress.done) {
    return <div className="tg-loading"><progress /><p>Starting…</p></div>
  }

  const { loaded, total } = progress
  return (
    <div className="tg-loading">
      {total > 0 ? <progress value={loaded} max={total} /> : <progress />}
      <p>Loading {formatMB(loaded)}{total > 0 ? ` / ${formatMB(total)}` : ''}</p>
    </div>
  )
}

function formatMB(n: number): string {
  return `${(n / (1024 * 1024)).toFixed(1)} MB`
}
