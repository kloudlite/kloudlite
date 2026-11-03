import * as React from 'react'
import { Button, ButtonGroup, ButtonGroupSeparator } from '@kloudlite/ui'

export const BetweenButtons = () => (
  <ButtonGroup>
    <Button variant="outline">Deploy</Button>
    <ButtonGroupSeparator />
    <Button variant="outline">Rollback</Button>
  </ButtonGroup>
)

export const MultipleSeparators = () => (
  <ButtonGroup>
    <Button variant="outline">Logs</Button>
    <ButtonGroupSeparator />
    <Button variant="outline">Metrics</Button>
    <ButtonGroupSeparator />
    <Button variant="outline">Events</Button>
  </ButtonGroup>
)

export const HorizontalOrientation = () => (
  <ButtonGroup orientation="vertical">
    <Button variant="outline">Start intercept</Button>
    <ButtonGroupSeparator orientation="horizontal" />
    <Button variant="outline">Stop intercept</Button>
  </ButtonGroup>
)
