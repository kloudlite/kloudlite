import { useState, useRef, useEffect, type KeyboardEvent } from 'react'
import { ArrowLeft, ArrowRight, RotateCw } from 'lucide-react'
import { cn } from '@kloudlite/lib'
import { useTabStore } from '@/store/tabs'

function normalizeUrl(input: string): string {
  const trimmed = input.trim()
  if (!trimmed) return ''
  if (/^https?:\/\//i.test(trimmed)) return trimmed
  if (trimmed.includes('.') && !trimmed.includes(' ')) {
    return `https://${trimmed}`
  }
  return `https://www.google.com/search?q=${encodeURIComponent(trimmed)}`
}

interface AddressBarProps {
  onNavigate: (url: string) => void
  onGoBack: () => void
  onGoForward: () => void
  onReload: () => void
}

export function AddressBar({ onNavigate, onGoBack, onGoForward, onReload }: AddressBarProps) {
  const { tabs, activeTabId } = useTabStore()
  const activeTab = tabs.find((t) => t.id === activeTabId)
  const [inputValue, setInputValue] = useState('')
  const [isEditing, setIsEditing] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (!isEditing && activeTab) {
      setInputValue(activeTab.url)
    }
  }, [activeTab?.url, activeTab?.id, isEditing])

  function handleKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter') {
      const url = normalizeUrl(inputValue)
      if (url) {
        onNavigate(url)
        setIsEditing(false)
        inputRef.current?.blur()
      }
    } else if (e.key === 'Escape') {
      setIsEditing(false)
      setInputValue(activeTab?.url ?? '')
      inputRef.current?.blur()
    }
  }

  if (!activeTab) return null

  return (
    <div className="no-drag flex h-10 items-center gap-1 border-b border-border bg-background px-2">
      <button
        className={cn(
          'rounded-sm p-1.5 transition-colors',
          activeTab.canGoBack
            ? 'text-foreground hover:bg-accent'
            : 'cursor-default text-muted-foreground/40'
        )}
        onClick={onGoBack}
        disabled={!activeTab.canGoBack}
      >
        <ArrowLeft className="h-4 w-4" />
      </button>
      <button
        className={cn(
          'rounded-sm p-1.5 transition-colors',
          activeTab.canGoForward
            ? 'text-foreground hover:bg-accent'
            : 'cursor-default text-muted-foreground/40'
        )}
        onClick={onGoForward}
        disabled={!activeTab.canGoForward}
      >
        <ArrowRight className="h-4 w-4" />
      </button>
      <button
        className="rounded-sm p-1.5 text-foreground transition-colors hover:bg-accent"
        onClick={onReload}
      >
        <RotateCw className={cn('h-4 w-4', activeTab.isLoading && 'animate-spin')} />
      </button>

      <div className="flex flex-1 items-center rounded-sm border border-input bg-muted/50 px-3 py-1">
        <input
          ref={inputRef}
          type="text"
          className="w-full bg-transparent text-sm text-foreground outline-none placeholder:text-muted-foreground"
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onFocus={() => {
            setIsEditing(true)
            setTimeout(() => inputRef.current?.select(), 0)
          }}
          onBlur={() => {
            setIsEditing(false)
            setInputValue(activeTab.url)
          }}
          onKeyDown={handleKeyDown}
          placeholder="Enter URL or search..."
          spellCheck={false}
        />
      </div>
    </div>
  )
}
