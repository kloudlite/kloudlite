import * as React from 'react'
import {
  Label,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@kloudlite/ui'

export const Closed = () => (
  <div className="w-80">
    <Select defaultValue="ap-south-1">
      <SelectTrigger>
        <SelectValue placeholder="Select a region" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="ap-south-1">ap-south-1 &middot; Mumbai</SelectItem>
        <SelectItem value="eu-west-1">eu-west-1 &middot; Ireland</SelectItem>
      </SelectContent>
    </Select>
  </div>
)

export const WithPlaceholder = () => (
  <div className="w-80 space-y-2">
    <Label htmlFor="np-trigger">Node pool</Label>
    <Select>
      <SelectTrigger id="np-trigger">
        <SelectValue placeholder="Select a node pool" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="spot-2x">spot-2x &middot; 2 vCPU</SelectItem>
        <SelectItem value="ondemand-4x">ondemand-4x &middot; 4 vCPU</SelectItem>
      </SelectContent>
    </Select>
  </div>
)

export const Disabled = () => (
  <div className="w-80">
    <Select disabled defaultValue="prod">
      <SelectTrigger>
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="prod">production &middot; locked by admin</SelectItem>
      </SelectContent>
    </Select>
  </div>
)

export const Open = () => (
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
