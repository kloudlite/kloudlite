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

export const Default = () => (
  <div className="w-96 space-y-2">
    <Label htmlFor="igi-1">Environment name</Label>
    <InputGroup>
      <InputGroupAddon align="inline-start">
        <InputGroupText>env/</InputGroupText>
      </InputGroupAddon>
      <InputGroupInput id="igi-1" defaultValue="dev-karthik" />
    </InputGroup>
  </div>
)

export const Placeholder = () => (
  <div className="w-96 space-y-2">
    <Label htmlFor="igi-2">Image tag</Label>
    <InputGroup>
      <InputGroupAddon align="inline-start">
        <InputGroupText>ghcr.io/</InputGroupText>
      </InputGroupAddon>
      <InputGroupInput id="igi-2" placeholder="kloudlite/api:v1.4.2" />
    </InputGroup>
  </div>
)

export const Disabled = () => (
  <div className="w-96 space-y-2">
    <Label htmlFor="igi-3">Cluster</Label>
    <InputGroup>
      <InputGroupAddon align="inline-start">
        <InputGroupText>k8s</InputGroupText>
      </InputGroupAddon>
      <InputGroupInput id="igi-3" disabled defaultValue="kl-prod-ap-south-1" />
    </InputGroup>
  </div>
)

export const Invalid = () => (
  <div className="w-96 space-y-2">
    <Label htmlFor="igi-4">Invite email</Label>
    <InputGroup>
      <InputGroupAddon align="inline-start">
        <InputGroupText>to</InputGroupText>
      </InputGroupAddon>
      <InputGroupInput id="igi-4" type="email" aria-invalid defaultValue="karthik@kloudlite" />
    </InputGroup>
    <p className="text-xs text-destructive">Enter a valid email address.</p>
  </div>
)
