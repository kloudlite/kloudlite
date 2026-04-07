import { useRef, useEffect } from 'react'

interface CodeEditorProps {
  value: string
  onChange?: (value: string) => void
  height?: string
  readOnly?: boolean
}

const FONT = 'SF Mono, Fira Code, JetBrains Mono, Menlo, Consolas, monospace'

// Basic YAML syntax highlighting — returns HTML with spans
function highlightYaml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    // Comments
    .replace(/(#.*)$/gm, '<span style="color:#5c6370">$1</span>')
    // Strings (double and single quoted)
    .replace(/("[^"]*"|'[^']*')/g, '<span style="color:#98c379">$1</span>')
    // Keys (before colon at start of line or after indent)
    .replace(/^(\s*)([\w-]+)(:)/gm, '$1<span style="color:#e06c75">$2</span><span style="color:#abb2bf">$3</span>')
    // List markers
    .replace(/^(\s*)(-)(\s)/gm, '$1<span style="color:#56b6c2">$2</span>$3')
    // Numbers
    .replace(/\b(\d+)\b/g, '<span style="color:#d19a66">$1</span>')
    // Booleans and nulls
    .replace(/\b(true|false|null|yes|no)\b/g, '<span style="color:#d19a66">$1</span>')
}

export function CodeEditor({ value, onChange, height = '300px', readOnly = false }: CodeEditorProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const highlightRef = useRef<HTMLPreElement>(null)
  const lineNumRef = useRef<HTMLDivElement>(null)
  const lines = value.split('\n')

  function handleScroll() {
    if (textareaRef.current && highlightRef.current && lineNumRef.current) {
      const scrollTop = textareaRef.current.scrollTop
      const scrollLeft = textareaRef.current.scrollLeft
      highlightRef.current.scrollTop = scrollTop
      highlightRef.current.scrollLeft = scrollLeft
      lineNumRef.current.scrollTop = scrollTop
    }
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === 'Tab') {
      e.preventDefault()
      const ta = e.currentTarget
      const start = ta.selectionStart
      const end = ta.selectionEnd
      const newValue = value.substring(0, start) + '  ' + value.substring(end)
      onChange?.(newValue)
      requestAnimationFrame(() => {
        ta.selectionStart = ta.selectionEnd = start + 2
      })
    }
  }

  return (
    <div className="relative flex overflow-hidden bg-[#282c34]" style={{ height, fontFamily: FONT }}>
      {/* Line numbers */}
      <div
        ref={lineNumRef}
        className="shrink-0 select-none overflow-hidden py-2 pr-3 pl-4 text-right"
        style={{ width: '52px' }}
      >
        {lines.map((_, i) => (
          <div key={i} className="text-[13px] leading-[22px] text-[#636d83]">
            {i + 1}
          </div>
        ))}
      </div>

      {/* Editor + highlight overlay */}
      <div className="relative flex-1 overflow-hidden">
        {/* Highlight layer */}
        <pre
          ref={highlightRef}
          aria-hidden
          className="pointer-events-none absolute inset-0 m-0 overflow-auto whitespace-pre py-2 pl-0 pr-4 text-[13px] leading-[22px] text-[#abb2bf]"
          style={{ fontFamily: FONT }}
          dangerouslySetInnerHTML={{ __html: highlightYaml(value) + '\n' }}
        />

        {/* Textarea — transparent text, caret visible */}
        <textarea
          ref={textareaRef}
          value={value}
          onChange={(e) => onChange?.(e.target.value)}
          onScroll={handleScroll}
          onKeyDown={handleKeyDown}
          readOnly={readOnly}
          spellCheck={false}
          className="absolute inset-0 h-full w-full resize-none overflow-auto bg-transparent py-2 pl-0 pr-4 text-[13px] leading-[22px] text-transparent caret-white outline-none"
          style={{ fontFamily: FONT, tabSize: 2 }}
        />
      </div>
    </div>
  )
}
