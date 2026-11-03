import * as React from 'react'
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuLabel,
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

export const Basic = () => (
  <Demo label="api-gateway &middot; deployment">
    <ContextMenuContent className="w-64">
      <ContextMenuLabel>api-gateway</ContextMenuLabel>
      <ContextMenuSeparator />
      <ContextMenuItem>View logs</ContextMenuItem>
      <ContextMenuItem>Restart deployment</ContextMenuItem>
    </ContextMenuContent>
  </Demo>
)

export const Inset = () => (
  <Demo label="Workload list">
    <ContextMenuContent className="w-64">
      <ContextMenuLabel inset>Sort by</ContextMenuLabel>
      <ContextMenuItem inset>Name</ContextMenuItem>
      <ContextMenuItem inset>Last deployed</ContextMenuItem>
    </ContextMenuContent>
  </Demo>
)

export const MultipleSections = () => (
  <Demo label="kloudlite-dev &middot; environment">
    <ContextMenuContent className="w-64">
      <ContextMenuLabel>Environment</ContextMenuLabel>
      <ContextMenuItem>Open in workspace</ContextMenuItem>
      <ContextMenuSeparator />
      <ContextMenuLabel>Danger zone</ContextMenuLabel>
      <ContextMenuItem>Delete environment</ContextMenuItem>
    </ContextMenuContent>
  </Demo>
)
