import * as React from 'react'
import {
  Label,
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectSeparator,
  SelectTrigger,
  SelectValue,
} from '@kloudlite/ui'

export const Basic = () => (
  <Select defaultValue="ap-south-1">
    <SelectTrigger className="w-64">
      <SelectValue placeholder="Select a region" />
    </SelectTrigger>
    <SelectContent>
      <SelectItem value="ap-south-1">ap-south-1 &middot; Mumbai</SelectItem>
      <SelectItem value="eu-west-1">eu-west-1 &middot; Ireland</SelectItem>
      <SelectItem value="us-east-1">us-east-1 &middot; N. Virginia</SelectItem>
    </SelectContent>
  </Select>
)

export const WithLabel = () => (
  <div className="w-64 space-y-2">
    <Label htmlFor="cluster-size">Cluster size</Label>
    <Select defaultValue="medium">
      <SelectTrigger id="cluster-size">
        <SelectValue placeholder="Select a size" />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          <SelectLabel>Shared</SelectLabel>
          <SelectItem value="small">Small &middot; 2 vCPU</SelectItem>
          <SelectItem value="medium">Medium &middot; 4 vCPU</SelectItem>
        </SelectGroup>
        <SelectSeparator />
        <SelectGroup>
          <SelectLabel>Dedicated</SelectLabel>
          <SelectItem value="large">Large &middot; 8 vCPU</SelectItem>
        </SelectGroup>
      </SelectContent>
    </Select>
  </div>
)

export const Placeholder = () => (
  <Select>
    <SelectTrigger className="w-64">
      <SelectValue placeholder="Select an environment" />
    </SelectTrigger>
    <SelectContent>
      <SelectItem value="dev">development</SelectItem>
      <SelectItem value="staging">staging</SelectItem>
    </SelectContent>
  </Select>
)

export const Disabled = () => (
  <Select disabled defaultValue="locked">
    <SelectTrigger className="w-64">
      <SelectValue />
    </SelectTrigger>
    <SelectContent>
      <SelectItem value="locked">Managed by your admin</SelectItem>
    </SelectContent>
  </Select>
)
