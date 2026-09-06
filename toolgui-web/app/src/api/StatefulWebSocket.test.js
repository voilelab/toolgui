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

// Drop the connection the server just accepted, as it does for a page it
// cannot serve.
async function closeLast() {
  const conn = sockets[sockets.length - 1]
  conn.onopen()
  conn.onclose()
  await vi.advanceTimersByTimeAsync(0)
}

test('a connection the server drops at once retries with a growing delay', async () => {
  const ws = new StatefulWebSocket('main', () => { })
  ws.init()
  await vi.advanceTimersByTimeAsync(0)
  expect(sockets).toHaveLength(1)

  await closeLast()
  await vi.advanceTimersByTimeAsync(100)
  expect(sockets).toHaveLength(1)
  await vi.advanceTimersByTimeAsync(500)
  expect(sockets).toHaveLength(2)

  // The second drop waits twice as long.
  await closeLast()
  await vi.advanceTimersByTimeAsync(600)
  expect(sockets).toHaveLength(2)
  await vi.advanceTimersByTimeAsync(500)
  expect(sockets).toHaveLength(3)
})

test('a connection that lived a while reconnects from the base delay', async () => {
  const ws = new StatefulWebSocket('index', () => { })
  ws.init()
  await vi.advanceTimersByTimeAsync(0)

  await closeLast()
  await vi.advanceTimersByTimeAsync(600)
  expect(sockets).toHaveLength(2)

  await vi.advanceTimersByTimeAsync(10000)
  await closeLast()
  await vi.advanceTimersByTimeAsync(600)
  expect(sockets).toHaveLength(3)
})
