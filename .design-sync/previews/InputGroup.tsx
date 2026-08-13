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
  <div className="w-96">
    <InputGroup>
      <InputGroupAddon align="inline-start">
        <InputGroupText>ghcr.io/</InputGroupText>
      </InputGroupAddon>
      <InputGroupInput defaultValue="kloudlite/api" />
      <InputGroupAddon align="inline-end">
        <InputGroupText>:v1.4.2</InputGroupText>
      </InputGroupAddon>
    </InputGroup>
  </div>
)

export const WithButton = () => (
  <div className="w-96 space-y-2">
    <Label htmlFor="ig-kubeconfig">Kubeconfig path</Label>
    <InputGroup>
      <InputGroupInput id="ig-kubeconfig" defaultValue="~/.kube/kl-prod.yaml" />
      <InputGroupAddon align="inline-end">
        <InputGroupButton size="sm" variant="ghost">Copy</InputGroupButton>
      </InputGroupAddon>
    </InputGroup>
  </div>
)

export const WithTextarea = () => (
  <div className="w-96">
    <InputGroup>
      <InputGroupAddon align="block-start">
        <InputGroupText>Deployment notes</InputGroupText>
      </InputGroupAddon>
      <InputGroupTextarea defaultValue={"Rolled api to v1.4.2 on ap-south-1.\nWatch p99 latency for 30m."} rows={3} />
    </InputGroup>
  </div>
)

export const Invalid = () => (
  <div className="w-96 space-y-2">
    <Label htmlFor="ig-port">Container port</Label>
    <InputGroup>
      <InputGroupAddon align="inline-start">
        <InputGroupText>tcp</InputGroupText>
      </InputGroupAddon>
      <InputGroupInput id="ig-port" aria-invalid defaultValue="98765" />
    </InputGroup>
    <p className="text-xs text-destructive">Port must be between 1 and 65535.</p>
  </div>
)
