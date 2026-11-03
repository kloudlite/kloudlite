import * as React from 'react'
import { Button, ButtonGroup, ButtonGroupText, Input } from '@kloudlite/ui'

export const LeadingLabel = () => (
  <ButtonGroup>
    <ButtonGroupText>Region</ButtonGroupText>
    <Button variant="outline">ap-south-1</Button>
  </ButtonGroup>
)

export const AroundInput = () => (
  <ButtonGroup className="w-96">
    <ButtonGroupText>https://</ButtonGroupText>
    <Input defaultValue="payments-api" />
    <ButtonGroupText>.kloudlite.io</ButtonGroupText>
  </ButtonGroup>
)

export const AsChildLabel = () => (
  <ButtonGroup>
    <ButtonGroupText asChild>
      <label htmlFor="replicas">Replicas</label>
    </ButtonGroupText>
    <Button id="replicas" variant="outline">
      3
    </Button>
  </ButtonGroup>
)

export const TrailingUnit = () => (
  <ButtonGroup>
    <Button variant="outline">Memory</Button>
    <ButtonGroupText>2048 MiB</ButtonGroupText>
  </ButtonGroup>
)
