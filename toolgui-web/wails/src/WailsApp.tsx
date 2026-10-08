import React, { Component } from "react"

import { App, AppConf, dispatchPack, UpdateEvent } from "@toolgui-web/lib"
import { backend, downloadFile, getAppConf, onEvent, sendEvent, uploadFile } from "./api/backend"

// packEventName is the Wails event carrying every pack. One event name keeps
// create/update/delete/result in the order the page produced them.
const PACK_EVENT_NAME = "toolgui:pack"

interface WailsAppState {
  appConf: AppConf | null
  pageName: string
  query: string
}

export class WailsApp extends Component<{}, WailsAppState> {
  appEle: React.RefObject<App>

  constructor(props: {}) {
    super(props)
    this.state = {
      appConf: null,
      pageName: '',
      query: '',
    }
    this.appEle = React.createRef()

    this.setup().catch((e) => { console.error(e) })
  }

  async setup() {
    const appConf = await getAppConf()
    const pageName = appConf.page_names.length > 0 ? appConf.page_names[0] : ''

    // Listen before the first run: packs emitted with no listener are lost.
    onEvent(PACK_EVENT_NAME, (packJSON: string) => {
      const app = this.appEle.current
      if (!app) {
        return
      }

      dispatchPack(app, JSON.parse(packJSON))
    })

    this.openPage(appConf, pageName, '')
  }

  // openPage renders the page and asks the backend for a session on it.
  // Start runs after the commit, so the ref the pack listener needs is set.
  // There is no address bar, so the query lives here.
  openPage(appConf: AppConf, pageName: string, query: string) {
    this.setState({ appConf, pageName, query }, () => {
      backend().Start(pageName, query).catch((e) => { console.error(e) })
    })
  }

  jumpToPage(pageName: string, query: string) {
    // A new session starts from an empty state, the same as loading another
    // page in the browser, so drop the values the old page collected.
    this.appEle.current.clearState()
    this.openPage(this.state.appConf, pageName, query)
  }

  render(): React.ReactNode {
    if (!this.state.appConf) {
      return <></>
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
        onNavigate={(name, query) => { this.jumpToPage(name, query) }}
        update={(event: UpdateEvent) => {
          sendEvent(event).catch((e) => { console.error(e) })
        }}
        upload={(file, id) => uploadFile(file, id)}
        download={(token) => downloadFile(token)} />
    )
  }
}
