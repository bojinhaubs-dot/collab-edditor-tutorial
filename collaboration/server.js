import { createServer } from 'node:http'
import { WebSocketServer } from 'ws'
import { setupWSConnection } from '@y/websocket-server/utils'

export function createCollaborationServer() {
  const server = createServer((_request, response) => {
    response.writeHead(200, { 'Content-Type': 'text/plain; charset=utf-8' })
    response.end('ok')
  })
  const sockets = new WebSocketServer({ noServer: true })

  // setupWSConnection handles the standard Yjs sync and awareness messages.
  sockets.on('connection', setupWSConnection)
  server.on('upgrade', (request, socket, head) => {
    sockets.handleUpgrade(request, socket, head, (client) => {
      sockets.emit('connection', client, request)
    })
  })

  return server
}
