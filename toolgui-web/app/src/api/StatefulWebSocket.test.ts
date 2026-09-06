import { afterEach, beforeEach, expect, test, vi } from 'vitest'

import { StatefulWebSocket } from './StatefulWebSocket'

// FakeWebSocket stands in for the browser's, so a test can decide when a
// socket opens, what it delivers and when it closes.
class FakeWebSocket {
  static opened: FakeWebSocket[] = []

  url: string
  sent: string[] = []
  closed = false

  onopen: () => void = () => { }
  onmessage: (e: { data: string }) => void = () => { }
  onclose: () => void = () => { }

  constructor(url: string) {
    this.url = url
    FakeWebSocket.opened.push(this)
  }

  send(data: string) {
    this.sent.push(data)
  }

  close() {
    if (this.closed) {
      return
    }

    this.closed = true
    this.onclose()
  }
}

// settle lets the pending health check resolve without moving the clock, so a
// socket that is only waiting on a fetch gets to open.
function settle() {
  return vi.advanceTimersByTimeAsync(0)
}

beforeEach(() => {
  vi.useFakeTimers()
  FakeWebSocket.opened = []
  globalThis.WebSocket = FakeWebSocket as any
  globalThis.fetch = vi.fn(async () => ({ status: 200 })) as any
  vi.spyOn(console, 'log').mockImplementation(() => { })
  vi.spyOn(console, 'error').mockImplementation(() => { })
})

afterEach(() => {
  vi.useRealTimers()
  vi.restoreAllMocks()
})

// connect runs the socket up to its first open connection.
async function connect() {
  const recv = vi.fn()
  const conn = new StatefulWebSocket('main', recv)
  conn.init()
  await settle()

  expect(FakeWebSocket.opened).toHaveLength(1)
  FakeWebSocket.opened[0].onopen()

  return { conn, recv }
}

test('connects without waiting first', async () => {
  const { conn } = await connect()

  expect(FakeWebSocket.opened[0].url).toMatch(/\/api\/update\/main$/)
  expect(conn.conn).not.toBeNull()
})

test('waits before reconnecting after a socket drops', async () => {
  await connect()

  FakeWebSocket.opened[0].close()
  await settle()

  // Without the backoff the health check answers straight away and the
  // reconnect happens in the same tick, spinning the server.
  expect(FakeWebSocket.opened).toHaveLength(1)

  await vi.advanceTimersByTimeAsync(200)
  expect(FakeWebSocket.opened).toHaveLength(2)
})

test('backs off further while the server keeps dropping the socket', async () => {
  await connect()

  FakeWebSocket.opened[0].close()
  await vi.advanceTimersByTimeAsync(200)
  expect(FakeWebSocket.opened).toHaveLength(2)

  FakeWebSocket.opened[1].onopen()
  FakeWebSocket.opened[1].close()
  await vi.advanceTimersByTimeAsync(200)
  expect(FakeWebSocket.opened).toHaveLength(2)

  await vi.advanceTimersByTimeAsync(100)
  expect(FakeWebSocket.opened).toHaveLength(3)
})

test('starts the backoff over after a connection that lasted', async () => {
  await connect()

  await vi.advanceTimersByTimeAsync(5000)
  FakeWebSocket.opened[0].close()
  await settle()

  expect(FakeWebSocket.opened).toHaveLength(2)
})

test('gives up on a fatal error instead of reconnecting', async () => {
  const { recv } = await connect()

  const pack = { success: false, error: 'page not found', fatal: true }
  FakeWebSocket.opened[0].onmessage({ data: JSON.stringify(pack) })

  // The error still has to reach the app, or the page would go quiet with
  // nothing on screen to say why.
  expect(recv).toHaveBeenCalledWith(pack)

  expect(FakeWebSocket.opened[0].closed).toBe(true)

  await vi.advanceTimersByTimeAsync(60000)
  expect(FakeWebSocket.opened).toHaveLength(1)
})

test('keeps reconnecting after an error the server may recover from', async () => {
  const { recv } = await connect()

  const pack = { success: false, error: 'state id already alive' }
  FakeWebSocket.opened[0].onmessage({ data: JSON.stringify(pack) })
  FakeWebSocket.opened[0].close()

  expect(recv).toHaveBeenCalledWith(pack)

  await vi.advanceTimersByTimeAsync(200)
  expect(FakeWebSocket.opened).toHaveLength(2)
})
