import { UploadResult } from "@toolgui-web/lib"

const healthCheckURL = "/api/health"

const fileUploadURL = "/api/files"

// Reconnect backoff. Without it a server that hands out a socket and drops it
// right away spins us in a tight loop: onclose walks straight back to Ping,
// and a server that is up answers the health check on the first try.
const MIN_RETRY_WAIT_MS = 200
const MAX_RETRY_WAIT_MS = 60000

// A connection that stayed open this long counts as healthy, so the drop that
// ends it starts its retries from scratch instead of inheriting the backoff.
const RETRY_RESET_MS = 5000

enum WebSocketState {
  Initial,
  Ping,
  TryConnect,
  Connected,

  // Dead is the end of the line: the server reported an error a reconnect
  // would only hit again.
  Dead,
}

enum WebSocketAction {
  OK,
  Error,
  Closed,
  Fatal,

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

  /** How long to wait before the next connection attempt. */
  retryWaitMS: number

  /** When the current connection opened, 0 while there is none. */
  connectedAt: number

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
    this.retryWaitMS = 0
    this.connectedAt = 0
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
      case WebSocketState.Dead:
        this.state = state
        break
      default:
        console.error('undefined state', state)
        throw new Error('undefined state')
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
          case WebSocketAction.Fatal:
            this.walkTo(WebSocketState.Dead)
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
          case WebSocketAction.Fatal:
            this.walkTo(WebSocketState.Dead)
            return
          default:
            break
        }
        break
      case WebSocketState.Dead:
        // A dead socket stays dead until the page is reloaded.
        return
    }

    console.error('undefine action on state', this.state, action)
    throw new Error('undefine action on state')
  }

  async ping() {
    while (1) {
      if (this.retryWaitMS > 0) {
        await sleep(this.retryWaitMS)
      }

      this.retryWaitMS = Math.min(
        Math.max(this.retryWaitMS * 1.5, MIN_RETRY_WAIT_MS),
        MAX_RETRY_WAIT_MS)

      try {
        const resp = await fetch(healthCheckURL)
        if (resp.status === 200) {
          break
        }

        console.error('health check', resp)
      } catch (e) {
        console.error(e)
      }
    }

    console.log('ping ok')
    this.walk(WebSocketAction.OK)
  }

  tryConnect() {
    this.conn = new WebSocket(getUpdateURI(this.pageName))
    var that = this

    this.conn.onopen = function () {
      that.connectedAt = Date.now()
      that.conn.send(JSON.stringify({ state_id: that.stateID }))
      console.log('socket open ok')
      that.walk(WebSocketAction.OK)
    }

    this.conn.onmessage = function (e) {
      const data = JSON.parse(e.data)

      // A fatal error is a property of the request, not of the connection:
      // an unknown page name stays unknown. Remember it so the close it comes
      // with ends the socket for good, and hand it on so the error still
      // reaches the screen.
      if (data.fatal) {
        that.fatal()
      }

      if (data.state_id) {
        that.stateID = data.state_id
        that.onStateIDChange()
        return
      }

      that.recv(data)
    }

    this.conn.onclose = function () {
      that.conn = null

      if (that.state === WebSocketState.Dead) {
        return
      }

      const lived = that.connectedAt === 0 ? 0 : Date.now() - that.connectedAt
      if (lived >= RETRY_RESET_MS) {
        that.retryWaitMS = 0
      }

      that.connectedAt = 0
      that.walk(WebSocketAction.Closed)
    }
  }

  /** Give up on the connection: stop reconnecting and close what is open. */
  fatal() {
    this.walk(WebSocketAction.Fatal)

    if (this.conn !== null) {
      this.conn.close()
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
