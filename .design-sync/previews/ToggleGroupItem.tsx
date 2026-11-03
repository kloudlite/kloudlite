import * as React from 'react'
import { ToggleGroup, ToggleGroupItem } from '@kloudlite/ui'

export const SelectedItem = () => (
  <ToggleGroup type="single" defaultValue="grid" variant="outline">
    <ToggleGroupItem value="list">List</ToggleGroupItem>
    <ToggleGroupItem value="grid">Grid</ToggleGroupItem>
  </ToggleGroup>
)

export const UnselectedItem = () => (
  <ToggleGroup type="single" defaultValue="list" variant="outline">
    <ToggleGroupItem value="list">List</ToggleGroupItem>
    <ToggleGroupItem value="grid">Grid</ToggleGroupItem>
  </ToggleGroup>
)

export const DisabledItem = () => (
  <ToggleGroup type="single" defaultValue="logs" variant="outline">
    <ToggleGroupItem value="logs">Logs</ToggleGroupItem>
    <ToggleGroupItem value="traces" disabled>
      Traces &middot; upgrade
    </ToggleGroupItem>
  </ToggleGroup>
)

export const MultipleSelected = () => (
  <ToggleGroup type="multiple" defaultValue={['warn', 'error']} variant="outline">
    <ToggleGroupItem value="info">Info</ToggleGroupItem>
    <ToggleGroupItem value="warn">Warn</ToggleGroupItem>
    <ToggleGroupItem value="error">Error</ToggleGroupItem>
  </ToggleGroup>
)
