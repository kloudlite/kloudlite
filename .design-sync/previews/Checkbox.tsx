import * as React from 'react'
import { Checkbox, Label } from '@kloudlite/ui'

export const Basic = () => (
  <div className="flex items-center gap-2">
    <Checkbox id="auto-scale" defaultChecked />
    <Label htmlFor="auto-scale">Enable cluster autoscaling</Label>
  </div>
)

export const States = () => (
  <div className="flex flex-col gap-4">
    <div className="flex items-center gap-2">
      <Checkbox id="cb-unchecked" />
      <Label htmlFor="cb-unchecked">Unchecked &mdash; spot instances</Label>
    </div>
    <div className="flex items-center gap-2">
      <Checkbox id="cb-checked" defaultChecked />
      <Label htmlFor="cb-checked">Checked &mdash; managed DNS</Label>
    </div>
    <div className="flex items-center gap-2">
      <Checkbox id="cb-indeterminate" checked="indeterminate" />
      <Label htmlFor="cb-indeterminate">Indeterminate &mdash; some namespaces</Label>
    </div>
  </div>
)

export const Disabled = () => (
  <div className="flex flex-col gap-4">
    <div className="flex items-center gap-2">
      <Checkbox id="cb-disabled" disabled />
      <Label htmlFor="cb-disabled" className="text-muted-foreground">
        GPU node pool (not available in ap-south-1)
      </Label>
    </div>
    <div className="flex items-center gap-2">
      <Checkbox id="cb-disabled-checked" disabled defaultChecked />
      <Label htmlFor="cb-disabled-checked" className="text-muted-foreground">
        Wireguard tunnel (always on)
      </Label>
    </div>
  </div>
)

export const WithDescription = () => (
  <div className="flex items-start gap-2 w-80">
    <Checkbox id="cb-telemetry" defaultChecked />
    <div className="space-y-2">
      <Label htmlFor="cb-telemetry">Ship cluster telemetry</Label>
      <p className="text-sm text-muted-foreground">
        Sends node and pod metrics to the Kloudlite control plane so the console can chart
        CPU and memory usage.
      </p>
    </div>
  </div>
)

export const InAList = () => (
  <div className="flex flex-col gap-4 w-80">
    <div className="text-sm font-medium">Namespaces to intercept</div>
    {['kl-platform', 'kl-core', 'default'].map((ns, i) => (
      <div key={ns} className="flex items-center gap-2">
        <Checkbox id={`ns-${ns}`} defaultChecked={i < 2} />
        <Label htmlFor={`ns-${ns}`} className="font-mono">
          {ns}
        </Label>
      </div>
    ))}
  </div>
)
