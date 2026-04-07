import { useEffect, useRef } from 'react'
import { EditorView, basicSetup } from 'codemirror'
import { EditorState } from '@codemirror/state'
import { yaml } from '@codemirror/lang-yaml'
import { oneDark } from '@codemirror/theme-one-dark'
import { indentUnit, indentOnInput } from '@codemirror/language'
import { indentWithTab } from '@codemirror/commands'
import { keymap } from '@codemirror/view'

interface CodeEditorProps {
  value: string
  onChange?: (value: string) => void
  height?: string
  readOnly?: boolean
}

const FONT = 'SF Mono, Fira Code, JetBrains Mono, Menlo, Consolas, monospace'

export function CodeEditor({ value, onChange, height = '300px', readOnly = false }: CodeEditorProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const viewRef = useRef<EditorView | null>(null)
  const onChangeRef = useRef(onChange)
  onChangeRef.current = onChange

  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    const state = EditorState.create({
      doc: value,
      extensions: [
        basicSetup,
        yaml(),
        oneDark,
        indentUnit.of('  '),
        indentOnInput(),
        keymap.of([indentWithTab]),
        EditorState.tabSize.of(2),
        EditorView.theme({
          '&': { height, fontSize: '12px' },
          '.cm-scroller': { fontFamily: FONT, overflow: 'auto' },
          '.cm-gutters': { borderRight: 'none' },
          '.cm-content': { padding: '8px 0' },
        }),
        ...(readOnly ? [EditorState.readOnly.of(true)] : []),
        EditorView.updateListener.of((update) => {
          if (update.docChanged && onChangeRef.current) {
            onChangeRef.current(update.state.doc.toString())
          }
        }),
      ],
    })

    const view = new EditorView({ state, parent: container })
    viewRef.current = view

    return () => {
      view.destroy()
      viewRef.current = null
    }
  }, [])

  return <div ref={containerRef} className="overflow-hidden rounded-b-xl" />
}
