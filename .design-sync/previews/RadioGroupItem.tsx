import * as React from 'react'
import { Label, RadioGroup, RadioGroupItem } from '@kloudlite/ui'

export const SelectedAndUnselected = () => (
  <RadioGroup defaultValue="ap-south-1" className="w-80">
    <div className="flex items-center gap-3">
      <RadioGroupItem value="ap-south-1" id="region-mum" />
      <Label htmlFor="region-mum">ap-south-1 &middot; Mumbai</Label>
    </div>
    <div className="flex items-center gap-3">
      <RadioGroupItem value="eu-west-1" id="region-dub" />
      <Label htmlFor="region-dub">eu-west-1 &middot; Ireland</Label>
    </div>
  </RadioGroup>
)

export const DisabledItems = () => (
  <RadioGroup defaultValue="shared" className="w-80">
    <div className="flex items-center gap-3">
      <RadioGroupItem value="shared" id="tier-shared" />
      <Label htmlFor="tier-shared">Shared runners</Label>
    </div>
    <div className="flex items-center gap-3">
      <RadioGroupItem value="dedicated" id="tier-dedicated" disabled />
      <Label htmlFor="tier-dedicated">Dedicated runners (upgrade required)</Label>
    </div>
  </RadioGroup>
)

export const CardRows = () => (
  <RadioGroup defaultValue="byoc" className="w-96">
    <Label
      htmlFor="host-cloud"
      className="flex items-center gap-3 border p-3 font-normal"
    >
      <RadioGroupItem value="cloud" id="host-cloud" />
      <span className="text-sm">Kloudlite Cloud &mdash; we run the cluster</span>
    </Label>
    <Label
      htmlFor="host-byoc"
      className="flex items-center gap-3 border p-3 font-normal"
    >
      <RadioGroupItem value="byoc" id="host-byoc" />
      <span className="text-sm">Bring your own cluster</span>
    </Label>
  </RadioGroup>
)
