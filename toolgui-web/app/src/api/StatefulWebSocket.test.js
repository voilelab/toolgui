import { afterEach, beforeEach, expect, test, vi } from 'vitest'

import { StatefulWebSocket } from './StatefulWebSocket'

let sockets

beforeEach(() => {
  vi.useFakeTimers()
  sockets = []

  vi.stubGlobal('WebSocket', class {
    constructor(url) {
      this.url = url
      sockets.push(this)
    }
    send() { }
    close() { }
  })

  // The health check the state machine pings before every connect.
  vi.stubGlobal('fetch', vi.fn(async () => ({ status: 200 })))
})

afterEach(() => {
  vi.unstubAllGlobals()
  vi.useRealTimers()
})

// The socket the state machine is on right now.
function last() {
  return sockets[sockets.length - 1]
}

// Let the awaits inside the state machine settle.
function flush() {
  return vi.advanceTimersByTimeAsync(0)
}

async function connect(ws) {
  ws.init()
  await flush()
  expect(sockets).toHaveLength(1)
}

test('a connection the server drops at once retries with a growing delay', async () => {
  const ws = new StatefulWebSocket('main', () => { })
  await connect(ws)

  last().onopen()
  last().onclose()
  await vi.advanceTimersByTimeAsync(100)
  expect(sockets).toHaveLength(1)
  await vi.advanceTimersByTimeAsync(500)
  expect(sockets).toHaveLength(2)

  // The second drop waits twice as long.
  last().onopen()
  last().onclose()
  await vi.advanceTimersByTimeAsync(600)
  expect(sockets).toHaveLength(2)
  await vi.advanceTimersByTimeAsync(500)
  expect(sockets).toHaveLength(3)
})

test('a page the server rejects for good is not asked for again', async () => {
  const ws = new StatefulWebSocket('main', () => { })
  await connect(ws)

  last().onopen()
  last().onmessage({ data: JSON.stringify({ success: false, fatal: true, error: 'page not found' }) })
  last().onclose()
  await vi.advanceTimersByTimeAsync(60000)
  expect(sockets).toHaveLength(1)
})

test('a connection that stayed open a while reconnects at once', async () => {
  const ws = new StatefulWebSocket('index', () => { })
  await connect(ws)

  last().onopen()
  await vi.advanceTimersByTimeAsync(10000)
  last().onclose()
  await flush()
  expect(sockets).toHaveLength(2)
})

test('a socket that never opens counts as short-lived', async () => {
  const ws = new StatefulWebSocket('index', () => { })
  await connect(ws)

  last().onopen()
  await vi.advanceTimersByTimeAsync(10000)
  last().onclose()
  await flush()

  // The reconnect hangs in its handshake and dies without ever opening. The
  // wait is the base delay, not the reset the previous connection earned.
  await vi.advanceTimersByTimeAsync(10000)
  last().onclose()
  await vi.advanceTimersByTimeAsync(100)
  expect(sockets).toHaveLength(2)
  await vi.advanceTimersByTimeAsync(500)
  expect(sockets).toHaveLength(3)
})
