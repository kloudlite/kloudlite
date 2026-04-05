import { useRef, useState, useLayoutEffect } from 'react'
import { X, Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { Tab } from '@/store/tabs'

interface TabItemProps {
  tab: Tab
  index: number
  isActive: boolean
  onSelect: () => void
  onClose: () => void
  onMove: (fromIndex: number, toIndex: number) => void
}

export function TabItem({ tab, index, isActive, onSelect, onClose, onMove }: TabItemProps) {
  const ref = useRef<HTMLDivElement>(null)
  const wrapperRef = useRef<HTMLDivElement>(null)
  const [dragOver, setDragOver] = useState<'above' | 'below' | null>(null)
  const [isDragging, setIsDragging] = useState(false)

  // FLIP: after DOM update, animate from old position to new
  useLayoutEffect(() => {
    const el = wrapperRef.current
    if (!el) return

    const prevRect = (el as any).__prevRect as DOMRect | undefined
    if (!prevRect) return
    delete (el as any).__prevRect

    const newRect = el.getBoundingClientRect()
    const deltaY = prevRect.top - newRect.top

    if (Math.abs(deltaY) < 1) return

    el.style.transform = `translateY(${deltaY}px)`
    el.style.transition = 'none'

    requestAnimationFrame(() => {
      el.style.transition = 'transform 200ms ease-out'
      el.style.transform = 'translateY(0)'
      const cleanup = () => {
        el.style.transform = ''
        el.style.transition = ''
        el.removeEventListener('transitionend', cleanup)
      }
      el.addEventListener('transitionend', cleanup)
    })
  })

  function handleDragStart(e: React.DragEvent) {
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', String(index))

    // Create a styled clone as drag image
    if (ref.current) {
      const clone = ref.current.cloneNode(true) as HTMLElement
      clone.style.position = 'fixed'
      clone.style.top = '-1000px'
      clone.style.opacity = '0.85'
      clone.style.background = 'rgba(255,255,255,0.12)'
      clone.style.borderRadius = '10px'
      clone.style.width = `${ref.current.offsetWidth}px`
      clone.style.pointerEvents = 'none'
      document.body.appendChild(clone)
      e.dataTransfer.setDragImage(clone, e.clientX - ref.current.getBoundingClientRect().left, e.clientY - ref.current.getBoundingClientRect().top)
      requestAnimationFrame(() => document.body.removeChild(clone))
    }

    requestAnimationFrame(() => setIsDragging(true))
  }

  function handleDragEnd() {
    setIsDragging(false)
  }

  function handleDragOver(e: React.DragEvent) {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
    const rect = ref.current?.getBoundingClientRect()
    if (rect) {
      const midY = rect.top + rect.height / 2
      setDragOver(e.clientY < midY ? 'above' : 'below')
    }
  }

  function handleDragLeave() {
    setDragOver(null)
  }

  function handleDrop(e: React.DragEvent) {
    e.preventDefault()
    setDragOver(null)
    const fromIndex = parseInt(e.dataTransfer.getData('text/plain'), 10)
    if (isNaN(fromIndex) || fromIndex === index) return
    const toIndex = dragOver === 'above' ? index : index + 1
    const adjustedTo = fromIndex < toIndex ? toIndex - 1 : toIndex
    // Capture all tab positions before the move for FLIP
    document.querySelectorAll('[data-tab-item]').forEach((el) => {
      (el as any).__prevRect = el.getBoundingClientRect()
    })
    onMove(fromIndex, adjustedTo)
  }

  return (
    <div ref={wrapperRef} data-tab-item className="relative mx-2.5">
      {/* Drop indicator — above */}
      {dragOver === 'above' && (
        <div className="absolute -top-[3px] left-2 right-2 z-10 flex items-center">
          <div className="h-1.5 w-1.5 rounded-full bg-sidebar-foreground/50" />
          <div className="h-[2px] flex-1 rounded-full bg-sidebar-foreground/50" />
          <div className="h-1.5 w-1.5 rounded-full bg-sidebar-foreground/50" />
        </div>
      )}

      <div
        ref={ref}
        className={cn(
          'group flex h-9 cursor-pointer items-center gap-2.5 rounded-[10px] px-3 text-[13px] transition-all duration-150',
          isActive
            ? 'bg-sidebar-foreground/[0.12] text-sidebar-foreground/90 font-medium'
            : 'text-sidebar-foreground/65 hover:bg-sidebar-foreground/[0.06] hover:text-sidebar-foreground/85',
          isDragging ? 'opacity-30' : 'transition-opacity duration-200'
        )}
        draggable
        onDragStart={handleDragStart}
        onDragEnd={handleDragEnd}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        onClick={onSelect}
      >
        {tab.isLoading ? (
          <Loader2 className="h-4 w-4 shrink-0 animate-spin opacity-50" />
        ) : tab.favicon ? (
          <img src={tab.favicon} alt="" className="h-4 w-4 shrink-0 rounded-sm" draggable={false} />
        ) : (
          <div className="h-4 w-4 shrink-0 rounded-full bg-sidebar-foreground/15" />
        )}
        <span className="min-w-0 flex-1 truncate">
          {tab.title || 'New Tab'}
        </span>
        <button
          className="shrink-0 rounded-md p-0.5 opacity-0 transition-all duration-150 hover:bg-sidebar-foreground/10 group-hover:opacity-100"
          onClick={(e) => {
            e.stopPropagation()
            onClose()
          }}
        >
          <X className="h-3 w-3" />
        </button>
      </div>

      {/* Drop indicator — below */}
      {dragOver === 'below' && (
        <div className="absolute -bottom-[3px] left-2 right-2 z-10 flex items-center">
          <div className="h-1.5 w-1.5 rounded-full bg-sidebar-foreground/50" />
          <div className="h-[2px] flex-1 rounded-full bg-sidebar-foreground/50" />
          <div className="h-1.5 w-1.5 rounded-full bg-sidebar-foreground/50" />
        </div>
      )}
    </div>
  )
}
