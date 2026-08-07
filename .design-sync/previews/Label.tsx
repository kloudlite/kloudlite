import * as React from 'react'
import { Checkbox, Input, Label, Switch } from '@kloudlite/ui'

export const WithInput = () => (
  <div className="w-80">
    <Label htmlFor="env-name">Environment name</Label>
    <Input id="env-name" defaultValue="staging-eu" />
  </div>
)

export const WithHelperText = () => (
  <div className="w-80">
    <Label htmlFor="cluster-region">Region</Label>
    <Input id="cluster-region" defaultValue="ap-south-1" />
    <p className="mt-1.5 text-sm text-muted-foreground">
      Workspaces are scheduled in this region and cannot be moved later.
    </p>
  </div>
)

export const WithCheckbox = () => (
  <div className="flex items-center gap-2">
    <Checkbox id="auto-suspend" defaultChecked />
    <Label htmlFor="auto-suspend" className="mb-0">
      Suspend idle workspaces after 30 minutes
    </Label>
  </div>
)

export const WithSwitch = () => (
  <div className="flex w-96 items-center justify-between">
    <Label htmlFor="public-endpoint" className="mb-0">
      Expose environment on a public endpoint
    </Label>
    <Switch id="public-endpoint" defaultChecked />
  </div>
)

export const FormSection = () => (
  <div className="w-96 space-y-4 rounded-md border border-border p-6">
    <div>
      <Label htmlFor="cluster-name">Cluster name</Label>
      <Input id="cluster-name" defaultValue="kloudlite-prod" />
    </div>
    <div>
      <Label htmlFor="node-pool">Node pool size</Label>
      <Input id="node-pool" defaultValue="4 nodes · 8 vCPU" />
    </div>
    <div className="flex items-center gap-2">
      <Checkbox id="enable-backups" defaultChecked />
      <Label htmlFor="enable-backups" className="mb-0">
        Enable nightly backups
      </Label>
    </div>
  </div>
)
