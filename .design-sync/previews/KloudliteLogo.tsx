import * as React from 'react'
import { KloudliteLogo } from '@kloudlite/ui'

// The logo SVG carries an intrinsic height of 22px, so previews scale it with
// an inline zoom instead of a utility class.
const Scale = ({ zoom, children }: { zoom: number; children: React.ReactNode }) => (
  <div style={{ zoom }}>{children}</div>
)

export const FullLockup = () => (
  <Scale zoom={3}>
    <KloudliteLogo linkToHome={false} />
  </Scale>
)

export const MarkOnly = () => (
  <Scale zoom={4}>
    <KloudliteLogo linkToHome={false} showText={false} />
  </Scale>
)

export const OnDarkSurface = () => (
  <div className="flex items-center justify-center rounded-md bg-foreground p-8">
    <Scale zoom={3}>
      <KloudliteLogo linkToHome={false} variant="white" />
    </Scale>
  </div>
)

export const InAppHeader = () => (
  <div className="flex w-96 items-center justify-between rounded-md border border-border bg-card p-4">
    <Scale zoom={1.5}>
      <KloudliteLogo linkToHome={false} />
    </Scale>
    <span className="text-sm text-muted-foreground">Console</span>
  </div>
)

export const SizeSweep = () => (
  <div className="flex flex-col gap-6">
    <Scale zoom={1}>
      <KloudliteLogo linkToHome={false} />
    </Scale>
    <Scale zoom={2}>
      <KloudliteLogo linkToHome={false} />
    </Scale>
    <Scale zoom={3}>
      <KloudliteLogo linkToHome={false} />
    </Scale>
  </div>
)
