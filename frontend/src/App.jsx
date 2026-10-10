import { useCallback, useEffect, useState } from 'react'
import CodeEditor from './CodeEditor.jsx'

const documentID = new URLSearchParams(window.location.search).get('id')

export default function App() {
  const [document, setDocument] = useState(null)
  const [content, setContent] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!/^[1-9]\d*$/.test(documentID ?? '')) {
      setError('请在地址中提供有效的文档 ID，例如 ?id=1。')
      setLoading(false)
      return
    }

    const controller = new AbortController()
    async function loadDocument() {
      try {
        const response = await fetch(`/documents/${documentID}`, {
          signal: controller.signal,
        })
        if (!response.ok) {
          throw new Error(response.status === 404
            ? `文档 ${documentID} 不存在。`
            : `加载文档失败（HTTP ${response.status}）。`)
        }
        const loaded = await response.json()
        setDocument(loaded)
        setContent(loaded.content)
      } catch (cause) {
        if (!controller.signal.aborted) {
          setError(cause instanceof TypeError
            ? '无法连接文档 API，请确认 Go 服务正在运行。'
            : cause.message)
        }
      } finally {
        if (!controller.signal.aborted) setLoading(false)
      }
    }
    loadDocument()
    return () => controller.abort()
  }, [])

  const handleChange = useCallback((nextContent) => setContent(nextContent), [])

  return (
    <main className="page">
      <header className="page-header">
        <p className="eyebrow">B1 · 单用户编辑</p>
        <h1>代码文档编辑器</h1>
        <p className="description">从 Go API 加载文档，在当前页面中编辑代码。</p>
      </header>

      {loading && <p role="status">正在加载文档…</p>}
      {error && <p className="error" role="alert">{error}</p>}

      {document && (
        <section className="document" aria-label="文档内容">
          <div className="document-bar">
            <div>
              <h2>{document.title}</h2>
              <span className="document-id">文档 ID：{document.id}</span>
            </div>
            <span className="mode">Go 语法高亮</span>
          </div>
          <CodeEditor initialContent={document.content} onChange={handleChange} />
          <div className="document-footer" aria-live="polite">
            <span>{content.length} 个字符</span>
            <span>{content === document.content ? '尚无本地修改' : '本地修改未保存'}</span>
          </div>
          <p className="note">当前阶段只支持本地编辑；刷新页面会重新加载 API 中的初始内容。</p>
        </section>
      )}
    </main>
  )
}
