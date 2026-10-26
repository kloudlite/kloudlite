import * as React from 'react'
import { ToggleGroup, ToggleGroupItem } from '@kloudlite/ui'

export const Basic = () => (
  <ToggleGroup type="single" defaultValue="logs" variant="outline">
    <ToggleGroupItem value="logs">Logs</ToggleGroupItem>
    <ToggleGroupItem value="events">Events</ToggleGroupItem>
    <ToggleGroupItem value="metrics">Metrics</ToggleGroupItem>
  </ToggleGroup>
)

export const SingleViewSwitcher = () => (
  <div className="flex w-96 items-center justify-between gap-3 rounded-md border p-4">
    <span className="text-sm font-medium text-foreground">Workspaces</span>
    <ToggleGroup type="single" defaultValue="grid" variant="outline" size="sm">
      <ToggleGroupItem value="list">List</ToggleGroupItem>
      <ToggleGroupItem value="grid">Grid</ToggleGroupItem>
    </ToggleGroup>
  </div>
)

export const Multiple = () => (
  <ToggleGroup type="multiple" defaultValue={['warn', 'error']} variant="outline">
    <ToggleGroupItem value="debug">Debug</ToggleGroupItem>
    <ToggleGroupItem value="info">Info</ToggleGroupItem>
    <ToggleGroupItem value="warn">Warn</ToggleGroupItem>
    <ToggleGroupItem value="error">Error</ToggleGroupItem>
  </ToggleGroup>
)

export const Sizes = () => (
  <div className="flex flex-col gap-3">
    <ToggleGroup type="single" defaultValue="cpu" variant="outline" size="sm">
      <ToggleGroupItem value="cpu">CPU</ToggleGroupItem>
      <ToggleGroupItem value="memory">Memory</ToggleGroupItem>
      <ToggleGroupItem value="network">Network</ToggleGroupItem>
    </ToggleGroup>
    <ToggleGroup type="single" defaultValue="cpu" variant="outline" size="lg">
      <ToggleGroupItem value="cpu">CPU</ToggleGroupItem>
      <ToggleGroupItem value="memory">Memory</ToggleGroupItem>
      <ToggleGroupItem value="network">Network</ToggleGroupItem>
    </ToggleGroup>
  </div>
)

export const WithDisabledItem = () => (
  <ToggleGroup type="single" defaultValue="logs" variant="outline">
    <ToggleGroupItem value="logs">Logs</ToggleGroupItem>
    <ToggleGroupItem value="events">Events</ToggleGroupItem>
    <ToggleGroupItem value="traces" disabled>
      Traces
    </ToggleGroupItem>
  </ToggleGroup>
)
