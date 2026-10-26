import * as React from 'react'
import {
  ContextMenu,
  ContextMenuCheckboxItem,
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

export const Checked = () => (
  <Demo label="Workload list">
    <ContextMenuContent className="w-64">
      <ContextMenuLabel>View options</ContextMenuLabel>
      <ContextMenuSeparator />
      <ContextMenuCheckboxItem checked>Show system namespaces</ContextMenuCheckboxItem>
      <ContextMenuCheckboxItem checked>Group by workload type</ContextMenuCheckboxItem>
    </ContextMenuContent>
  </Demo>
)

export const Mixed = () => (
  <Demo label="Workload list">
    <ContextMenuContent className="w-64">
      <ContextMenuCheckboxItem checked>Show system namespaces</ContextMenuCheckboxItem>
      <ContextMenuCheckboxItem>Show completed pods</ContextMenuCheckboxItem>
      <ContextMenuCheckboxItem>Show resource usage</ContextMenuCheckboxItem>
    </ContextMenuContent>
  </Demo>
)

export const WithDisabled = () => (
  <Demo label="Log stream">
    <ContextMenuContent className="w-64">
      <ContextMenuCheckboxItem checked>Follow tail</ContextMenuCheckboxItem>
      <ContextMenuCheckboxItem>Wrap long lines</ContextMenuCheckboxItem>
      <ContextMenuCheckboxItem disabled checked>Include previous container</ContextMenuCheckboxItem>
    </ContextMenuContent>
  </Demo>
)
