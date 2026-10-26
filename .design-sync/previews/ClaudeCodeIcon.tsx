import * as React from 'react'
import { ClaudeCodeIcon } from '@kloudlite/ui'

export const Default = () => (
  <div className="flex items-center justify-center border border-border bg-card p-6 text-card-foreground">
    <ClaudeCodeIcon className="w-24 h-24" />
  </div>
)

export const Sizes = () => (
  <div className="flex items-center gap-6 border border-border bg-card p-6 text-card-foreground">
    {[["'w-6 h-6'", "'24'"], ["'w-8 h-8'", "'32'"], ["'w-12 h-12'", "'48'"], ["'w-16 h-16'", "'64'"]].map(([cls, px]) => (
      <div key={px} className="flex flex-col items-center gap-2">
        <ClaudeCodeIcon className={cls} />
        <span className="text-xs text-muted-foreground">{px}px</span>
      </div>
    ))}
  </div>
)

export const OnLightSurface = () => (
  <div className="flex items-center gap-4 border border-border bg-background p-6 text-foreground">
    <ClaudeCodeIcon className="w-16 h-16" />
    <span className="text-sm font-medium">Claude Code &middot; workspace editor</span>
  </div>
)

export const OnDarkSurface = () => (
  <div className="flex items-center gap-4 border border-border bg-foreground p-6 text-background">
    <ClaudeCodeIcon className="w-16 h-16" />
    <span className="text-sm font-medium">Claude Code &middot; inverted surface</span>
  </div>
)

export const InlineWithText = () => (
  <div className="flex items-center gap-2 border border-border bg-card p-4 text-card-foreground">
    <ClaudeCodeIcon className="w-6 h-6" />
    <span className="text-sm">Launch <strong>Claude Code</strong> on env <code>dev-karthik</code></span>
  </div>
)
