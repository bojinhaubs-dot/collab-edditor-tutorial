import { useEffect, useRef } from 'react'
import { go } from '@codemirror/lang-go'
import { basicSetup, EditorView } from 'codemirror'

export default function CodeEditor({ initialContent, onChange }) {
  const containerRef = useRef(null)

  useEffect(() => {
    const view = new EditorView({
      doc: initialContent,
      extensions: [
        basicSetup,
        go(),
        EditorView.updateListener.of((update) => {
          if (update.docChanged) {
            onChange(update.state.doc.toString())
          }
        }),
      ],
      parent: containerRef.current,
    })

    // CodeMirror owns its DOM and document while mounted; React only observes edits.
    return () => view.destroy()
  }, [initialContent, onChange])

  return <div className="editor" ref={containerRef} aria-label="代码编辑器" />
}
