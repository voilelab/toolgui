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

// Reconnect backoff. Without it a connection the server closes right away --
// an unknown page name, say -- reconnects in a tight loop.
const RECONNECT_BASE_MS = 500
const RECONNECT_MAX_MS = 30000

// A connection that stayed open this long counts as healthy, so the next
// drop retries from the base delay.
const HEALTHY_CONN_MS = 5000

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

  /** Delay before the next reconnect, grown on each short-lived connection. */
  reconnectMS: number

  /** Timestamp the current connection opened at, 0 until it does. */
  openAt: number

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
    this.reconnectMS = 0
    this.openAt = 0
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
    if (this.reconnectMS > 0) {
      await sleep(this.reconnectMS)
    }

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
    // The handshake is not uptime: a socket that never opens counts as
    // short-lived.
    this.openAt = 0
    this.conn = new WebSocket(getUpdateURI(this.pageName))
    var that = this

    this.conn.onopen = function () {
      that.openAt = Date.now()
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

      if (that.openAt !== 0 && Date.now() - that.openAt >= HEALTHY_CONN_MS) {
        that.reconnectMS = 0
      } else {
        that.reconnectMS = Math.min(
          that.reconnectMS === 0 ? RECONNECT_BASE_MS : that.reconnectMS * 2,
          RECONNECT_MAX_MS)
      }

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