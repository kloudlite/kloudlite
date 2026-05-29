import { useState, useCallback, type ReactNode } from 'react'

interface DialogProps {
  title: string
  description?: string
  onClose: () => void
  children: ReactNode | ((close: () => void) => ReactNode)
  footer?: ReactNode | ((close: () => void) => ReactNode)
  maxWidth?: string
}

export function Dialog({ title, description, onClose, children, footer, maxWidth = '28rem' }: DialogProps) {
  const [exiting, setExiting] = useState(false)

  const close = useCallback(() => {
    setExiting(true)
    setTimeout(onClose, 150)
  }, [onClose])

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30" onClick={close}>
      <div
        className="w-full overflow-hidden rounded-2xl border border-border/40 bg-popover shadow-2xl"
        style={{
          maxWidth,
          animation: exiting ? 'popover-out 150ms ease-in forwards' : 'popover-in 150ms ease-out'
        }}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="border-b border-border/30 px-6 py-4">
          <h2 className="text-[16px] font-semibold text-foreground">{title}</h2>
          {description && (
            <p className="mt-0.5 text-[12px] text-muted-foreground">{description}</p>
          )}
        </div>

        <div className="px-6 py-5">
          {typeof children === 'function' ? children(close) : children}
        </div>

        {footer && (
          <div className="flex justify-end gap-2 border-t border-border/30 px-6 py-4">
            {typeof footer === 'function' ? footer(close) : footer}
          </div>
        )}
      </div>
    </div>
  )
}
