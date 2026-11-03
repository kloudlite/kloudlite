import * as React from 'react'
import { Label, Slider } from '@kloudlite/ui'

export const Basic = () => (
  <div className="w-80">
    <Slider defaultValue={[3]} min={1} max={10} step={1} />
  </div>
)

export const ReplicaCount = () => (
  <div className="w-80 space-y-3">
    <div className="flex items-center justify-between">
      <Label>Replicas</Label>
      <span className="text-sm tabular-nums text-muted-foreground">4</span>
    </div>
    <Slider defaultValue={[4]} min={1} max={12} step={1} />
  </div>
)

export const CpuRange = () => (
  <div className="w-80 space-y-3">
    <Label>CPU limit range (vCPU)</Label>
    <Slider defaultValue={[2, 6]} min={0} max={8} step={1} />
    <p className="text-xs text-muted-foreground">Requests 2 vCPU, limits 6 vCPU</p>
  </div>
)

export const Disabled = () => (
  <div className="w-80 space-y-3">
    <Label>Node pool size (managed by plan)</Label>
    <Slider defaultValue={[8]} min={0} max={16} step={1} disabled />
  </div>
)
