import { createCollaborationServer } from './server.js'

const host = process.env.COLLAB_HOST || '127.0.0.1'
const port = Number(process.env.COLLAB_PORT || '1234')

if (!Number.isInteger(port) || port < 1 || port > 65535) {
  throw new Error('COLLAB_PORT must be an integer from 1 to 65535')
}

createCollaborationServer().listen(port, host, () => {
  console.log(`Yjs collaboration server listening at ws://${host}:${port}`)
})
