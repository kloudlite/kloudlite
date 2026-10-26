import * as React from 'react'
import {
  Button,
  ButtonGroup,
  ButtonGroupSeparator,
  ButtonGroupText,
} from '@kloudlite/ui'

export const Horizontal = () => (
  <ButtonGroup>
    <Button variant="outline">Logs</Button>
    <Button variant="outline">Metrics</Button>
    <Button variant="outline">Events</Button>
  </ButtonGroup>
)

export const Vertical = () => (
  <ButtonGroup orientation="vertical">
    <Button variant="outline">Restart deployment</Button>
    <Button variant="outline">Scale node pool</Button>
    <Button variant="outline">Rotate credentials</Button>
  </ButtonGroup>
)

export const WithText = () => (
  <ButtonGroup>
    <ButtonGroupText>Replicas</ButtonGroupText>
    <Button variant="outline">&minus;</Button>
    <Button variant="outline">3</Button>
    <Button variant="outline">+</Button>
  </ButtonGroup>
)

export const WithSeparator = () => (
  <ButtonGroup>
    <Button variant="outline">Deploy</Button>
    <ButtonGroupSeparator />
    <Button variant="outline">Rollback</Button>
    <ButtonGroupSeparator />
    <Button variant="outline">Pause</Button>
  </ButtonGroup>
)

export const SelectedState = () => (
  <ButtonGroup>
    <Button variant="default">ap-south-1</Button>
    <Button variant="outline">eu-west-1</Button>
    <Button variant="outline" disabled>
      us-east-1
    </Button>
  </ButtonGroup>
)
