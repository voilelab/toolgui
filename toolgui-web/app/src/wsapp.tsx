import React, { Component } from "react"

import { App, dispatchPack, pageFromLocation } from "@toolgui-web/lib"
import { AppConf } from "@toolgui-web/lib"
import { StatefulWebSocket } from "./api/StatefulWebSocket"
import { getAppConf } from "./api/AppConfAPI"

interface WSState {
  appConf: AppConf | null
  pageName: string | null
  conn: StatefulWebSocket | null
}

export class WSApp extends Component<{}, WSState> {
  appEle: React.RefObject<App>

  constructor(props: any) {
    super(props)
    this.state = {
      appConf: null,
      pageName: null,
      conn: null,
    }
    this.appEle = React.createRef()

    this.setup()
  }

  async setup() {
    const appConf = await getAppConf()

    const pageName = pageFromLocation(window.location,
      appConf.hash_page_name_mode, appConf.page_names).name

    // The query is read on every connect, so a reconnect sees the address
    // bar as it is now.
    const conn = new StatefulWebSocket(pageName, pack => {
      dispatchPack(this.appEle.current, pack)
    }, () => pageFromLocation(window.location,
      appConf.hash_page_name_mode, appConf.page_names).query)

    conn.onConnect = () => {
      this.state.conn.send({})
    }

    conn.onStateIDChange = () => {
      this.appEle.current.clearState()
    }

    conn.init()

    // A hash change alone opens no new session, so Back and Forward between
    // pages reload. ReplaceQuery uses replaceState, which fires none.
    if (appConf.hash_page_name_mode) {
      window.addEventListener('hashchange', () => { window.location.reload() })
    }

    this.setState({ appConf, pageName, conn })
  }

  render(): React.ReactNode {
    if (!this.state.appConf) {
      return <></>
    }

    return (
      <App appConf={this.state.appConf}
        ref={this.appEle}
        update={(pack) => { this.state.conn.send(pack) }}
        upload={(f, id) => { return this.state.conn.uploadFile(f, id) }}
        download={(token) => { return this.state.conn.downloadFile(token) }} />
    )
  }
}