import React, { Component } from "react"

import { App, dispatchPack } from "@toolgui-web/lib"
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

    var pageName = ''
    if (appConf.hash_page_name_mode) {
      if (window.location.hash) {
        // should be #/{name}
        pageName = window.location.hash.substring(2)
      } else if (appConf.page_names.length > 0) {
        pageName = appConf.page_names[0]
      }
    } else {
      pageName = window.location.pathname.substring(1)
    }

    // A page the app does not have has nothing to connect to: the server
    // closes the socket at once, and reconnecting only spams its log. App
    // renders the not-found page from appConf on its own.
    if (!appConf.page_confs[pageName]) {
      this.setState({ appConf, pageName, conn: null })
      return
    }

    const conn = new StatefulWebSocket(pageName, pack => {
      dispatchPack(this.appEle.current, pack)
    })

    conn.onConnect = () => {
      this.state.conn.send({})
    }

    conn.onStateIDChange = () => {
      this.appEle.current.clearState()
    }

    conn.init()

    this.setState({ appConf, pageName, conn })
  }

  render(): React.ReactNode {
    if (!this.state.appConf) {
      return <></>
    }

    return (
      <App appConf={this.state.appConf}
        ref={this.appEle}
        update={(pack) => { this.state.conn?.send(pack) }}
        upload={(f) => {
          if (!this.state.conn) {
            return Promise.resolve({ ok: false, error: 'not connected' })
          }
          return this.state.conn.uploadFile(f)
        }} />
    )
  }
}