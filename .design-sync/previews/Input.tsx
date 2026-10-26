import * as React from 'react'
import { Input, Label } from '@kloudlite/ui'

export const Default = () => (
  <div className="w-80">
    <Input defaultValue="ghcr.io/kloudlite/api:v1.4.2" />
  </div>
)

export const WithLabelAndPlaceholder = () => (
  <div className="w-80 space-y-2">
    <Label htmlFor="env-name">Environment name</Label>
    <Input id="env-name" placeholder="dev-karthik" />
  </div>
)

export const Types = () => (
  <div className="w-80 space-y-4">
    <div className="space-y-2">
      <Label htmlFor="in-email">Invite email</Label>
      <Input id="in-email" type="email" defaultValue="karthik@kloudlite.io" />
    </div>
    <div className="space-y-2">
      <Label htmlFor="in-pass">Registry password</Label>
      <Input id="in-pass" type="password" defaultValue="kloudlite-registry" />
    </div>
    <div className="space-y-2">
      <Label htmlFor="in-port">Container port</Label>
      <Input id="in-port" type="number" defaultValue={8080} />
    </div>
    <div className="space-y-2">
      <Label htmlFor="in-file">Kubeconfig file</Label>
      <Input id="in-file" type="file" />
    </div>
  </div>
)

export const Disabled = () => (
  <div className="w-80 space-y-2">
    <Label htmlFor="in-cluster">Cluster</Label>
    <Input id="in-cluster" disabled defaultValue="kl-prod-ap-south-1" />
  </div>
)

export const Invalid = () => (
  <div className="w-80 space-y-2">
    <Label htmlFor="in-branch">Git branch</Label>
    <Input id="in-branch" aria-invalid defaultValue="feature/ Deployments" />
    <p className="text-xs text-destructive">Branch names cannot contain spaces.</p>
  </div>
)
