import * as React from 'react'
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuLabel,
  ContextMenuRadioGroup,
  ContextMenuRadioItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from '@kloudlite/ui'

// Radix's ContextMenu.Root exposes no `open`/`defaultOpen`, so a static render can
// never paint the menu. This fires the same `contextmenu` event a right-click would,
// once on mount. Note: `asChild` on ContextMenuTrigger drops Radix's props in this
// bundle, so the trigger is styled directly instead of wrapping a custom element.
function useAutoOpen() {
  const ref = React.useRef<HTMLSpanElement>(null)
  React.useEffect(() => {
    const el = ref.current
    if (!el) return
    const r = el.getBoundingClientRect()
    el.dispatchEvent(
      new MouseEvent('contextmenu', {
        bubbles: true,
        cancelable: true,
        clientX: Math.round(r.left + 24),
        clientY: Math.round(r.bottom - 2),
      })
    )
  }, [])
  return ref
}

const TRIGGER =
  'flex h-20 w-72 items-center justify-center border border-dashed border-input text-sm text-muted-foreground'

const Demo = ({ label, children }: { label: string; children: React.ReactNode }) => {
  const ref = useAutoOpen()
  return (
    <div className="h-96 w-96">
      <ContextMenu>
        <ContextMenuTrigger ref={ref} className={TRIGGER}>
          {label}
        </ContextMenuTrigger>
        {children}
      </ContextMenu>
    </div>
  )
}

export const LogLevel = () => (
  <Demo label="Log stream">
    <ContextMenuContent className="w-64">
      <ContextMenuLabel>Log level</ContextMenuLabel>
      <ContextMenuSeparator />
      <ContextMenuRadioGroup value="info">
        <ContextMenuRadioItem value="debug">Debug</ContextMenuRadioItem>
        <ContextMenuRadioItem value="info">Info</ContextMenuRadioItem>
        <ContextMenuRadioItem value="error">Error</ContextMenuRadioItem>
      </ContextMenuRadioGroup>
    </ContextMenuContent>
  </Demo>
)

export const Region = () => (
  <Demo label="Environment settings">
    <ContextMenuContent className="w-64">
      <ContextMenuLabel>Deploy region</ContextMenuLabel>
      <ContextMenuSeparator />
      <ContextMenuRadioGroup value="ap-south-1">
        <ContextMenuRadioItem value="ap-south-1">ap-south-1 &middot; Mumbai</ContextMenuRadioItem>
        <ContextMenuRadioItem value="eu-west-1">eu-west-1 &middot; Ireland</ContextMenuRadioItem>
        <ContextMenuRadioItem value="us-east-1" disabled>
          us-east-1 &middot; quota exceeded
        </ContextMenuRadioItem>
      </ContextMenuRadioGroup>
    </ContextMenuContent>
  </Demo>
)

export const WithinMenu = () => (
  <Demo label="api-gateway &middot; deployment">
    <ContextMenuContent className="w-64">
      <ContextMenuItem>View logs</ContextMenuItem>
      <ContextMenuSeparator />
      <ContextMenuLabel>Log level</ContextMenuLabel>
      <ContextMenuRadioGroup value="error">
        <ContextMenuRadioItem value="info">Info</ContextMenuRadioItem>
        <ContextMenuRadioItem value="error">Error</ContextMenuRadioItem>
      </ContextMenuRadioGroup>
    </ContextMenuContent>
  </Demo>
)
