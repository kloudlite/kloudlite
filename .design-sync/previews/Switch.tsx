import * as React from 'react'
import { Label, Switch } from '@kloudlite/ui'

export const Basic = () => (
  <div className="flex items-center gap-3">
    <Switch id="autoscale" defaultChecked />
    <Label htmlFor="autoscale">Auto-scale node pool</Label>
  </div>
)

export const States = () => (
  <div className="w-80 space-y-4">
    <div className="flex items-center gap-3">
      <Switch id="on" defaultChecked />
      <Label htmlFor="on">Enabled</Label>
    </div>
    <div className="flex items-center gap-3">
      <Switch id="off" />
      <Label htmlFor="off">Disabled</Label>
    </div>
  </div>
)

export const SettingsList = () => (
  <div className="w-96 space-y-5 border p-4">
    <div className="flex items-center justify-between gap-4">
      <div className="space-y-1">
        <Label>Auto-scale node pool</Label>
        <p className="text-xs text-muted-foreground">Add nodes when CPU exceeds 80%</p>
      </div>
      <Switch defaultChecked />
    </div>
    <div className="flex items-center justify-between gap-4">
      <div className="space-y-1">
        <Label>Public ingress</Label>
        <p className="text-xs text-muted-foreground">Expose this environment on the internet</p>
      </div>
      <Switch />
    </div>
    <div className="flex items-center justify-between gap-4">
      <div className="space-y-1">
        <Label>Ship logs to S3</Label>
        <p className="text-xs text-muted-foreground">Archive workload logs nightly</p>
      </div>
      <Switch defaultChecked />
    </div>
  </div>
)

export const DisabledStates = () => (
  <div className="w-80 space-y-4">
    <div className="flex items-center gap-3">
      <Switch id="d-on" defaultChecked disabled />
      <Label htmlFor="d-on">Managed by your admin</Label>
    </div>
    <div className="flex items-center gap-3">
      <Switch id="d-off" disabled />
      <Label htmlFor="d-off">Not available on this plan</Label>
    </div>
  </div>
)
