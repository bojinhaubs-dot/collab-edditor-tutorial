import { useEffect, useState } from 'react'
import * as Y from 'yjs'
import { WebsocketProvider } from 'y-websocket'

const websocketURL = import.meta.env.VITE_COLLAB_WS_URL || 'ws://127.0.0.1:1234'

const labels = {
  connecting: '连接中',
  connected: '已连接协同服务',
  disconnected: '已断开（网络离线）',
  reconnecting: '已断开，正在重连',
}

export default function CollaborationStatus({ documentID }) {
  const [status, setStatus] = useState('connecting')

  useEffect(() => {
    const ydoc = new Y.Doc()
    const provider = new WebsocketProvider(websocketURL, `document-${documentID}`, ydoc)

    const onStatus = ({ status: connection }) => {
      setStatus(connection === 'connected'
        ? 'connected'
        : navigator.onLine ? 'reconnecting' : 'disconnected')
    }
    const onOffline = () => setStatus('disconnected')
    const onOnline = () => setStatus(provider.wsconnected ? 'connected' : 'reconnecting')

    provider.on('status', onStatus)
    window.addEventListener('offline', onOffline)
    window.addEventListener('online', onOnline)

    return () => {
      window.removeEventListener('offline', onOffline)
      window.removeEventListener('online', onOnline)
      provider.off('status', onStatus)
      provider.destroy()
      ydoc.destroy()
    }
  }, [documentID])

  return <span className={`connection-status ${status}`} role="status" aria-live="polite">{labels[status]}</span>
}
