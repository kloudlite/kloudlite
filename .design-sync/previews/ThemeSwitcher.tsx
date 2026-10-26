import * as React from 'react'
import { Badge, ThemeSwitcher } from '@kloudlite/ui'

export const Basic = () => (
  <div className="flex items-center gap-2 rounded-md border p-4">
    <ThemeSwitcher />
    <span className="text-sm text-muted-foreground">Appearance</span>
  </div>
)

export const InAppHeader = () => (
  <div className="w-96 rounded-md border bg-card">
    <div className="flex items-center justify-between gap-3 border-b px-4 py-2">
      <span className="text-sm font-medium text-foreground">kloudlite / production</span>
      <div className="flex items-center gap-2">
        <Badge variant="secondary">eu-west-1</Badge>
        <ThemeSwitcher />
      </div>
    </div>
    <div className="p-4">
      <p className="text-sm text-muted-foreground">
        12 workspaces reconciled &middot; last sync 4 minutes ago
      </p>
    </div>
  </div>
)

export const WithServerAction = () => (
  <div className="flex w-80 items-center justify-between gap-3 rounded-md border p-4">
    <div className="space-y-2">
      <p className="text-sm font-medium text-foreground">Theme</p>
      <p className="text-xs text-muted-foreground">
        Persisted to a cookie so SSR renders the right palette.
      </p>
    </div>
    <ThemeSwitcher setThemeCookie={async () => {}} />
  </div>
)
