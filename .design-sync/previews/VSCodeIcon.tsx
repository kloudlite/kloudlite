import * as React from 'react'
import { VSCodeIcon } from '@kloudlite/ui'

export const Default = () => (
  <div className="flex items-center justify-center border border-border bg-card p-6 text-card-foreground">
    <VSCodeIcon className="w-24 h-24" />
  </div>
)

export const Sizes = () => (
  <div className="flex items-center gap-6 border border-border bg-card p-6 text-card-foreground">
    {[["'w-6 h-6'", "'24'"], ["'w-8 h-8'", "'32'"], ["'w-12 h-12'", "'48'"], ["'w-16 h-16'", "'64'"]].map(([cls, px]) => (
      <div key={px} className="flex flex-col items-center gap-2">
        <VSCodeIcon className={cls} />
        <span className="text-xs text-muted-foreground">{px}px</span>
      </div>
    ))}
  </div>
)

export const OnLightSurface = () => (
  <div className="flex items-center gap-4 border border-border bg-background p-6 text-foreground">
    <VSCodeIcon className="w-16 h-16" />
    <span className="text-sm font-medium">VS Code &middot; workspace editor</span>
  </div>
)

export const OnDarkSurface = () => (
  <div className="flex items-center gap-4 border border-border bg-foreground p-6 text-background">
    <VSCodeIcon className="w-16 h-16" />
    <span className="text-sm font-medium">VS Code &middot; inverted surface</span>
  </div>
)

export const InlineWithText = () => (
  <div className="flex items-center gap-2 border border-border bg-card p-4 text-card-foreground">
    <VSCodeIcon className="w-6 h-6" />
    <span className="text-sm">Launch <strong>VS Code</strong> on env <code>dev-karthik</code></span>
  </div>
)
