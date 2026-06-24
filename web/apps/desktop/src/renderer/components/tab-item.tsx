import { useEffect, useRef, useState, useLayoutEffect } from 'react'
import { X } from 'lucide-react'
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
  const [open, setOpen] = useState(false)
  const [closing, setClosing] = useState(false)
  const [faviconFailed, setFaviconFailed] = useState(false)

  useEffect(() => {
    const id = requestAnimationFrame(() => setOpen(true))
    return () => cancelAnimationFrame(id)
  }, [])

  useEffect(() => {
    setFaviconFailed(false)
  }, [tab.favicon])

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

    el.animate(
      [{ transform: `translateY(${deltaY}px)` }, { transform: 'translateY(0)' }],
      { duration: 260, easing: 'cubic-bezier(0.22,1,0.36,1)' }
    )
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

  function closeWithAnimation() {
    if (closing) return
    setClosing(true)
    setOpen(false)
    setTimeout(onClose, 240)
  }

  return (
    <div
      ref={wrapperRef}
      data-tab-item
      className="relative mx-2.5 overflow-hidden"
      style={{
        height: open && !closing ? 44 : 0,
        opacity: open && !closing ? 1 : 0,
        transition: 'height 240ms cubic-bezier(0.22,1,0.36,1), opacity 180ms ease'
      }}
    >
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
          'group flex h-10 cursor-default items-center gap-3 rounded-[10px] px-3 text-[14px] transition-all duration-200',
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
        style={{
          transform: open && !closing ? 'translateY(0) scale(1)' : 'translateY(-6px) scale(0.98)',
          transition: 'transform 240ms cubic-bezier(0.22,1,0.36,1), background-color 150ms ease, color 150ms ease, opacity 180ms ease'
        }}
      >
        {tab.favicon && !faviconFailed && (
          <img
            src={tab.favicon}
            alt=""
            className="h-5 w-5 shrink-0 rounded-sm"
            draggable={false}
            onError={() => setFaviconFailed(true)}
          />
        )}
        <span className={cn('min-w-0 flex-1 truncate', tab.isLoading && 'loading-text')}>
          {tab.title || 'New Tab'}
        </span>
        <button
          className="shrink-0 rounded-md p-0.5 opacity-0 transition-all duration-150 hover:bg-sidebar-foreground/10 group-hover:opacity-100"
          onClick={(e) => {
            e.stopPropagation()
            closeWithAnimation()
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
