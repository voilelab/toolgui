import React, { Component } from "react"

import { App } from "./App"
import { AppConf } from "./AppConf"
import { dispatchPack } from "./dispatch"
import { DownloadFunc } from "./Download"
import { PageLocation, pageFromLocation, pageHref } from "./pageurl"
import { UpdateEvent } from "./UpdateEvent"
import { UploadFunc } from "./Upload"

// SessionBackend is a transport holding one session at a time: a desktop
// window or a browser tab. Payloads are the websocket transport's.
export interface SessionBackend {
  appConf(): Promise<AppConf>
  // start opens a session on a page with its encoded query, replacing the
  // current one, and draws it once.
  start(pageName: string, query: string): Promise<void>
  update(event: UpdateEvent): Promise<void>
  uploadFile: UploadFunc
  downloadFile: DownloadFunc
}

interface SessionAppProps {
  // connect makes the backend, handing it where packs go. Called once.
  connect: (onPack: (pack: any) => void) => SessionBackend

  // useHash keeps the page in the URL hash, `#/detail?group=a`, so a page is
  // linkable and Back works. Left out, the page lives only in memory.
  useHash?: boolean

  // embed is passed on to App.
  embed?: boolean

  // loading is shown until the app config arrives.
  loading?: React.ReactNode
}

interface SessionAppState {
  appConf: AppConf | null
  pageName: string
  query: string
  error: string | null
}

// SessionApp runs App over a SessionBackend: navigation opens a new session,
// the way a page load does on the web.
export class SessionApp extends Component<SessionAppProps, SessionAppState> {
  private appEle = React.createRef<App>()
  private backend: SessionBackend
  private onHashChange: (() => void) | null = null

  constructor(props: SessionAppProps) {
    super(props)
    this.state = { appConf: null, pageName: '', query: '', error: null }

    // Before the first run: packs delivered with no listener are lost.
    this.backend = props.connect((pack) => {
      const app = this.appEle.current
      if (app) {
        dispatchPack(app, pack)
      }
    })

    this.setup().catch((e) => { this.fail(e) })
  }

  private async setup() {
    const appConf = await this.backend.appConf()

    if (this.props.useHash) {
      const fromHash = () => pageFromLocation(window.location, true, appConf.page_names)
      this.onHashChange = () => { this.openPage(appConf, fromHash()) }
      window.addEventListener('hashchange', this.onHashChange)
      this.openPage(appConf, fromHash())
      return
    }

    this.openPage(appConf, { name: appConf.page_names[0] ?? '', query: '' })
  }

  componentWillUnmount() {
    if (this.onHashChange) {
      window.removeEventListener('hashchange', this.onHashChange)
    }
  }

  // openPage renders the page, then starts its session once the ref the pack
  // listener needs is set.
  private openPage(appConf: AppConf, { name, query }: PageLocation) {
    // A new session starts from an empty state.
    this.appEle.current?.clearState()

    this.setState({ appConf, pageName: name, query }, () => {
      this.backend.start(name, query).catch((e) => { this.fail(e) })
    })
  }

  private hashMode(): boolean {
    return !!this.props.useHash && !!this.state.appConf?.hash_page_name_mode
  }

  private navigate(name: string, query: string) {
    if (this.hashMode()) {
      // Let the URL drive, so Back works.
      window.location.hash = pageHref(name, query, true)
      return
    }

    this.openPage(this.state.appConf, { name, query })
  }

  // replaceQuery keeps the address bar in step with Params.ReplaceQuery.
  // replaceState fires no hashchange, so the session stays.
  private replaceQuery(query: string) {
    if (this.hashMode()) {
      window.history.replaceState(window.history.state, '',
        pageHref(this.state.pageName, query, true))
    }
  }

  private fail(e: any) {
    console.error(e)
    this.setState({ error: String(e) })
  }

  render(): React.ReactNode {
    if (this.state.error) {
      return <p className="notification is-danger">{this.state.error}</p>
    }

    if (!this.state.appConf) {
      return this.props.loading ?? <></>
    }

    return (
      // key remounts App on navigation, resetting its tree like a page load.
      // The query is in it too: other parameters make another page load.
      <App key={`${this.state.pageName}?${this.state.query}`}
        ref={this.appEle}
        appConf={this.state.appConf}
        pageName={this.state.pageName}
        query={this.state.query}
        embed={this.props.embed}
        onNavigate={(name, query) => { this.navigate(name, query) }}
        onReplaceQuery={(query) => { this.replaceQuery(query) }}
        update={(event: UpdateEvent) => {
          this.backend.update(event).catch((e) => { console.error(e) })
        }}
        upload={(file, id) => this.backend.uploadFile(file, id)}
        download={(token) => this.backend.downloadFile(token)} />
    )
  }
}
