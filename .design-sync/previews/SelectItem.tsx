import * as React from 'react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@kloudlite/ui'

export const Items = () => (
  <div className="flex h-64 w-80 flex-col">
    <Select open defaultValue="eu-west-1">
      <SelectTrigger>
        <SelectValue placeholder="Select a region" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="ap-south-1">ap-south-1 &middot; Mumbai</SelectItem>
        <SelectItem value="eu-west-1">eu-west-1 &middot; Ireland</SelectItem>
        <SelectItem value="us-east-1">us-east-1 &middot; N. Virginia</SelectItem>
      </SelectContent>
    </Select>
  </div>
)

export const SelectedIndicator = () => (
  <div className="flex h-64 w-80 flex-col">
    <Select open defaultValue="staging">
      <SelectTrigger>
        <SelectValue placeholder="Select an environment" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="dev">development</SelectItem>
        <SelectItem value="staging">staging</SelectItem>
        <SelectItem value="prod">production</SelectItem>
      </SelectContent>
    </Select>
  </div>
)

export const DisabledItem = () => (
  <div className="flex h-64 w-80 flex-col">
    <Select open defaultValue="ondemand-4x">
      <SelectTrigger>
        <SelectValue placeholder="Select a node pool" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="ondemand-4x">ondemand-4x &middot; 4 vCPU</SelectItem>
        <SelectItem value="spot-2x">spot-2x &middot; 2 vCPU</SelectItem>
        <SelectItem value="gpu-a10" disabled>
          gpu-a10 &middot; quota exceeded
        </SelectItem>
      </SelectContent>
    </Select>
  </div>
)
