import { useState, useCallback, useEffect, type ReactNode } from 'react'
import { cn } from '@/lib/utils'

interface SheetProps {
  side?: 'right' | 'left' | 'bottom'
  onClose: () => void
  children: ReactNode | ((close: () => void) => ReactNode)
  width?: string
  height?: string
}

const animations = {
  right: {
    in: 'sheet-in-right 250ms cubic-bezier(0.32, 0.72, 0, 1)',
    out: 'sheet-out-right 200ms cubic-bezier(0.32, 0, 0.67, 0) forwards',
  },
  left: {
    in: 'sheet-in-left 250ms cubic-bezier(0.32, 0.72, 0, 1)',
    out: 'sheet-out-left 200ms cubic-bezier(0.32, 0, 0.67, 0) forwards',
  },
  bottom: {
    in: 'sheet-in-bottom 250ms cubic-bezier(0.32, 0.72, 0, 1)',
    out: 'sheet-out-bottom 200ms cubic-bezier(0.32, 0, 0.67, 0) forwards',
  },
}

export function Sheet({ side = 'right', onClose, children, width = '720px', height = '70vh' }: SheetProps) {
  const [exiting, setExiting] = useState(false)

  const close = useCallback(() => {
    setExiting(true)
    setTimeout(onClose, 200)
  }, [onClose])

  useEffect(() => {
    function handleKey(e: KeyboardEvent) {
      if (e.key === 'Escape') close()
    }
    window.addEventListener('keydown', handleKey)
    return () => window.removeEventListener('keydown', handleKey)
  }, [close])

  const isHorizontal = side === 'left' || side === 'right'
  const anim = animations[side]

  return (
    <div
      className="fixed inset-0 z-50 bg-black/40"
      style={{
        animation: exiting ? 'fade-out 200ms ease-in forwards' : 'fade-in 200ms ease-out',
      }}
      onClick={close}
    >
      <div
        className={cn(
          'fixed border-border/50 bg-popover shadow-2xl',
          side === 'right' && 'right-0 top-0 h-full border-l',
          side === 'left' && 'left-0 top-0 h-full border-r',
          side === 'bottom' && 'bottom-0 left-0 right-0 border-t rounded-t-2xl'
        )}
        style={{
          width: isHorizontal ? width : undefined,
          height: side === 'bottom' ? height : undefined,
          animation: exiting ? anim.out : anim.in,
        }}
        onClick={(e) => e.stopPropagation()}
      >
        {typeof children === 'function' ? children(close) : children}
      </div>
    </div>
  )
}
