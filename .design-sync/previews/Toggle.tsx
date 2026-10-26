import * as React from 'react'
import { Toggle } from '@kloudlite/ui'

export const Basic = () => <Toggle defaultPressed>Follow logs</Toggle>

export const Variants = () => (
  <div className="flex items-center gap-3">
    <Toggle defaultPressed>Wrap lines</Toggle>
    <Toggle variant="outline" defaultPressed>
      Follow logs
    </Toggle>
    <Toggle variant="outline">Show timestamps</Toggle>
  </div>
)

export const Sizes = () => (
  <div className="flex items-center gap-3">
    <Toggle variant="outline" size="sm" defaultPressed>
      sm
    </Toggle>
    <Toggle variant="outline" size="default" defaultPressed>
      default
    </Toggle>
    <Toggle variant="outline" size="lg" defaultPressed>
      lg
    </Toggle>
  </div>
)

export const OnAndOff = () => (
  <div className="flex flex-col gap-3">
    <div className="flex items-center gap-3">
      <Toggle variant="outline" defaultPressed>
        Show system namespaces
      </Toggle>
      <span className="text-xs text-muted-foreground">on</span>
    </div>
    <div className="flex items-center gap-3">
      <Toggle variant="outline">Show system namespaces</Toggle>
      <span className="text-xs text-muted-foreground">off</span>
    </div>
  </div>
)

export const Disabled = () => (
  <div className="flex items-center gap-3">
    <Toggle variant="outline" defaultPressed disabled>
      Stream to stdout
    </Toggle>
    <Toggle disabled>Debug tracing</Toggle>
  </div>
)
