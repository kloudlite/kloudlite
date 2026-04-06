import { useRef, useEffect } from 'react'

interface CodeEditorProps {
  value: string
  onChange?: (value: string) => void
  height?: string
  readOnly?: boolean
}

export function CodeEditor({ value, onChange, height = '300px', readOnly = false }: CodeEditorProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const lineCountRef = useRef<HTMLDivElement>(null)
  const lines = value.split('\n')

  function handleScroll() {
    if (textareaRef.current && lineCountRef.current) {
      lineCountRef.current.scrollTop = textareaRef.current.scrollTop
    }
  }

  return (
    <div className="flex overflow-hidden" style={{ height, fontFamily: 'SF Mono, Fira Code, JetBrains Mono, Menlo, Consolas, monospace' }}>
      {/* Line numbers */}
      <div
        ref={lineCountRef}
        className="shrink-0 select-none overflow-hidden bg-[#282c34] py-3 pr-3 text-right"
        style={{ width: '48px' }}
      >
        {lines.map((_, i) => (
          <div key={i} className="text-[12px] leading-[20px] text-[#636d83]">
            {i + 1}
          </div>
        ))}
      </div>

      {/* Editor */}
      <textarea
        ref={textareaRef}
        className="w-full resize-none bg-[#282c34] py-3 pl-3 pr-4 text-[12px] leading-[20px] text-[#abb2bf] outline-none"
        value={value}
        onChange={(e) => onChange?.(e.target.value)}
        onScroll={handleScroll}
        readOnly={readOnly}
        spellCheck={false}
        style={{ tabSize: 2 }}
      />
    </div>
  )
}
