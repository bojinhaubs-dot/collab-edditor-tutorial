import assert from 'node:assert/strict'
import { randomUUID } from 'node:crypto'
import { once } from 'node:events'
import { test } from 'node:test'
import WebSocket from 'ws'
import * as Y from 'yjs'
import { WebsocketProvider } from 'y-websocket'
import { docs } from '@y/websocket-server/utils'
import { createCollaborationServer } from './server.js'

function waitFor(provider, event, matches = () => true) {
  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => {
      provider.off(event, onEvent)
      reject(new Error(`Timed out waiting for ${event}`))
    }, 8000)
    function onEvent(value) {
      if (!matches(value)) return
      clearTimeout(timeout)
      provider.off(event, onEvent)
      resolve(value)
    }
    provider.on(event, onEvent)
  })
}

test('Yjs provider receives sync messages and reconnects', async () => {
  const server = createCollaborationServer()
  server.listen(0, '127.0.0.1')
  await once(server, 'listening')

  const ydoc = new Y.Doc()
  const room = `b2-test-${randomUUID()}`
  const provider = new WebsocketProvider(
    `ws://127.0.0.1:${server.address().port}`,
    room,
    ydoc,
    { WebSocketPolyfill: WebSocket },
  )

  try {
    const upgraded = once(provider.ws, 'upgrade')
    const firstMessage = once(provider.ws, 'message')
    const connected = waitFor(provider, 'status', ({ status }) => status === 'connected')
    const synced = waitFor(provider, 'sync', (value) => value === true)
    const [response] = await upgraded
    const [message, isBinary] = await firstMessage
    await connected
    await synced
    assert.equal(response.statusCode, 101)
    assert.equal(isBinary, true)
    assert.ok((message.byteLength ?? message.length) > 0)
    assert.equal(provider.wsconnected, true)
    assert.equal(provider.synced, true)

    const disconnected = waitFor(provider, 'status', ({ status }) => status === 'disconnected')
    const reconnected = waitFor(provider, 'status', ({ status }) => status === 'connected')
    provider.ws.close()
    await disconnected
    await reconnected
    assert.equal(provider.wsconnected, true)
  } finally {
    provider.destroy()
    ydoc.destroy()
    await new Promise((resolve, reject) => server.close((error) => error ? reject(error) : resolve()))
    docs.get(room)?.destroy()
    docs.delete(room)
  }
})
