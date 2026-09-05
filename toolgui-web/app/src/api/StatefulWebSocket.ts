import { UploadResult } from "@toolgui-web/lib"

const healthCheckURL = "/api/health"

const fileUploadURL = "/api/files"

enum WebSocketState {
  Initial,
  Ping,
  TryConnect,
  Connected,

  // TODO: Dead
}

enum WebSocketAction {
  OK,
  Error,
  Closed,

  // TODO: support timeout check
}

function sleep(ms: number) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

function getSocketURI() {
  var scheme = 'ws'
  if (window.location.origin.startsWith('https')) {
    scheme = 'wss'
  }

  return `${scheme}://${window.location.host}`
}

function getUpdateURI(pageName: string) {
  return `${getSocketURI()}/api/update/${pageName}`
}

export class StatefulWebSocket {
  conn: WebSocket | null
  state: WebSocketState
  pageName: string
  stateID: string

  /** Receive pack from connected websocket. */
  recv: (pack: any) => void

  /** Call when stateID is assigned a new value from server. */
  onStateIDChange: () => void = () => { }

  /** Call when state change from TryConnect to Connected. */
  onConnect: () => void = () => { }

  constructor(pageName: string, recv: (pack: any) => void) {
    this.state = WebSocketState.Initial
    this.pageName = pageName
    this.conn = null
    this.stateID = ''
    this.recv = recv
  }

  init() {
    if (this.state !== WebSocketState.Initial) {
      throw new Error('init should call on state Initial')
    }

    this.walk(WebSocketAction.OK)
  }

  walkTo(state: WebSocketState) {
    switch (state) {
      case WebSocketState.Ping:
        this.state = state
        this.ping()
        break
      case WebSocketState.TryConnect:
        this.state = state
        this.tryConnect()
        break
      case WebSocketState.Connected:
        this.state = state
        this.onConnect()
        break
      default:
        console.error('undefined state', state)
        throw new Error('undefine state')
    }
  }

  walk(action: WebSocketAction) {
    switch (this.state) {
      case WebSocketState.Initial:
        if (action === WebSocketAction.OK) {
          this.walkTo(WebSocketState.Ping)
          return
        }
        break
      case WebSocketState.Ping:
        if (action === WebSocketAction.OK) {
          this.walkTo(WebSocketState.TryConnect)
          return
        }
        break
      case WebSocketState.TryConnect:
        switch (action) {
          case WebSocketAction.OK:
            this.walkTo(WebSocketState.Connected)
            return
          case WebSocketAction.Error:
          case WebSocketAction.Closed:
            this.walkTo(WebSocketState.Ping)
            return
          default:
            break
        }
        break
      case WebSocketState.Connected:
        switch (action) {
          case WebSocketAction.Error:
          case WebSocketAction.Closed:
            this.walkTo(WebSocketState.Ping)
            return
          default:
            break
        }
        break
    }

    console.error('undefine action on state', this.state, action)
    throw new Error('undefine action on state')
  }

  async ping() {
    var waitMS = 200

    while (1) {
      var resp: Response
      try {
        resp = await fetch(healthCheckURL)
        if (resp.status === 200) {
          break
        }

        console.error('health check', resp)
      } catch (e) {
        console.error(e)
      }

      await sleep(waitMS)

      waitMS *= 1.5
      waitMS = Math.min(waitMS, 60000)
    }

    console.log('ping ok')
    this.walk(WebSocketAction.OK)
  }

  tryConnect() {
    this.conn = new WebSocket(getUpdateURI(this.pageName))
    var that = this

    this.conn.onopen = function () {
      that.conn.send(JSON.stringify({ state_id: that.stateID }))
      console.log('socket open ok')
      that.walk(WebSocketAction.OK)
    }

    this.conn.onmessage = function (e) {
      const data = JSON.parse(e.data)
      if (data.state_id) {
        that.stateID = data.state_id
        that.onStateIDChange()
        return
      }

      that.recv(data)
    }

    this.conn.onclose = function () {
      that.conn = null
      that.walk(WebSocketAction.Closed)
    }
  }

  send(pack: any) {
    if (this.state !== WebSocketState.Connected || this.conn === null) {
      throw new Error('websocket is not prepared')
    }

    this.conn.send(JSON.stringify(pack))
  }

  async uploadFile(file: File): Promise<UploadResult> {
    if (this.stateID === '') {
      return { ok: false, error: 'state id is not prepared' }
    }

    const formData = new FormData()
    formData.append('file', file, file.name)

    const resp = await fetch(fileUploadURL, {
      method: 'POST',
      body: formData,
      headers: { STATE_ID: this.stateID },
    })

    if (!resp.ok) {
      return { ok: false, error: `upload failed with status ${resp.status}` }
    }

    return { ok: true }
  }
}