import * as React from 'react'
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
  InputGroupText,
  InputGroupTextarea,
  Label,
} from '@kloudlite/ui'

export const Prefix = () => (
  <div className="w-96 space-y-2">
    <Label htmlFor="igt-1">Registry image</Label>
    <InputGroup>
      <InputGroupAddon align="inline-start">
        <InputGroupText>ghcr.io/</InputGroupText>
      </InputGroupAddon>
      <InputGroupInput id="igt-1" defaultValue="kloudlite/api:v1.4.2" />
    </InputGroup>
  </div>
)

export const Suffix = () => (
  <div className="w-96 space-y-2">
    <Label htmlFor="igt-2">Workspace subdomain</Label>
    <InputGroup>
      <InputGroupInput id="igt-2" defaultValue="dev-karthik" />
      <InputGroupAddon align="inline-end">
        <InputGroupText>.kloudlite.io</InputGroupText>
      </InputGroupAddon>
    </InputGroup>
  </div>
)

export const BothEnds = () => (
  <div className="w-96 space-y-2">
    <Label htmlFor="igt-3">Port forward</Label>
    <InputGroup>
      <InputGroupAddon align="inline-start">
        <InputGroupText>localhost:</InputGroupText>
      </InputGroupAddon>
      <InputGroupInput id="igt-3" defaultValue="8080" />
      <InputGroupAddon align="inline-end">
        <InputGroupText>tcp</InputGroupText>
      </InputGroupAddon>
    </InputGroup>
  </div>
)

export const AsHelperRow = () => (
  <div className="w-96">
    <InputGroup>
      <InputGroupInput defaultValue="release/v1.4" />
      <InputGroupAddon align="block-end">
        <InputGroupText>Tracking origin/release/v1.4</InputGroupText>
      </InputGroupAddon>
    </InputGroup>
  </div>
)
