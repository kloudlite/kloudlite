import * as React from 'react'
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuLabel,
  ContextMenuSeparator,
  ContextMenuShortcut,
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
      <ContextMenuItem>
        View logs
        <ContextMenuShortcut>&#8984;L</ContextMenuShortcut>
      </ContextMenuItem>
      <ContextMenuItem>
        Restart deployment
        <ContextMenuShortcut>&#8984;R</ContextMenuShortcut>
      </ContextMenuItem>
    </ContextMenuContent>
  </Demo>
)

export const AcrossSections = () => (
  <Demo label="kloudlite-dev &middot; environment">
    <ContextMenuContent className="w-72">
      <ContextMenuLabel>kloudlite-dev</ContextMenuLabel>
      <ContextMenuSeparator />
      <ContextMenuItem>
        Open in workspace
        <ContextMenuShortcut>&#8984;&#8629;</ContextMenuShortcut>
      </ContextMenuItem>
      <ContextMenuItem>
        Copy kubeconfig
        <ContextMenuShortcut>&#8984;&#8679;K</ContextMenuShortcut>
      </ContextMenuItem>
      <ContextMenuSeparator />
      <ContextMenuItem>
        Delete environment
        <ContextMenuShortcut>&#8984;&#9003;</ContextMenuShortcut>
      </ContextMenuItem>
    </ContextMenuContent>
  </Demo>
)
